package internal

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"gitlink.org.cn/cloudream/common/pkgs/logger"
	"gitlink.org.cn/cloudream/common/utils/lo2"
	"gitlink.org.cn/cloudream/common/utils/serder"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var ErrSelfServiceDown = errors.New("self service is down, need to restart")

type serviceStatus struct {
	Info           ServiceInfo
	LockRequestIDs []string
}

type ServiceInfoActor struct {
	cfg     *Config
	etcdCli *clientv3.Client

	lock           sync.Mutex
	selfInfo       ServiceInfo
	leaseID        *clientv3.LeaseID
	leaseKeepAlive chan any
	services       map[string]*serviceStatus
	releaseActor   *ReleaseActor
}

func NewServiceInfoActor(cfg *Config, etcdCli *clientv3.Client, baseSelfInfo ServiceInfo) *ServiceInfoActor {
	return &ServiceInfoActor{
		cfg:      cfg,
		etcdCli:  etcdCli,
		selfInfo: baseSelfInfo,
	}
}

func (a *ServiceInfoActor) Init(releaseActor *ReleaseActor) {
	a.releaseActor = releaseActor
}

func (a *ServiceInfoActor) GetSelfInfo() *ServiceInfo {
	return &a.selfInfo
}

func (a *ServiceInfoActor) ResetState(ctx context.Context, currentServices []ServiceInfo, currentLocks []LockRequestData) ([]string, error) {
	a.lock.Lock()
	defer a.lock.Unlock()

	if a.leaseID != nil {
		a.etcdCli.Revoke(ctx, *a.leaseID)
		close(a.leaseKeepAlive)
		a.leaseID = nil
	}

	// 生成并注册服务信息
	a.selfInfo.ID = uuid.NewString()

	infoData, err := serder.ObjectToJSON(a.selfInfo)
	if err != nil {
		return nil, fmt.Errorf("service info to json: %w", err)
	}

	lease, err := a.etcdCli.Grant(ctx, a.cfg.EtcdLockLeaseTimeSec)
	if err != nil {
		return nil, fmt.Errorf("granting lease: %w", err)
	}
	a.leaseID = &lease.ID

	keepAliveChan, err := a.etcdCli.Lease.KeepAlive(context.Background(), lease.ID)
	if err != nil {
		a.etcdCli.Revoke(ctx, lease.ID)
		return nil, fmt.Errorf("starting keep lease alive: %w", err)
	}
	a.leaseKeepAlive = make(chan any)

	go func() {
		for {
			select {
			case _, ok := <-keepAliveChan:
				if !ok {
					logger.Std.Warnf("lease keep alive channel closed, will try to open again")

					var err error
					keepAliveChan, err = a.etcdCli.Lease.KeepAlive(context.Background(), lease.ID)
					if err != nil {
						logger.Std.Warnf("starting keep lease alive: %s", err.Error())
						return
					}
				}

			case <-a.leaseKeepAlive:
				return
			}
		}
	}()

	_, err = a.etcdCli.Put(ctx, MakeServiceInfoKey(a.selfInfo.ID), string(infoData), clientv3.WithLease(lease.ID))
	if err != nil {
		a.etcdCli.Revoke(ctx, lease.ID)
		return nil, fmt.Errorf("putting service info to etcd: %w", err)
	}

	// 导入当前已有的服务信息和锁信息
	a.services = make(map[string]*serviceStatus)
	for _, svc := range currentServices {
		a.services[svc.ID] = &serviceStatus{
			Info: svc,
		}
	}
	// 直接添加自己的信息
	a.services[a.selfInfo.ID] = &serviceStatus{
		Info: a.selfInfo,
	}

	// 导入锁信息的过程中可能会发现未注册信息的锁服务的锁，把他们挑出来释放掉
	var willReleaseIDs []string
	for _, lock := range currentLocks {
		svc, ok := a.services[lock.SerivceID]
		if !ok {
			willReleaseIDs = append(willReleaseIDs, lock.ID)
			continue
		}

		svc.LockRequestIDs = append(svc.LockRequestIDs, lock.ID)
	}

	return willReleaseIDs, nil
}

func (a *ServiceInfoActor) OnServiceEvent(evt ServiceEvent) error {
	a.lock.Lock()
	defer a.lock.Unlock()

	// TODO 可以考虑打印一点日志

	if evt.IsNew {
		if evt.Info.ID != a.selfInfo.ID {
			logger.Std.WithField("ID", evt.Info.ID).Infof("new service up")
			a.services[evt.Info.ID] = &serviceStatus{
				Info: evt.Info,
			}
		}

	} else {
		logger.Std.WithField("ID", evt.Info.ID).Infof("service down, will release all its locks")

		status, ok := a.services[evt.Info.ID]
		if !ok {
			return nil
		}

		a.releaseActor.DelayRelease(status.LockRequestIDs)

		delete(a.services, evt.Info.ID)

		// 如果收到的被删除服务信息是自己的，那么自己要重启，重新获取全量数据
		if evt.Info.ID == a.selfInfo.ID {
			return ErrSelfServiceDown
		}
	}

	return nil
}

func (a *ServiceInfoActor) OnLockRequestEvent(evt LockRequestEvent) {
	a.lock.Lock()
	defer a.lock.Unlock()

	status, ok := a.services[evt.Data.SerivceID]
	if !ok {
		if evt.IsLocking {
			// 加锁的是一个没有注册过的锁服务，可能是因为这个锁服务之前网络发生了波动，
			// 在波动期间它注册的信息过期，于是被大家认为服务下线，清理掉了它管理的锁，
			// 而在网络恢复回来之后，它还没有意识到自己被认为下线了，于是还在提交锁请求。
			// 为了防止它加了这个锁之后又崩溃，导致的无限锁定，它加的锁我们都直接释放。
			logger.Std.WithField("RequestID", evt.Data.ID).
				WithField("ServiceID", evt.Data.SerivceID).
				Warnf("the lock request is from an unknow service, will release it")

			a.releaseActor.Release([]string{evt.Data.ID})
		}
		return
	}

	if evt.IsLocking {
		status.LockRequestIDs = append(status.LockRequestIDs, evt.Data.ID)
	} else {
		status.LockRequestIDs = lo2.Remove(status.LockRequestIDs, evt.Data.ID)
	}
}
