package distlock

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"gitlink.org.cn/cloudream/common/pkgs/actor"
	"gitlink.org.cn/cloudream/common/pkgs/distlock/internal"
	"gitlink.org.cn/cloudream/common/pkgs/logger"
	"gitlink.org.cn/cloudream/common/utils/serder"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type AcquireOption struct {
	Timeout time.Duration
	Lease   time.Duration
}

type AcquireOptionFn func(opt *AcquireOption)

func WithTimeout(timeout time.Duration) AcquireOptionFn {
	return func(opt *AcquireOption) {
		opt.Timeout = timeout
	}
}

func WithLease(time time.Duration) AcquireOptionFn {
	return func(opt *AcquireOption) {
		opt.Lease = time
	}
}

type PathProvider struct {
	Path     []any
	Provider internal.LockProvider
}

func NewPathProvider(prov internal.LockProvider, path ...any) PathProvider {
	return PathProvider{
		Path:     path,
		Provider: prov,
	}
}

type Service struct {
	cfg     *internal.Config
	etcdCli *clientv3.Client

	cmdChan          *actor.CommandChannel
	acquireActor     *internal.AcquireActor
	releaseActor     *internal.ReleaseActor
	providersActor   *internal.ProvidersActor
	watchEtcdActor   *internal.WatchEtcdActor
	leaseActor       *internal.LeaseActor
	serviceInfoActor *internal.ServiceInfoActor
}

func NewService(cfg *internal.Config, initProvs []PathProvider) (*Service, error) {
	etcdCli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{cfg.EtcdAddress},
		Username:    cfg.EtcdUsername,
		Password:    cfg.EtcdPassword,
		DialTimeout: time.Second * 5,
	})
	if err != nil {
		return nil, fmt.Errorf("new etcd client failed, err: %w", err)
	}

	svc := &Service{
		cfg:     cfg,
		etcdCli: etcdCli,
		cmdChan: actor.NewCommandChannel(),
	}

	svc.acquireActor = internal.NewAcquireActor(cfg, etcdCli)
	svc.releaseActor = internal.NewReleaseActor(cfg, etcdCli)
	svc.providersActor = internal.NewProvidersActor()
	svc.watchEtcdActor = internal.NewWatchEtcdActor(etcdCli)
	svc.leaseActor = internal.NewLeaseActor()
	svc.serviceInfoActor = internal.NewServiceInfoActor(cfg, etcdCli, internal.ServiceInfo{
		Description: cfg.ServiceDescription,
	})

	svc.acquireActor.Init(svc.providersActor)
	svc.leaseActor.Init(svc.releaseActor)
	svc.providersActor.Init()
	svc.watchEtcdActor.Init(
		func(event internal.LockRequestEvent) {
			err := svc.providersActor.OnLockRequestEvent(event)
			if err != nil {
				logger.Std.Warnf("%s, will reset service state", err.Error())
				svc.cmdChan.Send(func() { svc.doResetState() })
				return
			}

			svc.acquireActor.TryAcquireNow()
			svc.releaseActor.OnLockRequestEvent(event)
			svc.serviceInfoActor.OnLockRequestEvent(event)
		},
		func(event internal.ServiceEvent) {
			err := svc.serviceInfoActor.OnServiceEvent(event)
			if err != nil {
				logger.Std.Warnf("%s, will reset service state", err.Error())
				svc.cmdChan.Send(func() { svc.doResetState() })
			}
		},
		func(err error) {
			logger.Std.Warnf("%s, will reset service state", err.Error())
			svc.cmdChan.Send(func() { svc.doResetState() })
		},
	)

	for _, prov := range initProvs {
		svc.providersActor.AddProvider(prov.Provider, prov.Path...)
	}

	return svc, nil
}

// Acquire 请求一批锁。成功后返回锁请求ID
func (svc *Service) Acquire(req internal.LockRequest, opts ...AcquireOptionFn) (string, error) {
	var opt = AcquireOption{
		Timeout: time.Second * 10,
	}
	for _, fn := range opts {
		fn(&opt)
	}

	ctx := context.Background()
	if opt.Timeout != 0 {
		var cancel func()
		ctx, cancel = context.WithTimeout(ctx, opt.Timeout)
		defer cancel()
	}

	reqID, err := svc.acquireActor.Acquire(ctx, req)
	if err != nil {
		return "", err
	}

	if opt.Lease > 0 {
		// TODO 不影响结果，但考虑打日志
		err := svc.leaseActor.Add(reqID, opt.Lease)
		if err != nil {
			logger.Std.Warnf("adding lease: %s", err.Error())
		}
	}

	return reqID, nil
}

// Renew 续约锁。只有在加锁时设置了续约时间才有意义
func (svc *Service) Renew(reqID string) error {
	return svc.leaseActor.Renew(reqID)
}

