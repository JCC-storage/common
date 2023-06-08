package distlock

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"gitlink.org.cn/cloudream/common/pkg/trie"
	"gitlink.org.cn/cloudream/common/utils/serder"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

const (
	LOCK_REQUEST_DATA_PREFIX = "/distlock/lockRequest/data"
	LOCK_REQUEST_INDEX       = "/distlock/lockRequest/index"
	LOCK_REQUEST_LOCK_NAME   = "/distlock/lockRequest/lock"
)

type Service struct {
	cfg     *Config
	etcdCli *clientv3.Client

	providersLock             sync.Mutex
	provdersTrie              trie.Trie[LockProvider]
	allProviders              []LockProvider
	localLockReqIndex         int64
	waitLocalLockReqIndex     int64
	waitLocalLockReqIndexChan chan any
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

	return &Service{
		cfg:     cfg,
		etcdCli: etcdCli,
	}, nil
}

// Acquire 请求一批锁。成功后返回锁请求ID
func (svc *Service) Acquire(req LockRequest) (reqID string, err error) {
	// TODO 根据不同的错误设置不同的错误类型，方便上层进行后续处理
	unlock, err := svc.lockEtcdRequestData()
	if err != nil {
		return "", fmt.Errorf("acquire etcd request data lock failed, err: %w", err)
	}
	defer unlock()

	index, err := svc.getEtcdLockRequestIndex()
	if err != nil {
		return "", err
	}

	// 测试锁，并获得锁数据
	reqData, err := svc.testLockRequestAndMakeData(index, req)
	if err != nil {
		return "", err
	}

	// 锁成功，提交锁数据

	nextIndexStr := strconv.FormatInt(index+1, 10)

	reqBytes, err := serder.ObjectToJSON(reqData)
	if err != nil {
		return "", fmt.Errorf("serialize lock request data failed, err: %w", err)
	}

	txResp, err := svc.etcdCli.Txn(context.Background()).
		// 文档上没有说明如果If为空，会执行Then还是Else，所以为了避免问题，设定一个恒成立的条件。
		// 正常情况下，锁定全局锁期间index是不可能变化的，所以下面这个条件一定成立。
		// 注：由于是字符串比较，所以修改此值的时候，必须保证是10进制，且无前后空格。
		If(clientv3.Compare(clientv3.Value(LOCK_REQUEST_INDEX), "=", strconv.FormatInt(index, 10))).
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
}

