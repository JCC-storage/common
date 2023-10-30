package distlock

import (
	"context"
	"fmt"
	"strconv"
	"time"

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

	acquireActor := internal.NewAcquireActor(cfg, etcdCli)
	releaseActor := internal.NewReleaseActor(cfg, etcdCli)
	providersActor := internal.NewProvidersActor()
	watchEtcdActor := internal.NewWatchEtcdActor(etcdCli)
	leaseActor := internal.NewLeaseActor()
	serviceInfoActor := internal.NewServiceInfoActor(cfg, etcdCli)

	acquireActor.Init(providersActor)
	leaseActor.Init(releaseActor)
	providersActor.Init()
	watchEtcdActor.Init(
		func(event internal.LockRequestEvent) {
			providersActor.OnLockRequestEvent(event)
			acquireActor.TryAcquireNow()
			releaseActor.OnLockRequestEvent(event)
			serviceInfoActor.OnLockRequestEvent(event)
		},
		func(event internal.ServiceEvent) {
			serviceInfoActor.OnServiceEvent(event)
		},
	)

	for _, prov := range initProvs {
		providersActor.AddProvider(prov.Provider, prov.Path...)
	}

	return &Service{
		cfg:              cfg,
		etcdCli:          etcdCli,
		acquireActor:     acquireActor,
		releaseActor:     releaseActor,
		providersActor:   providersActor,
		watchEtcdActor:   watchEtcdActor,
		leaseActor:       leaseActor,
		serviceInfoActor: serviceInfoActor,
	}, nil
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

	go func() {
		// TODO 处理错误
		err := svc.watchEtcdActor.Serve()
		if err != nil {
			logger.Std.Warnf("serving watch etcd actor actor failed, err: %s", err.Error())
		}
	}()

	go func() {
		// TODO 处理错误
		err := svc.leaseActor.Serve()
		if err != nil {
			logger.Std.Warnf("serving lease actor failed, err: %s", err.Error())
		}
	}()

	// TODO context
	err := svc.resetState(context.Background())
	if err != nil {
		// TODO 关闭其他的Actor，或者更好的错误处理方式
		return fmt.Errorf("init data failed, err: %w", err)
	}

	// TODO 防止退出的临时解决办法
	ch := make(chan any)
	<-ch

	return nil
}

// ResetState 重置内部状态。注：只要调用到了此函数，无论在哪一步出的错，
// 都要将内部状态视为已被破坏，直到成功调用了此函数才能继续后面的步骤
func (svc *Service) resetState(ctx context.Context) error {
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

	// 先停止监听等定时事件
	svc.watchEtcdActor.Stop()
	svc.leaseActor.Stop()

	// 然后将新获取到的状态装填到Actor中。注：执行顺序需要考虑Actor会被谁调用，不会被调用的优先Reset。
	releasingIDs, err := svc.serviceInfoActor.ResetState(ctx, svcInfo, reqData)
	if err != nil {
		return fmt.Errorf("reseting service info actor: %w", err)
	}

	svc.acquireActor.ResetState(svc.serviceInfoActor.GetSelfInfo().ID)

	svc.leaseActor.ResetState()

	err = svc.providersActor.ResetState(index, reqData)
	if err != nil {
		return fmt.Errorf("reseting providers actor: %w", err)
	}

	svc.releaseActor.ResetState(releasingIDs)

	// 重置完了之后再启动监听
	svc.watchEtcdActor.Start(txResp.Header.Revision)
	svc.leaseActor.Start()
	return nil
}
