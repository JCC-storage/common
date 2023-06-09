package distlock

import (
	"fmt"

	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	LOCK_REQUEST_DATA_PREFIX = "/distlock/lockRequest/data"
	LOCK_REQUEST_INDEX       = "/distlock/lockRequest/index"
	LOCK_REQUEST_LOCK_NAME   = "/distlock/lockRequest/lock"
)

type Service struct {
	cfg     *Config
	etcdCli *clientv3.Client

	main      *mainActor
	providers *providersActor
	watchEtcd *watchEtcdActor
}

func NewService(cfg *Config) (*Service, error) {
	etcdCli, err := clientv3.New(clientv3.Config{
		Endpoints: []string{cfg.EtcdAddress},
		Username:  cfg.EtcdUsername,
		Password:  cfg.EtcdPassword,
	})

	if err != nil {
		return nil, fmt.Errorf("new etcd client failed, err: %w", err)
	}

	mainActor := newMainActor()
	providersActor := newProvidersActor()
	watchEtcdActor := newWatchEtcdActor()

	mainActor.Init(watchEtcdActor, providersActor)
	providersActor.Init()
	watchEtcdActor.Init(providersActor)

	return &Service{
		cfg:       cfg,
		etcdCli:   etcdCli,
		main:      mainActor,
		providers: providersActor,
		watchEtcd: watchEtcdActor,
	}, nil
}

// Acquire 请求一批锁。成功后返回锁请求ID
func (svc *Service) Acquire(req LockRequest) (reqID string, err error) {
	return svc.main.Acquire(req)
}

// Renew 续约锁
func (svc *Service) Renew(reqID string) error {
	panic("todo")

}

// Release 释放锁
func (svc *Service) Release(reqID string) error {
	return svc.main.Release(reqID)
}

func (svc *Service) Serve() error {
	go func() {
		// TODO 处理错误
		svc.providers.Serve()
	}()

	go func() {
		// TODO 处理错误
		svc.watchEtcd.Serve()
	}()

	// 考虑更好的错误处理方式
	return svc.main.Serve()
}