func (svc *Service) getEtcdLockRequestIndex() (int64, error) {
	indexKv, err := svc.etcdCli.Get(context.Background(), LOCK_REQUEST_INDEX)
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

func (svc *Service) testLockRequestAndMakeData(latestIndex int64, req LockRequest) (lockRequestData, error) {
	svc.providersLock.Lock()
	defer svc.providersLock.Unlock()

	// 等待本地状态同步到最新
	if svc.localLockReqIndex < latestIndex {
		ch := make(chan any, 1)
		svc.waitLocalLockReqIndex = latestIndex
		svc.waitLocalLockReqIndexChan = ch

		svc.providersLock.Unlock()

		// TODO 超时
		<-ch

		// 等待完全同步完成，那么再次加锁，防止本地状态被更改。
		// 设计上来说，锁定了etcd中的全局锁之后，不可能再有更改的事件发生，因此只要本地状态同步到了最新，
		// watch协程就不会再收到事件，然后更改本地状态，但跨协程修改本地状态存在内存可见性问题，所以还是需要加锁来同步一下
		svc.providersLock.Lock()
	}

	// 判断锁能否锁成功，并生成锁数据的字符串表示
	reqData := lockRequestData{}

	for _, lock := range req.Locks {
		n, ok := svc.provdersTrie.WalkEnd(lock.Path)
		if !ok || n.Value == nil {
			return lockRequestData{}, fmt.Errorf("lock provider not found for path %v", lock.Path)
		}

		err := n.Value.CanLock(lock)
		if err != nil {
			return lockRequestData{}, err
		}

		targetStr, err := n.Value.GetTargetString(lock.Target)
		if err != nil {
			return lockRequestData{}, fmt.Errorf("get lock target string failed, err: %w", err)
		}

		reqData.Locks = append(reqData.Locks, lockData{
			Path:   lock.Path,
			Name:   lock.Name,
			Target: targetStr,
		})
	}

	return reqData, nil
}

// Renew 续约锁
func (svc *Service) Renew(lockReqID string) error {
	panic("todo")

}

// Release 释放锁
func (svc *Service) Release(lockReqID string) error {
	panic("todo")

}

func (svc *Service) Serve() error {
	return svc.watchRequestData()
}

func (svc *Service) lockEtcdRequestData() (unlock func(), err error) {
	lease, err := svc.etcdCli.Grant(context.Background(), svc.cfg.LockRequestDataConfig.LeaseTimeSec)
	if err != nil {
		return nil, fmt.Errorf("grant lease failed, err: %w", err)
	}

	session, err := concurrency.NewSession(svc.etcdCli, concurrency.WithLease(lease.ID))
	if err != nil {
		return nil, fmt.Errorf("new session failed, err: %w", err)
	}
	defer session.Close()

	mutex := concurrency.NewMutex(session, LOCK_REQUEST_LOCK_NAME)

	timeout, cancelFunc := context.WithTimeout(context.Background(),
		time.Duration(svc.cfg.LockRequestDataConfig.AcquireTimeoutMs)*time.Millisecond)
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

func (svc *Service) watchRequestData() error {
	// TODO 考虑增加状态字段，调用API时根据状态字段来判断能不能调用成功
	err := svc.loadInitData()
	if err != nil {
		return fmt.Errorf("load init data failed, err: %w", err)
	}

	dataWatchChan := svc.etcdCli.Watch(context.Background(), LOCK_REQUEST_DATA_PREFIX, clientv3.WithPrefix())

	for {
		select {
		case msg := <-dataWatchChan:
			if msg.Canceled {
				return fmt.Errorf("watch canceled, err: %w", msg.Err())
			}

			err := svc.applyEvents(msg)
			if err != nil {
				return err
			}
		}
	}
}

func (svc *Service) loadInitData() error {
	index, locks, err := svc.getInitDataFromEtcd()
	if err != nil {
		return fmt.Errorf("get init data from etcd failed, err: %w", err)
	}

	err = svc.resetLocalLockRequestData(index, locks)
	if err != nil {
		return fmt.Errorf("reset local lock request data failed, err: %w", err)
	}

	return nil
}

func (svc *Service) getInitDataFromEtcd() ([]*mvccpb.KeyValue, []*mvccpb.KeyValue, error) {
	unlock, err := svc.lockEtcdRequestData()
	if err != nil {
		return nil, nil, fmt.Errorf("try lock request data failed, err: %w", err)
	}
	defer unlock()

	index, err := svc.etcdCli.Get(context.Background(), LOCK_REQUEST_INDEX)
	if err != nil {
		return nil, nil, fmt.Errorf("get lock request index failed, err: %w", err)
	}

	data, err := svc.etcdCli.Get(context.Background(), LOCK_REQUEST_DATA_PREFIX, clientv3.WithPrefix())
	if err != nil {
		return nil, nil, fmt.Errorf("get lock request data failed, err: %w", err)
	}

	return index.Kvs, data.Kvs, nil
}

func (svc *Service) resetLocalLockRequestData(index []*mvccpb.KeyValue, locks []*mvccpb.KeyValue) error {
	svc.providersLock.Lock()
	defer svc.providersLock.Unlock()

	// 先清空所有的锁数据
	for _, p := range svc.allProviders {
		p.Clear()
	}

	// 然后再导入全量数据
	for _, kv := range locks {
		err := svc.lockLockRequest(kv)
		if err != nil {
			return err
		}
	}

	// 更新本地index
	if len(index) == 0 {
		svc.localLockReqIndex = 0

	} else {
		val, err := strconv.ParseInt(string(index[0].Value), 0, 64)
		if err != nil {
			return fmt.Errorf("parse lock request index failed, err: %w", err)
		}

		svc.localLockReqIndex = val
	}

	// 检查是否有等待同步进度的需求
	if svc.waitLocalLockReqIndexChan != nil && svc.waitLocalLockReqIndex <= svc.localLockReqIndex {
		close(svc.waitLocalLockReqIndexChan)
		svc.waitLocalLockReqIndexChan = nil
	}

	return nil
}

func (svc *Service) applyEvents(watchResp clientv3.WatchResponse) error {
	handledCnt := 0

	svc.providersLock.Lock()
	defer svc.providersLock.Unlock()

	for _, e := range watchResp.Events {
		var err error

		// 只监听新建和删除的事件，因为在设计上约定只有这两种事件才会影响Index
		if e.Type == clientv3.EventTypeDelete {
			err = svc.unlockLockRequest(e.Kv)
			handledCnt++

		} else if e.IsCreate() {
			err = svc.lockLockRequest(e.Kv)
			handledCnt++
		}

		if err != nil {
			return fmt.Errorf("apply event failed, err: %w", err)
		}
	}

	// 处理了多少事件，Index就往后移动多少个
	svc.localLockReqIndex += int64(handledCnt)

	// 检查是否有等待同步进度的需求
	if svc.waitLocalLockReqIndexChan != nil && svc.waitLocalLockReqIndex <= svc.localLockReqIndex {
		close(svc.waitLocalLockReqIndexChan)
		svc.waitLocalLockReqIndexChan = nil
	}

	return nil
}

func (svc *Service) lockLockRequest(kv *mvccpb.KeyValue) error {
	reqID := getLockRequestID(string(kv.Key))

	var req lockRequestData
	err := serder.JSONToObject(kv.Value, &req)
	if err != nil {
		return fmt.Errorf("parse lock request data")
	}

	for _, lockData := range req.Locks {
		node, ok := svc.provdersTrie.WalkEnd(lockData.Path)
		if !ok || node.Value == nil {
			return fmt.Errorf("lock provider not found for path %v", lockData.Path)
		}

		target, err := node.Value.ParseTargetString(lockData.Target)
		if err != nil {
			return fmt.Errorf("parse target data failed, err: %w", err)
		}

		err = node.Value.Lock(reqID, Lock{
			Path:   lockData.Path,
			Name:   lockData.Name,
			Target: target,
		})
		if err != nil {
			return fmt.Errorf("locking failed, err: %w", err)
		}
	}
	return nil
}

func (svc *Service) unlockLockRequest(kv *mvccpb.KeyValue) error {
	reqID := getLockRequestID(string(kv.Key))

	var req lockRequestData
	err := serder.JSONToObject(kv.Value, &req)
	if err != nil {
		return fmt.Errorf("parse lock request data")
	}

	for _, lockData := range req.Locks {
		node, ok := svc.provdersTrie.WalkEnd(lockData.Path)
		if !ok || node.Value == nil {
			return fmt.Errorf("lock provider not found for path %v", lockData.Path)
		}

		target, err := node.Value.ParseTargetString(lockData.Target)
		if err != nil {
			return fmt.Errorf("parse target data failed, err: %w", err)
		}

		err = node.Value.Unlock(reqID, Lock{
			Path:   lockData.Path,
			Name:   lockData.Name,
			Target: target,
		})
		if err != nil {
			return fmt.Errorf("unlocking failed, err: %w", err)
		}
	}
	return nil
}
