package internal

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"gitlink.org.cn/cloudream/common/pkg/actor"
	"gitlink.org.cn/cloudream/common/pkg/distlock"
	"gitlink.org.cn/cloudream/common/utils/serder"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

const (
	LOCK_REQUEST_DATA_PREFIX = "/distlock/lockRequest/data"
	LOCK_REQUEST_INDEX       = "/distlock/lockRequest/index"
	LOCK_REQUEST_LOCK_NAME   = "/distlock/lockRequest/lock"
)

type lockData struct {
	Path   []string `json:"path"`
	Name   string   `json:"name"`
	Target string   `json:"target"`
}

type manyAcquireResult struct {
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

	watchEtcdActor *WatchEtcdActor
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

func (a *MainActor) Init(watchEtcdActor *WatchEtcdActor, providersActor *ProvidersActor) {
	a.watchEtcdActor = watchEtcdActor
	a.providersActor = providersActor
}

// Acquire 请求一批锁。成功后返回锁请求ID
func (a *MainActor) Acquire(req distlock.LockRequest) (reqID string, err error) {
	rets, err := a.AcquireMany([]distlock.LockRequest{req})
	if err != nil {
		return "", err
	}

	if rets[0].Err != nil {
		return "", err
	}

	return rets[0].RequestID, nil
}

// AcquireAny 尝试多个锁请求。目前的实现会在第一个获取成功后就直接返回
func (a *MainActor) AcquireMany(reqs []distlock.LockRequest) (rets []manyAcquireResult, err error) {
	return actor.WaitValue(a.commandChan, func() ([]manyAcquireResult, error) {
		// TODO 根据不同的错误设置不同的错误类型，方便上层进行后续处理
		unlock, err := a.acquireEtcdRequestDataLock()
		if err != nil {
			return nil, fmt.Errorf("acquire etcd request data lock failed, err: %w", err)
		}
		defer unlock()

		index, err := a.getEtcdLockRequestIndex()
		if err != nil {
			return nil, err
		}

		// 等待本地状态同步到最新
		// TODO 配置等待时间
		err = a.providersActor.WaitIndexUpdated(index, a.cfg.EtcdLockAcquireTimeoutMs)
		if err != nil {
			return nil, err
		}

		rets := make([]manyAcquireResult, len(reqs))
		for i := 0; i < len(reqs); i++ {
			// 测试锁，并获得锁数据
			reqData, err := a.providersActor.TestLockRequestAndMakeData(reqs[i])
			if err == nil {
				nextIndexStr := strconv.FormatInt(index+1, 10)
				reqData.ID = nextIndexStr

				// 锁成功，提交锁数据
				err := a.submitLockRequest(reqData)

				rets[i] = manyAcquireResult{
					IsTried:   true,
					RequestID: nextIndexStr,
					Err:       err,
				}

				break

			} else {
				rets[i] = manyAcquireResult{
					IsTried: true,
					Err:     err,
				}
			}
		}

		return rets, nil
	})
}

func (a *MainActor) submitLockRequest(reqData LockRequestData) error {
	reqBytes, err := serder.ObjectToJSON(reqData)
	if err != nil {
		return fmt.Errorf("serialize lock request data failed, err: %w", err)
	}

	var etcdOps []clientv3.Op
	if a.cfg.SubmitLockRequestWithoutLease {
		etcdOps = []clientv3.Op{
			clientv3.OpPut(LOCK_REQUEST_INDEX, reqData.ID),
			clientv3.OpPut(makeEtcdLockRequestKey(reqData.ID), string(reqBytes)),
		}

	} else {
		etcdOps = []clientv3.Op{
			clientv3.OpPut(LOCK_REQUEST_INDEX, reqData.ID),
			// 归属到当前连接的租约，在当前连接断开后，能自动解锁
			clientv3.OpPut(makeEtcdLockRequestKey(reqData.ID), string(reqBytes), clientv3.WithLease(a.lockRequestLeaseID)),
		}
	}
	txResp, err := a.etcdCli.Txn(context.Background()).Then(etcdOps...).Commit()
	if err != nil {
		return fmt.Errorf("submit lock request data failed, err: %w", err)
	}
	if !txResp.Succeeded {
		return fmt.Errorf("submit lock request data failed for lock request data index changed")
	}

	return nil
}

// Release 释放锁
func (a *MainActor) Release(reqID string) error {
	return actor.Wait(a.commandChan, func() error {
		// TODO 根据不同的错误设置不同的错误类型，方便上层进行后续处理
		unlock, err := a.acquireEtcdRequestDataLock()
		if err != nil {
			return fmt.Errorf("acquire etcd request data lock failed, err: %w", err)
		}
		defer unlock()

		index, err := a.getEtcdLockRequestIndex()
		if err != nil {
			return err
		}

		lockReqKey := makeEtcdLockRequestKey(reqID)
		delResp, err := a.etcdCli.Delete(context.Background(), lockReqKey)
		if err != nil {
			return fmt.Errorf("delete lock request data failed, err: %w", err)
		}

		if delResp.Deleted == 0 {
			// TODO 可以考虑返回一个更有辨识度的错误
			return fmt.Errorf("lock request data not found")
		}

		nextIndexStr := strconv.FormatInt(index+1, 10)
		_, err = a.etcdCli.Put(context.Background(), LOCK_REQUEST_INDEX, nextIndexStr)
		if err != nil {
			return fmt.Errorf("update lock request data index failed, err: %w", err)
		}

		return nil
	})
}

func (a *MainActor) acquireEtcdRequestDataLock() (unlock func(), err error) {
	lease, err := a.etcdCli.Grant(context.Background(), a.cfg.EtcdLockLeaseTimeSec)
	if err != nil {
		return nil, fmt.Errorf("grant lease failed, err: %w", err)
	}

	session, err := concurrency.NewSession(a.etcdCli, concurrency.WithLease(lease.ID))
	if err != nil {
		return nil, fmt.Errorf("new session failed, err: %w", err)
	}

	mutex := concurrency.NewMutex(session, LOCK_REQUEST_LOCK_NAME)

	timeout, cancelFunc := context.WithTimeout(context.Background(),
		time.Duration(a.cfg.EtcdLockAcquireTimeoutMs)*time.Millisecond)
	defer cancelFunc()

	err = mutex.Lock(timeout)
	if err != nil {
		session.Close()
		return nil, fmt.Errorf("acquire lock failed, err: %w", err)
	}

	return func() {
		mutex.Unlock(context.Background())
		session.Close()
	}, nil
}

func (a *MainActor) getEtcdLockRequestIndex() (int64, error) {
	indexKv, err := a.etcdCli.Get(context.Background(), LOCK_REQUEST_INDEX)
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
	return actor.Wait(a.commandChan, func() error {
		// 使用事务一次性获取index和锁数据，就不需要加全局锁了
		txResp, err := a.etcdCli.Txn(context.Background()).
			Then(
				clientv3.OpGet(LOCK_REQUEST_INDEX),
				clientv3.OpGet(LOCK_REQUEST_DATA_PREFIX, clientv3.WithPrefix()),
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

		// 先停止监听，再重置锁状态，最后恢复监听

		err = a.watchEtcdActor.StopWatching()
		if err != nil {
			return fmt.Errorf("stop watching etcd failed, err: %w", err)
		}

		err = a.providersActor.ResetState(index, reqData)
		if err != nil {
			return fmt.Errorf("reset lock providers state failed, err: %w", err)
		}

		err = a.watchEtcdActor.StartWatching()
		if err != nil {
			return fmt.Errorf("start watching etcd failed, err: %w", err)
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
