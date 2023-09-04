package internal

import (
	"context"
	"fmt"
	"strconv"

	"gitlink.org.cn/cloudream/common/pkgs/actor"
	"gitlink.org.cn/cloudream/common/pkgs/distlock"
	"gitlink.org.cn/cloudream/common/utils/serder"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/clientv3util"
	"go.etcd.io/etcd/client/v3/concurrency"
)

const (
	EtcdLockRequestData  = "/distlock/lockRequest/data"
	EtcdLockRequestIndex = "/distlock/lockRequest/index"
	EtcdLockRequestLock  = "/distlock/lockRequest/lock"
)

type lockData struct {
	Path   []string `json:"path"`
	Name   string   `json:"name"`
	Target string   `json:"target"`
}

type acquireManyResult struct {
	IsTried   bool
	RequestID string
	Err       error
}

type LockRequestData struct {
	ID    string     `json:"id"`
	Locks []lockData `json:"locks"`
}

type MainActor struct {
	cfg     *distlock.Config
	etcdCli *clientv3.Client

	commandChan *actor.CommandChannel

	providersActor *ProvidersActor

	lockRequestLeaseID clientv3.LeaseID
}

func NewMainActor(cfg *distlock.Config, etcdCli *clientv3.Client) *MainActor {
	return &MainActor{
		cfg:         cfg,
		etcdCli:     etcdCli,
		commandChan: actor.NewCommandChannel(),
	}
}

func (a *MainActor) Init(providersActor *ProvidersActor) {
	a.providersActor = providersActor
}

// Acquire 请求一批锁。成功后返回锁请求ID
func (a *MainActor) Acquire(ctx context.Context, req distlock.LockRequest) (reqID string, err error) {
	rets, err := a.AcquireMany(ctx, []distlock.LockRequest{req})
	if err != nil {
		return "", err
	}

	if rets[0].Err != nil {
		return "", rets[0].Err
	}

	return rets[0].RequestID, nil
}

// AcquireAny 尝试多个锁请求。目前的实现会在第一个获取成功后就直接返回
func (a *MainActor) AcquireMany(ctx context.Context, reqs []distlock.LockRequest) (rets []acquireManyResult, err error) {
	return actor.WaitValue(context.TODO(), a.commandChan, func() ([]acquireManyResult, error) {
		// TODO 根据不同的错误设置不同的错误类型，方便上层进行后续处理
		unlock, err := a.acquireEtcdRequestDataLock(ctx)
		if err != nil {
			return nil, fmt.Errorf("acquire etcd request data lock failed, err: %w", err)
		}
		defer unlock()

		index, err := a.getEtcdLockRequestIndex(ctx)
		if err != nil {
			return nil, err
		}

		// 等待本地状态同步到最新
		// TODO 配置等待时间
		err = a.providersActor.WaitIndexUpdated(ctx, index)
		if err != nil {
			return nil, err
		}

		rets := make([]acquireManyResult, len(reqs))
		for i := 0; i < len(reqs); i++ {
			// 测试锁，并获得锁数据
			reqData, err := a.providersActor.TestLockRequestAndMakeData(reqs[i])
			if err == nil {
				nextIndexStr := strconv.FormatInt(index+1, 10)
				reqData.ID = nextIndexStr

				// 锁成功，提交锁数据
				err := a.submitLockRequest(ctx, reqData)

				rets[i] = acquireManyResult{
					IsTried:   true,
					RequestID: nextIndexStr,
					Err:       err,
				}

				break

			} else {
				rets[i] = acquireManyResult{
					IsTried: true,
					Err:     err,
				}
			}
		}

		return rets, nil
	})
}

func (a *MainActor) submitLockRequest(ctx context.Context, reqData LockRequestData) error {
	reqBytes, err := serder.ObjectToJSON(reqData)
	if err != nil {
		return fmt.Errorf("serialize lock request data failed, err: %w", err)
	}

	var etcdOps []clientv3.Op
	if a.cfg.SubmitLockRequestWithoutLease {
		etcdOps = []clientv3.Op{
			clientv3.OpPut(EtcdLockRequestIndex, reqData.ID),
			clientv3.OpPut(makeEtcdLockRequestKey(reqData.ID), string(reqBytes)),
		}

	} else {
		etcdOps = []clientv3.Op{
			clientv3.OpPut(EtcdLockRequestIndex, reqData.ID),
			// 归属到当前连接的租约，在当前连接断开后，能自动解锁
			// TODO 不能直接给RequestData上租约，因为如果在别的服务已经获取到锁的情况下，
			// 如果当前服务崩溃，删除消息会立刻发送出去，这就破坏了锁的约定（在锁定期间其他服务不能修改数据）
			clientv3.OpPut(makeEtcdLockRequestKey(reqData.ID), string(reqBytes)), //, clientv3.WithLease(a.lockRequestLeaseID)),
		}
	}
	txResp, err := a.etcdCli.Txn(ctx).Then(etcdOps...).Commit()
	if err != nil {
		return fmt.Errorf("submit lock request data failed, err: %w", err)
	}
	if !txResp.Succeeded {
		return fmt.Errorf("submit lock request data failed for lock request data index changed")
	}

	return nil
}