// Release 释放锁
func (svc *Service) Release(reqID string) {
	svc.releaseActor.Release([]string{reqID})

	// TODO 不影响结果，但考虑打日志
	err := svc.leaseActor.Remove(reqID)
	if err != nil {
		logger.Std.Warnf("removing lease: %s", err.Error())
	}
}

func (svc *Service) Serve() error {
	// TODO 需要停止service的方法
	// 目前已知问题：
	// 1. client退出时直接中断进程，此时AcquireActor可能正在进行重试，于是导致Etcd锁没有解除就退出了进程。
	// 虽然由于租约的存在不会导致系统长期卡死，但会影响client的使用

	go svc.watchEtcdActor.Serve()

	go svc.leaseActor.Serve()

	svc.cmdChan.Send(func() { svc.doResetState() })

	cmdChan := svc.cmdChan.BeginChanReceive()
	defer svc.cmdChan.CloseChanReceive()

	for {
		select {
		case cmd := <-cmdChan:
			cmd()
		}
	}

	return nil
}

func (svc *Service) doResetState() {
	logger.Std.Infof("start reset state")
	// TODO context
	err := svc.resetState(context.Background())
	if err != nil {
		logger.Std.Warnf("reseting state: %s, will try again after 3 seconds", err.Error())
		<-time.After(time.Second * 3)
		svc.cmdChan.Send(func() { svc.doResetState() })
		return
	}
	logger.Std.Infof("reset state success")
}

// ResetState 重置内部状态。注：只要调用到了此函数，无论在哪一步出的错，
// 都要将内部状态视为已被破坏，直到成功调用了此函数才能继续后面的步骤。
// 如果调用失败，服务将进入维护模式，届时可以接受请求，但不会处理请求，直到调用成功为止。
func (svc *Service) resetState(ctx context.Context) error {
	// 让服务都进入维护模式
	svc.watchEtcdActor.Stop()
	svc.leaseActor.Stop()
	svc.acquireActor.EnterMaintenance()
	svc.releaseActor.EnterMaintenance()

	// 必须使用事务一次性获取所有数据
	txResp, err := svc.etcdCli.Txn(ctx).
		Then(
			clientv3.OpGet(internal.EtcdLockRequestIndex),
			clientv3.OpGet(internal.EtcdLockRequestDataPrefix, clientv3.WithPrefix()),
			clientv3.OpGet(internal.EtcdServiceInfoPrefix, clientv3.WithPrefix()),
		).
		Commit()
	if err != nil {
		return fmt.Errorf("getting etcd data: %w", err)
	}

	// 解析Index
	var index int64 = 0
	indexKvs := txResp.Responses[0].GetResponseRange().Kvs
	if len(indexKvs) > 0 {
		val, err := strconv.ParseInt(string(indexKvs[0].Value), 0, 64)
		if err != nil {
			return fmt.Errorf("parsing lock request index: %w", err)
		}
		index = val
	}

	// 解析锁请求数据
	var reqData []internal.LockRequestData
	lockKvs := txResp.Responses[1].GetResponseRange().Kvs
	for _, kv := range lockKvs {
		var req internal.LockRequestData
		err := serder.JSONToObject(kv.Value, &req)
		if err != nil {
			return fmt.Errorf("parsing lock request data: %w", err)
		}

		reqData = append(reqData, req)
	}

	// 解析服务信息数据
	var svcInfo []internal.ServiceInfo
	svcInfoKvs := txResp.Responses[2].GetResponseRange().Kvs
	for _, kv := range svcInfoKvs {
		var info internal.ServiceInfo
		err := serder.JSONToObject(kv.Value, &info)
		if err != nil {
			return fmt.Errorf("parsing service info data: %w", err)
		}

		svcInfo = append(svcInfo, info)
	}

	// 然后将新获取到的状态装填到Actor中
	releasingIDs, err := svc.serviceInfoActor.ResetState(ctx, svcInfo, reqData)
	if err != nil {
		return fmt.Errorf("reseting service info actor: %w", err)
	}

	// 要在acquireActor之前，因为acquireActor会调用它的WaitLocalIndexTo
	err = svc.providersActor.ResetState(index, reqData)
	if err != nil {
		return fmt.Errorf("reseting providers actor: %w", err)
	}

	svc.acquireActor.ResetState(svc.serviceInfoActor.GetSelfInfo().ID)

	// ReleaseActor没有什么需要Reset的状态
	svc.releaseActor.Release(releasingIDs)

	// 重置完了之后再退出维护模式
	svc.watchEtcdActor.Start(txResp.Header.Revision)
	svc.leaseActor.Start()
	svc.acquireActor.LeaveMaintenance()
	svc.releaseActor.LeaveMaintenance()

	svc.acquireActor.TryAcquireNow()
	svc.releaseActor.TryReleaseNow()

	return nil
}
