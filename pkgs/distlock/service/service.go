package service

import (
	"context"
	"fmt"
	"time"

	"gitlink.org.cn/cloudream/common/pkgs/distlock"
	"gitlink.org.cn/cloudream/common/pkgs/distlock/service/internal"
	"gitlink.org.cn/cloudream/common/pkgs/logger"
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

	mainActor      *internal.MainActor
	providersActor *internal.ProvidersActor
	watchEtcdActor *internal.WatchEtcdActor
	leaseActor     *internal.LeaseActor
	retryActor     *internal.RetryActor

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

	mainActor := internal.NewMainActor(cfg, etcdCli)
	providersActor := internal.NewProvidersActor()
	watchEtcdActor := internal.NewWatchEtcdActor(etcdCli)
	leaseActor := internal.NewLeaseActor()
	retryActor := internal.NewRetryActor()

	mainActor.Init(providersActor)
	providersActor.Init()
	watchEtcdActor.Init()
	leaseActor.Init(mainActor)
	retryActor.Init(mainActor)

	for _, prov := range initProvs {
		providersActor.AddProvider(prov.Provider, prov.Path...)
	}

	return &Service{
		cfg:            cfg,
		etcdCli:        etcdCli,
		mainActor:      mainActor,
		providersActor: providersActor,
		watchEtcdActor: watchEtcdActor,
		leaseActor:     leaseActor,
		retryActor:     retryActor,
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

	reqID, err := svc.mainActor.Acquire(ctx, req)
	if err != nil {
		fut, err := svc.retryActor.Retry(ctx, req, err)
		if err != nil {
			return "", fmt.Errorf("retrying failed, err: %w", err)
		}

		// Retry 如果超时，Retry内部会设置fut为Failed，所以这里可以用Background无限等待
		reqID, err = fut.WaitValue(context.Background())
		if err != nil {
			return "", err
		}
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
func (svc *Service) Release(reqID string) error {
	err := svc.mainActor.Release(context.TODO(), reqID)

	// TODO 不影响结果，但考虑打日志
	err2 := svc.leaseActor.Remove(reqID)
	if err2 != nil {
		logger.Std.Warnf("removing lease: %s", err2.Error())
	}

	return err
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
		err := svc.mainActor.Serve()
		if err != nil {
			logger.Std.Warnf("serving main actor failed, err: %s", err.Error())
		}
	}()

	go func() {
		// TODO 处理错误
		err := svc.leaseActor.Serve()
		if err != nil {
			logger.Std.Warnf("serving lease actor failed, err: %s", err.Error())
		}
	}()

	go func() {
		// TODO 处理错误
		err := svc.retryActor.Serve()
		if err != nil {
			logger.Std.Warnf("serving retry actor failed, err: %s", err.Error())
		}
	}()

	err := svc.mainActor.ReloadEtcdData()
	if err != nil {
		// TODO 关闭其他的Actor，或者更好的错误处理方式
		return fmt.Errorf("init data failed, err: %w", err)
	}

	svc.lockReqEventWatcher.OnEvent = func(events []internal.LockRequestEvent) {
		svc.providersActor.ApplyLockRequestEvents(events)
		svc.retryActor.OnLocalStateUpdated()
	}
	err = svc.watchEtcdActor.AddEventWatcher(&svc.lockReqEventWatcher)
	if err != nil {
		// TODO 关闭其他的Actor，或者更好的错误处理方式
		return fmt.Errorf("add lock request event watcher failed, err: %w", err)
	}

	err = svc.watchEtcdActor.StartWatching()
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