// Release 释放锁
func (a *MainActor) Release(ctx context.Context, reqID string) error {
	return actor.Wait(context.TODO(), a.commandChan, func() error {
		// TODO 根据不同的错误设置不同的错误类型，方便上层进行后续处理
		unlock, err := a.acquireEtcdRequestDataLock(ctx)
		if err != nil {
			return fmt.Errorf("acquire etcd request data lock failed, err: %w", err)
		}
		defer unlock()

		index, err := a.getEtcdLockRequestIndex(ctx)
		if err != nil {
			return err
		}

		lockReqKey := makeEtcdLockRequestKey(reqID)

		txResp, err := a.etcdCli.Txn(ctx).
			If(clientv3util.KeyExists(lockReqKey)).
			Then(clientv3.OpDelete(lockReqKey), clientv3.OpPut(EtcdLockRequestIndex, strconv.FormatInt(index+1, 10))).Commit()
		if err != nil {
			return fmt.Errorf("updating lock request data index: %w", err)
		}
		if !txResp.Succeeded {
			return fmt.Errorf("updating lock request data failed")
		}

		return nil
	})
}

func (a *MainActor) acquireEtcdRequestDataLock(ctx context.Context) (unlock func(), err error) {
	lease, err := a.etcdCli.Grant(context.Background(), a.cfg.EtcdLockLeaseTimeSec)
	if err != nil {
		return nil, fmt.Errorf("grant lease failed, err: %w", err)
	}

	session, err := concurrency.NewSession(a.etcdCli, concurrency.WithLease(lease.ID))
	if err != nil {
		return nil, fmt.Errorf("new session failed, err: %w", err)
	}

	mutex := concurrency.NewMutex(session, EtcdLockRequestLock)

	err = mutex.Lock(ctx)
	if err != nil {
		session.Close()
		return nil, fmt.Errorf("acquire lock failed, err: %w", err)
	}

	return func() {
		mutex.Unlock(context.Background())
		session.Close()
	}, nil
}

func (a *MainActor) getEtcdLockRequestIndex(ctx context.Context) (int64, error) {
	indexKv, err := a.etcdCli.Get(ctx, EtcdLockRequestIndex)
	if err != nil {
		return 0, fmt.Errorf("get lock request index failed, err: %w", err)
	}

	if len(indexKv.Kvs) == 0 {
		return 0, nil
	}

	index, err := strconv.ParseInt(string(indexKv.Kvs[0].Value), 0, 64)
	if err != nil {
		return 0, fmt.Errorf("parse lock request index failed, err: %w", err)
	}

	return index, nil
}

func (a *MainActor) ReloadEtcdData() error {
	return actor.Wait(context.TODO(), a.commandChan, func() error {
		// 使用事务一次性获取index和锁数据，就不需要加全局锁了
		txResp, err := a.etcdCli.Txn(context.Background()).
			Then(
				clientv3.OpGet(EtcdLockRequestIndex),
				clientv3.OpGet(EtcdLockRequestData, clientv3.WithPrefix()),
			).
			Commit()
		if err != nil {
			return fmt.Errorf("get etcd data failed, err: %w", err)
		}
		if !txResp.Succeeded {
			return fmt.Errorf("get etcd data failed")
		}

		indexKvs := txResp.Responses[0].GetResponseRange().Kvs
		lockKvs := txResp.Responses[1].GetResponseRange().Kvs

		var index int64
		var reqData []LockRequestData

		// 解析Index
		if len(indexKvs) > 0 {
			val, err := strconv.ParseInt(string(indexKvs[0].Value), 0, 64)
			if err != nil {
				return fmt.Errorf("parse lock request index failed, err: %w", err)
			}
			index = val

		} else {
			index = 0
		}

		// 解析锁请求数据
		for _, kv := range lockKvs {
			var req LockRequestData
			err := serder.JSONToObject(kv.Value, &req)
			if err != nil {
				return fmt.Errorf("parse lock request data failed, err: %w", err)
			}

			reqData = append(reqData, req)
		}

		err = a.providersActor.ResetState(index, reqData)
		if err != nil {
			return fmt.Errorf("reset lock providers state failed, err: %w", err)
		}

		return nil
	})
}

func (a *MainActor) Serve() error {
	lease, err := a.etcdCli.Grant(context.Background(), a.cfg.EtcdLockLeaseTimeSec)
	if err != nil {
		return fmt.Errorf("grant lease failed, err: %w", err)
	}
	a.lockRequestLeaseID = lease.ID

	cmdChan := a.commandChan.BeginChanReceive()
	defer a.commandChan.CloseChanReceive()

	for {
		select {
		case cmd, ok := <-cmdChan:
			if !ok {
				return fmt.Errorf("command channel closed")
			}

			// TODO Actor启动时，如果第一个调用的是Acquire，那么就会在Acquire中等待本地锁数据同步到最新。
			// 此时命令的执行也会被阻塞，导致ReloadEtcdData命令无法执行，因此产生死锁，最后Acquire超时失败。
			// 此处暂时使用单独的goroutine的来执行命令，避免阻塞。
			go cmd()
		}
	}
}
