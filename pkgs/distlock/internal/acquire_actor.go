package internal

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"gitlink.org.cn/cloudream/common/pkgs/future"
	"gitlink.org.cn/cloudream/common/pkgs/logger"
	"gitlink.org.cn/cloudream/common/utils/lo2"
	"gitlink.org.cn/cloudream/common/utils/serder"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

var ErrAcquiringTimeout = errors.New("acquiring timeout")

type acquireInfo struct {
	Request  LockRequest
	Callback *future.SetValueFuture[string]
	LastErr  error
}

type AcquireActor struct {
	cfg            *Config
	etcdCli        *clientv3.Client
	providersActor *ProvidersActor

	isMaintenance   bool
	serviceID       string
	acquirings      []*acquireInfo
	lock            sync.Mutex
	doAcquiringChan chan any
}

func NewAcquireActor(cfg *Config, etcdCli *clientv3.Client) *AcquireActor {
	return &AcquireActor{
		cfg:             cfg,
		etcdCli:         etcdCli,
		isMaintenance:   true,
		doAcquiringChan: make(chan any, 1),
	}
}

func (a *AcquireActor) Init(providersActor *ProvidersActor) {
	a.providersActor = providersActor
}

// Acquire 请求一批锁。成功后返回锁请求ID
func (a *AcquireActor) Acquire(ctx context.Context, req LockRequest) (string, error) {
	info := &acquireInfo{
		Request:  req,
		Callback: future.NewSetValue[string](),
	}

	func() {
		a.lock.Lock()
		defer a.lock.Unlock()

		a.acquirings = append(a.acquirings, info)

		// 如果处于维护模式，那么只接受请求，不实际去处理
		if a.isMaintenance {
			return
		}

		select {
		case a.doAcquiringChan <- nil:
		default:
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

		a.acquirings = lo2.Remove(a.acquirings, info)
		if info.LastErr != nil {
			info.Callback.SetError(info.LastErr)
		} else {
			info.Callback.SetError(ErrAcquiringTimeout)
		}
	}()

	// 此处不能直接用ctx去等Callback，原因是Wait超时不代表锁没有获取到，这会导致锁泄露。
	return info.Callback.Wait(context.Background())
}

// TryAcquireNow 重试一下内部还没有成功的锁请求。不会阻塞调用者
func (a *AcquireActor) TryAcquireNow() {
	go func() {
		a.lock.Lock()
		defer a.lock.Unlock()

		// 处于维护模式中时，即使是主动触发Acqurire也不予理会
		if a.isMaintenance {
			return
		}

		select {
		case a.doAcquiringChan <- nil:
		default:
		}
	}()
}

// 进入维护模式。维护模式期间只接受请求，不处理请求。
func (a *AcquireActor) EnterMaintenance() {
	a.lock.Lock()
	defer a.lock.Unlock()

	a.isMaintenance = true
}

// 退出维护模式。退出之后建议调用一下TryAcquireNow。
func (a *AcquireActor) LeaveMaintenance() {
	a.lock.Lock()
	defer a.lock.Unlock()

	a.isMaintenance = false
}

func (a *AcquireActor) ResetState(serviceID string) {
	a.lock.Lock()
	defer a.lock.Unlock()

	a.serviceID = serviceID
}

func (a *AcquireActor) Serve() {
	for {
		// 离开了select块之后doAcquiringChan的buf就会空出来，
		// 如果之后成功提交了一个锁请求，那么WatchEtcd会收到事件，然后调用此Actor的回调再次设置doAcquiringChan。
		// 因此无论多少个锁请求同时提交，或者是在doAcquiring期间提交，都不会因为某一个成功了，其他的就连试都不试就开始等待。
		// 如果没有一个锁请求提交成功，那自然是已经尝试过所有锁请求了，此时等待新事件到来后F再来尝试也是合理的。
		select {
		case <-a.doAcquiringChan:
		}

		// 如果没有锁请求，那么就不需要进行加锁操作
		a.lock.Lock()
		if len(a.acquirings) == 0 {
			a.lock.Unlock()
			continue
		}
		a.lock.Unlock()

		err := a.doAcquiring()
		if err != nil {
			logger.Std.Debugf("doing acquiring: %s", err.Error())
		}
	}
}

// 返回true代表成功提交了一个锁
func (a *AcquireActor) doAcquiring() error {
	// TODO 配置等待时间
	ctx := context.Background()

	// 在获取全局锁的时候不用锁Actor，只有获取成功了，才加锁
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
	err = a.providersActor.WaitLocalIndexTo(ctx, index)
	if err != nil {
		return err
	}

	a.lock.Lock()
	defer a.lock.Unlock()
	// TODO 可以考虑一次性获得多个锁
	for i := 0; i < len(a.acquirings); i++ {
		req := a.acquirings[i]

		// 测试锁，并获得锁数据
		reqData, err := a.providersActor.TestLockRequestAndMakeData(req.Request)
		if err != nil {
			req.LastErr = err
			continue
		}

		nextIndexStr := strconv.FormatInt(index+1, 10)
		reqData.ID = nextIndexStr
		reqData.SerivceID = a.serviceID
		reqData.Reason = req.Request.Reason
		reqData.Timestamp = time.Now().Unix()

		// 锁成功，提交锁数据
		err = a.submitLockRequest(ctx, nextIndexStr, reqData)
		if err != nil {
			req.LastErr = err
			continue
		}

		req.Callback.SetValue(reqData.ID)
		a.acquirings = lo2.RemoveAt(a.acquirings, i)
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
		clientv3.OpPut(MakeEtcdLockRequestKey(reqData.ID), string(reqBytes)),
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
