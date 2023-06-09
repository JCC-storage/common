package distlock

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"gitlink.org.cn/cloudream/common/pkg/actor"
	"gitlink.org.cn/cloudream/common/utils/serder"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

type mainActor struct {
	cfg     *Config
	etcdCli *clientv3.Client

	commandChan *actor.CommandChannel

	watchEtcdActor *watchEtcdActor
	providersActor *providersActor
}

func newMainActor() *mainActor {
	return &mainActor{
		commandChan: actor.NewCommandChannel(),
	}
}

func (a *mainActor) Init(watchEtcdActor *watchEtcdActor, providersActor *providersActor) {
	a.watchEtcdActor = watchEtcdActor
	a.providersActor = providersActor
}

// Acquire 请求一批锁。成功后返回锁请求ID
func (a *mainActor) Acquire(req LockRequest) (reqID string, err error) {
	return actor.WaitValue[string](a.commandChan, func() (string, error) {
		// TODO 根据不同的错误设置不同的错误类型，方便上层进行后续处理
		unlock, err := a.acquireEtcdRequestDataLock()
		if err != nil {
			return "", fmt.Errorf("acquire etcd request data lock failed, err: %w", err)
		}
		defer unlock()

		index, err := a.getEtcdLockRequestIndex()
		if err != nil {
			return "", err
		}

		// 等待本地状态同步到最新
		err = a.providersActor.WaitIndexUpdated(index)
		if err != nil {
			return "", err
		}

		// 测试锁，并获得锁数据
		reqData, err := a.providersActor.TestLockRequestAndMakeData(req)
		if err != nil {
			return "", err
		}

		// 锁成功，提交锁数据

		nextIndexStr := strconv.FormatInt(index+1, 10)

		reqData.ID = nextIndexStr

		reqBytes, err := serder.ObjectToJSON(reqData)
		if err != nil {
			return "", fmt.Errorf("serialize lock request data failed, err: %w", err)
		}

		txResp, err := a.etcdCli.Txn(context.Background()).
			Then(
				clientv3.OpPut(LOCK_REQUEST_INDEX, nextIndexStr),
				clientv3.OpPut(makeEtcdLockRequestKey(nextIndexStr), string(reqBytes)),
			).
			Commit()
		if err != nil {
			return "", fmt.Errorf("submit lock request data failed, err: %w", err)
		}
		if !txResp.Succeeded {
			return "", fmt.Errorf("submit lock request data failed for lock request data index changed")
		}

		return nextIndexStr, nil
	})
}

// Release 释放锁
func (a *mainActor) Release(reqID string) error {
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

func (a *mainActor) acquireEtcdRequestDataLock() (unlock func(), err error) {
	lease, err := a.etcdCli.Grant(context.Background(), a.cfg.EtcdLockLeaseTimeSec)
	if err != nil {
		return nil, fmt.Errorf("grant lease failed, err: %w", err)
	}

	session, err := concurrency.NewSession(a.etcdCli, concurrency.WithLease(lease.ID))
	if err != nil {
		return nil, fmt.Errorf("new session failed, err: %w", err)
	}
	defer session.Close()

	mutex := concurrency.NewMutex(session, LOCK_REQUEST_LOCK_NAME)

	timeout, cancelFunc := context.WithTimeout(context.Background(),
		time.Duration(a.cfg.EtcdLockAcquireTimeoutMs)*time.Millisecond)
	defer cancelFunc()

	err = mutex.Lock(timeout)
	if err != nil {
		return nil, fmt.Errorf("acquire lock failed, err: %w", err)
	}

	return func() {
		mutex.Unlock(context.Background())
		session.Close()
	}, nil
}

func (a *mainActor) getEtcdLockRequestIndex() (int64, error) {
	indexKv, err := a.etcdCli.Get(context.Background(), LOCK_REQUEST_INDEX)
	if err != nil {
		return 0, fmt.Errorf("get lock request index failed, err: %w", err)
	}

	if len(indexKv.Kvs) == 0 {
		return 0, fmt.Errorf("lock request index not found in etcd")
	}

	index, err := strconv.ParseInt(string(indexKv.Kvs[0].Value), 0, 64)
	if err != nil {
		return 0, fmt.Errorf("parse lock request index failed, err: %w", err)
	}

	return index, nil
}

func (a *mainActor) ReloadEtcdData() error {
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
		var reqData []lockRequestData

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
			var req lockRequestData
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

func (a *mainActor) Serve() error {
	for {
		select {
		case cmd, ok := <-a.commandChan.ChanReceive():
			if !ok {
				return fmt.Errorf("command channel closed")
			}

			cmd()
		}
	}
}
