package internal

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"gitlink.org.cn/cloudream/common/pkgs/future"
	"gitlink.org.cn/cloudream/common/pkgs/logger"
	"gitlink.org.cn/cloudream/common/pkgs/trie"
)

var ErrWaitIndexUpdateTimeout = errors.New("waitting local index updating timeout")

type indexWaiter struct {
	Index    int64
	Callback *future.SetVoidFuture
}

type ProvidersActor struct {
	localLockReqIndex int64
	provdersTrie      trie.Trie[LockProvider]
	allProviders      []LockProvider

	indexWaiters []indexWaiter
	lock         sync.Mutex
}

func NewProvidersActor() *ProvidersActor {
	return &ProvidersActor{}
}

func (a *ProvidersActor) AddProvider(prov LockProvider, path ...any) {
	a.provdersTrie.Create(path).Value = prov
	a.allProviders = append(a.allProviders, prov)
}

func (a *ProvidersActor) Init() {
}

func (a *ProvidersActor) WaitIndexUpdated(ctx context.Context, index int64) error {
	fut := future.NewSetVoid()

	a.lock.Lock()
	if index <= a.localLockReqIndex {
		fut.SetVoid()
	} else {
		a.indexWaiters = append(a.indexWaiters, indexWaiter{
			Index:    index,
			Callback: fut,
		})
	}
	a.lock.Unlock()

	return fut.Wait(ctx)
}

func (a *ProvidersActor) OnLockRequestEvent(evt LockRequestEvent) {
	func() {
		a.lock.Lock()
		defer a.lock.Unlock()

		if evt.IsLocking {
			err := a.lockLockRequest(evt.Data)
			if err != nil {
				// TODO 发生这种错误需要重新加载全量状态，下同
				logger.Std.Warnf("applying locking event: %s", err.Error())
				return
			}

		} else {
			err := a.unlockLockRequest(evt.Data)
			if err != nil {
				logger.Std.Warnf("applying unlocking event: %s", err.Error())
				return
			}
		}

		a.localLockReqIndex++
	}()

	// 检查是否有等待同步进度的需求
	a.wakeUpIndexWaiter()
}

func (svc *ProvidersActor) lockLockRequest(reqData LockRequestData) error {
	for _, lockData := range reqData.Locks {
		node, ok := svc.provdersTrie.WalkEnd(lockData.Path)
		if !ok || node.Value == nil {
			return fmt.Errorf("lock provider not found for path %v", lockData.Path)
		}

		target, err := node.Value.ParseTargetString(lockData.Target)
		if err != nil {
			return fmt.Errorf("parse target data failed, err: %w", err)
		}

		err = node.Value.Lock(reqData.ID, Lock{
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

func (svc *ProvidersActor) unlockLockRequest(reqData LockRequestData) error {
	for _, lockData := range reqData.Locks {
		node, ok := svc.provdersTrie.WalkEnd(lockData.Path)
		if !ok || node.Value == nil {
			return fmt.Errorf("lock provider not found for path %v", lockData.Path)
		}

		target, err := node.Value.ParseTargetString(lockData.Target)
		if err != nil {
			return fmt.Errorf("parse target data failed, err: %w", err)
		}

		err = node.Value.Unlock(reqData.ID, Lock{
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

// TestLockRequestAndMakeData 判断锁能否锁成功，并生成锁数据的字符串表示。注：不会生成请求ID。
// 在检查单个锁是否能上锁时，不会考虑同一个锁请求中的其他的锁影响。简单来说，就是同一个请求中的锁可以互相冲突。
func (a *ProvidersActor) TestLockRequestAndMakeData(req LockRequest) (LockRequestData, error) {
	a.lock.Lock()
	defer a.lock.Unlock()

	reqData := LockRequestData{}

	for _, lock := range req.Locks {
		n, ok := a.provdersTrie.WalkEnd(lock.Path)
		if !ok || n.Value == nil {
			return LockRequestData{}, fmt.Errorf("lock provider not found for path %v", lock.Path)
		}

		err := n.Value.CanLock(lock)
		if err != nil {
			return LockRequestData{}, err
		}

		targetStr, err := n.Value.GetTargetString(lock.Target)
		if err != nil {
			return LockRequestData{}, fmt.Errorf("get lock target string failed, err: %w", err)
		}

		reqData.Locks = append(reqData.Locks, lockData{
			Path:   lock.Path,
			Name:   lock.Name,
			Target: targetStr,
		})
	}

	return reqData, nil
}

func (a *ProvidersActor) ResetState(index int64, lockRequestData []LockRequestData) error {
	a.lock.Lock()
	defer a.lock.Unlock()

	var err error

	for _, p := range a.allProviders {
		p.Clear()
	}

	for _, reqData := range lockRequestData {
		err = a.lockLockRequest(reqData)
		if err != nil {
			err = fmt.Errorf("applying lock request data: %w", err)
			break
		}
	}

	a.localLockReqIndex = index

	// 内部状态已被破坏，停止所有监听器
	for _, w := range a.indexWaiters {
		w.Callback.SetError(ErrWaitIndexUpdateTimeout)
	}
	a.indexWaiters = nil

	return err
}

func (a *ProvidersActor) wakeUpIndexWaiter() {
	var resetWaiters []indexWaiter
	for _, waiter := range a.indexWaiters {
		if waiter.Index <= a.localLockReqIndex {
			waiter.Callback.SetVoid()
		} else {
			resetWaiters = append(resetWaiters, waiter)
		}
	}
	a.indexWaiters = resetWaiters
}
