package service

import (
	"fmt"
	"time"

	"gitlink.org.cn/cloudream/common/pkg/distlock"
	"gitlink.org.cn/cloudream/common/pkg/distlock/service/internal"
	"gitlink.org.cn/cloudream/common/pkg/logger"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type AcquireOption struct {
	RetryTimeMs  int // 如果第一次获取锁失败，则在这个时间内进行重试。为0不进行重试。
	LeaseTimeSec int // 锁的租约时间。为0不设置租约。
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

func NewService(cfg *distlock.Config) (*Service, error) {
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

	initProviders(providersActor)

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
func (svc *Service) Acquire(req distlock.LockRequest, opts ...AcquireOption) (string, error) {
	var opt AcquireOption
	if len(opts) > 0 {
		opt = opts[0]
	}

	reqID, err := svc.mainActor.Acquire(req)
	if err != nil {
		if opt.RetryTimeMs <= 0 {
			return "", err
		}

		fut, err := svc.retryActor.Retry(req, time.Duration(opt.RetryTimeMs)*time.Millisecond, err)
		if err != nil {
			return "", fmt.Errorf("retrying failed, err: %w", err)
		}

		reqID, err = fut.WaitValue()
		if err != nil {
			return "", err
		}
	}

	if opt.LeaseTimeSec > 0 {
		// TODO 不影响结果，但考虑打日志
		err := svc.leaseActor.Add(reqID, time.Duration(opt.LeaseTimeSec)*time.Second)
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
	err := svc.mainActor.Release(reqID)

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
