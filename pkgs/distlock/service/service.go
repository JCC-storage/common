package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"gitlink.org.cn/cloudream/common/pkgs/distlock"
	"gitlink.org.cn/cloudream/common/pkgs/distlock/service/internal"
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
	Provider distlock.LockProvider
}

func NewPathProvider(prov distlock.LockProvider, path ...any) PathProvider {
	return PathProvider{
		Path:     path,
		Provider: prov,
	}
}

type Service struct {
	cfg     *distlock.Config
	etcdCli *clientv3.Client

	acquireActor   *internal.AcquireActor
	releaseActor   *internal.ReleaseActor
	providersActor *internal.ProvidersActor
	watchEtcdActor *internal.WatchEtcdActor
	leaseActor     *internal.LeaseActor

	lockReqEventWatcher internal.LockRequestEventWatcher
}

func NewService(cfg *distlock.Config, initProvs []PathProvider) (*Service, error) {
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

	acquireActor.Init(providersActor)
	providersActor.Init()
	watchEtcdActor.Init()
	leaseActor.Init(releaseActor)

	for _, prov := range initProvs {
		providersActor.AddProvider(prov.Provider, prov.Path...)
	}

	return &Service{
		cfg:            cfg,
		etcdCli:        etcdCli,
		acquireActor:   acquireActor,
		releaseActor:   releaseActor,
		providersActor: providersActor,
		watchEtcdActor: watchEtcdActor,
		leaseActor:     leaseActor,
	}, nil
}

// Acquire 请求一批锁。成功后返回锁请求ID
func (svc *Service) Acquire(req distlock.LockRequest, opts ...AcquireOptionFn) (string, error) {
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
	// 1. client退出时直接中断进程，此时RetryActor可能正在进行Retry，于是导致Etcd锁没有解除就退出了进程。
	// 虽然由于租约的存在不会导致系统长期卡死，但会影响client的使用

	go func() {
		// TODO 处理错误
		err := svc.providersActor.Serve()
		if err != nil {
			logger.Std.Warnf("serving providers actor failed, err: %s", err.Error())
		}
	}()

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

	revision, err := svc.loadState()
	if err != nil {
		// TODO 关闭其他的Actor，或者更好的错误处理方式
		return fmt.Errorf("init data failed, err: %w", err)
	}

	svc.lockReqEventWatcher.OnEvent = func(events []internal.LockRequestEvent) {
		svc.acquireActor.TryAcquireNow()
		svc.providersActor.ApplyLockRequestEvents(events)
	}
	err = svc.watchEtcdActor.AddEventWatcher(&svc.lockReqEventWatcher)
	if err != nil {
		// TODO 关闭其他的Actor，或者更好的错误处理方式
		return fmt.Errorf("add lock request event watcher failed, err: %w", err)
	}

	err = svc.watchEtcdActor.StartWatching(revision)
	if err != nil {
		// TODO 关闭其他的Actor，或者更好的错误处理方式
		return fmt.Errorf("start watching etcd failed, err: %w", err)
	}

	err = svc.leaseActor.StartChecking()
	if err != nil {
		// TODO 关闭其他的Actor，或者更好的错误处理方式
		return fmt.Errorf("start checking lease failed, err: %w", err)
	}

	// TODO 防止退出的临时解决办法
	ch := make(chan any)
	<-ch

	return nil
}

func (svc *Service) loadState() (int64, error) {
	// 使用事务一次性获取index和锁数据，就不需要加全局锁了
	txResp, err := svc.etcdCli.Txn(context.Background()).
		Then(
			clientv3.OpGet(internal.EtcdLockRequestIndex),
			clientv3.OpGet(internal.EtcdLockRequestData, clientv3.WithPrefix()),
		).
		Commit()
	if err != nil {
		return 0, fmt.Errorf("get etcd data failed, err: %w", err)
	}

	indexKvs := txResp.Responses[0].GetResponseRange().Kvs
	lockKvs := txResp.Responses[1].GetResponseRange().Kvs

	var index int64
	var reqData []internal.LockRequestData

	// 解析Index
	if len(indexKvs) > 0 {
		val, err := strconv.ParseInt(string(indexKvs[0].Value), 0, 64)
		if err != nil {
			return 0, fmt.Errorf("parse lock request index failed, err: %w", err)
		}
		index = val

	} else {
		index = 0
	}

	// 解析锁请求数据
	for _, kv := range lockKvs {
		var req internal.LockRequestData
		err := serder.JSONToObject(kv.Value, &req)
		if err != nil {
			return 0, fmt.Errorf("parse lock request data failed, err: %w", err)
		}

		reqData = append(reqData, req)
	}

	err = svc.providersActor.ResetState(index, reqData)
	if err != nil {
		return 0, fmt.Errorf("reset lock providers state failed, err: %w", err)
	}

	return txResp.Header.Revision, nil
}
