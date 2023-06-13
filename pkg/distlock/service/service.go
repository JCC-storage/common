package service

import (
	"fmt"
	"time"

	"gitlink.org.cn/cloudream/common/pkg/distlock"
	"gitlink.org.cn/cloudream/common/pkg/distlock/service/internal"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type Service struct {
	cfg     *distlock.Config
	etcdCli *clientv3.Client

	mainActor      *internal.MainActor
	providersActor *internal.ProvidersActor
	watchEtcdActor *internal.WatchEtcdActor
	leaseActor     *internal.LeaseActor
}

func NewService(cfg *distlock.Config) (*Service, error) {
	etcdCli, err := clientv3.New(clientv3.Config{
		Endpoints: []string{cfg.EtcdAddress},
		Username:  cfg.EtcdUsername,
		Password:  cfg.EtcdPassword,
	})

	if err != nil {
		return nil, fmt.Errorf("new etcd client failed, err: %w", err)
	}

	mainActor := internal.NewMainActor()
	providersActor := internal.NewProvidersActor()
	watchEtcdActor := internal.NewWatchEtcdActor()
	leaseActor := internal.NewLeaseActor()

	mainActor.Init(watchEtcdActor, providersActor)
	providersActor.Init()
	watchEtcdActor.Init(providersActor)
	leaseActor.Init(mainActor)

	return &Service{
		cfg:            cfg,
		etcdCli:        etcdCli,
		mainActor:      mainActor,
		providersActor: providersActor,
		watchEtcdActor: watchEtcdActor,
		leaseActor:     leaseActor,
	}, nil
}

// Acquire 请求一批锁。成功后返回锁请求ID
func (svc *Service) Acquire(req distlock.LockRequest) (string, error) {
	reqID, err := svc.mainActor.Acquire(req)
	if err != nil {
		return "", err
	}

	// TODO 不影响结果，但考虑打日志
	svc.leaseActor.Add(reqID, time.Duration(svc.cfg.LockRequestLeaseTimeSec)*time.Second)

	return reqID, nil
}

// Renew 续约锁
func (svc *Service) Renew(reqID string) error {
	return svc.leaseActor.Renew(reqID, time.Duration(svc.cfg.LockRequestLeaseTimeSec)*time.Second)
}

// Release 释放锁
func (svc *Service) Release(reqID string) error {
	err := svc.mainActor.Release(reqID)

	// TODO 不影响结果，但考虑打日志
	svc.leaseActor.Remove(reqID)

	return err
}

func (svc *Service) Serve() error {
	go func() {
		// TODO 处理错误
		svc.providersActor.Serve()
	}()

	go func() {
		// TODO 处理错误
		svc.watchEtcdActor.Serve()
	}()

	go func() {
		// TODO 处理错误
		svc.mainActor.Serve()
	}()

	go func() {
		// TODO 处理错误
		svc.leaseActor.Server()
	}()

	err := svc.mainActor.ReloadEtcdData()
	if err != nil {
		// TODO 关闭其他的Actor，或者更好的错误处理方式
		return fmt.Errorf("init data failed, err: %w", err)
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

	// TODO 临时解决办法
	ch := make(chan any)
	<-ch

	return nil
}
