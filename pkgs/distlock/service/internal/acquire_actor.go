package internal

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"

	"gitlink.org.cn/cloudream/common/pkgs/distlock"
	"gitlink.org.cn/cloudream/common/pkgs/future"
	"gitlink.org.cn/cloudream/common/pkgs/logger"
	mylo "gitlink.org.cn/cloudream/common/utils/lo"
	"gitlink.org.cn/cloudream/common/utils/serder"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

var ErrAcquiringTimeout = errors.New("acquiring timeout")

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
type LockRequestData struct {
	ID    string     `json:"id"`
	Locks []lockData `json:"locks"`
}

type acquireInfo struct {
	Request  distlock.LockRequest
	Callback *future.SetValueFuture[string]
	LastErr  error
}

type AcquireActor struct {
	cfg            *distlock.Config
	etcdCli        *clientv3.Client
	providersActor *ProvidersActor

	acquirings []*acquireInfo
	lock       sync.Mutex
}

func NewAcquireActor(cfg *distlock.Config, etcdCli *clientv3.Client) *AcquireActor {
	return &AcquireActor{
		cfg:     cfg,
		etcdCli: etcdCli,
	}
}

func (a *AcquireActor) Init(providersActor *ProvidersActor) {
	a.providersActor = providersActor
}

// Acquire 请求一批锁。成功后返回锁请求ID
func (a *AcquireActor) Acquire(ctx context.Context, req distlock.LockRequest) (string, error) {
	info := &acquireInfo{
		Request:  req,
		Callback: future.NewSetValue[string](),
	}

	func() {
		a.lock.Lock()
		defer a.lock.Unlock()

		a.acquirings = append(a.acquirings, info)
		// TODO 处理错误
		err := a.doAcquiring()
		if err != nil {
			logger.Std.Debugf("doing acquiring: %s", err.Error())
		}
	}()

	go func() {
		info.Callback.Wait(ctx)

		a.lock.Lock()
		defer a.lock.Unlock()

		// 调用Callback时都加了锁，所以此处的IsComplete判断可以作为后续操作的依据
		if info.Callback.IsComplete() {
			return
		}

		a.acquirings = mylo.Remove(a.acquirings, info)
		if info.LastErr != nil {
			info.Callback.SetError(info.LastErr)
		} else {
			info.Callback.SetError(ErrAcquiringTimeout)
		}
	}()

	// 此处不能直接用ctx去等Callback，原因是Wait超时不代表锁没有获取到，这会导致锁泄露。
	return info.Callback.WaitValue(context.Background())
}

// TryAcquireNow 重试一下内部还没有成功的锁请求。不会阻塞调用者
func (a *AcquireActor) TryAcquireNow() {
	go func() {
		a.lock.Lock()
		defer a.lock.Unlock()

		err := a.doAcquiring()
		if err != nil {
			logger.Std.Debugf("doing acquiring: %s", err.Error())
		}
	}()
}

func (a *AcquireActor) doAcquiring() error {
	ctx := context.Background()

	if len(a.acquirings) == 0 {
		return nil
	}

	// TODO 根据不同的错误设置不同的错误类型，方便上层进行后续处理
	unlock, err := acquireEtcdRequestDataLock(ctx, a.etcdCli, a.cfg.EtcdLockLeaseTimeSec)
	if err != nil {
		return fmt.Errorf("acquire etcd request data lock failed, err: %w", err)
	}
	defer unlock()

	index, err := getEtcdLockRequestIndex(ctx, a.etcdCli)
	if err != nil {
		return err
	}

	// 等待本地状态同步到最新
	// TODO 配置等待时间
	err = a.providersActor.WaitIndexUpdated(ctx, index)
	if err != nil {
		return err
	}

	// TODO 可以考虑一次性获得多个锁
	for i := 0; i < len(a.acquirings); i++ {
		// 测试锁，并获得锁数据
		reqData, err := a.providersActor.TestLockRequestAndMakeData(a.acquirings[i].Request)
		if err != nil {
			a.acquirings[i].LastErr = err
			continue
		}

		nextIndexStr := strconv.FormatInt(index+1, 10)
		reqData.ID = nextIndexStr

		// 锁成功，提交锁数据
		err = a.submitLockRequest(ctx, nextIndexStr, reqData)
		if err != nil {
			a.acquirings[i].LastErr = err
			continue
		}

		a.acquirings[i].Callback.SetValue(reqData.ID)
		a.acquirings = mylo.RemoveAt(a.acquirings, i)
		break
	}

	return nil
}

func (a *AcquireActor) submitLockRequest(ctx context.Context, index string, reqData LockRequestData) error {
	reqBytes, err := serder.ObjectToJSON(reqData)
	if err != nil {
		return fmt.Errorf("serialize lock request data failed, err: %w", err)
	}

	etcdOps := []clientv3.Op{
		clientv3.OpPut(EtcdLockRequestIndex, index),
		clientv3.OpPut(makeEtcdLockRequestKey(reqData.ID), string(reqBytes)),
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

func acquireEtcdRequestDataLock(ctx context.Context, etcdCli *clientv3.Client, etcdLockLeaseTimeSec int64) (unlock func(), err error) {
	lease, err := etcdCli.Grant(context.Background(), etcdLockLeaseTimeSec)
	if err != nil {
		return nil, fmt.Errorf("grant lease failed, err: %w", err)
	}

	session, err := concurrency.NewSession(etcdCli, concurrency.WithLease(lease.ID))
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

func getEtcdLockRequestIndex(ctx context.Context, etcdCli *clientv3.Client) (int64, error) {
	indexKv, err := etcdCli.Get(ctx, EtcdLockRequestIndex)
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
