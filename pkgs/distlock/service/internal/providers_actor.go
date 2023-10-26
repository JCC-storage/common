package internal

import (
	"context"
	"fmt"

	"gitlink.org.cn/cloudream/common/pkgs/actor"
	"gitlink.org.cn/cloudream/common/pkgs/distlock"
	"gitlink.org.cn/cloudream/common/pkgs/future"
	"gitlink.org.cn/cloudream/common/pkgs/logger"
	"gitlink.org.cn/cloudream/common/pkgs/trie"
)

type indexWaiter struct {
	Index  int64
	Future *future.SetVoidFuture
}

type ProvidersActor struct {
	localLockReqIndex int64
	provdersTrie      trie.Trie[distlock.LockProvider]
	allProviders      []distlock.LockProvider

	indexWaiters []indexWaiter

	commandChan *actor.CommandChannel
}

func NewProvidersActor() *ProvidersActor {
	return &ProvidersActor{
		commandChan: actor.NewCommandChannel(),
	}
}

func (a *ProvidersActor) AddProvider(prov distlock.LockProvider, path ...any) {
	a.provdersTrie.Create(path).Value = prov
	a.allProviders = append(a.allProviders, prov)
}

func (a *ProvidersActor) Init() {
}

func (a *ProvidersActor) WaitIndexUpdated(ctx context.Context, index int64) error {
	fut := future.NewSetVoid()

	a.commandChan.Send(func() {
		if index <= a.localLockReqIndex {
			fut.SetVoid()
		} else {
			a.indexWaiters = append(a.indexWaiters, indexWaiter{
				Index:  index,
				Future: fut,
			})
		}
	})

	return fut.Wait(ctx)
}

func (a *ProvidersActor) ApplyLockRequestEvents(events []LockRequestEvent) {
	a.commandChan.Send(func() {
		for _, op := range events {
			if op.IsLocking {
				err := a.lockLockRequest(op.Data)
				if err != nil {
					// TODO 发生这种错误需要重新加载全量状态，下同
					logger.Std.Warnf("applying locking event: %s", err.Error())
					return
				}

			} else {
				err := a.unlockLockRequest(op.Data)
				if err != nil {
					logger.Std.Warnf("applying unlocking event: %s", err.Error())
					return
				}
			}

			// 处理了多少事件，Index就往后移动多少个
			a.localLockReqIndex++
		}

		// 检查是否有等待同步进度的需求
		a.wakeUpIndexWaiter()
	})
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

		err = node.Value.Lock(reqData.ID, distlock.Lock{
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

		err = node.Value.Unlock(reqData.ID, distlock.Lock{
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
func (a *ProvidersActor) TestLockRequestAndMakeData(req distlock.LockRequest) (LockRequestData, error) {
	return actor.WaitValue(context.TODO(), a.commandChan, func() (LockRequestData, error) {
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
	})
}

// ResetState 重置内部状态
func (a *ProvidersActor) ResetState(index int64, lockRequestData []LockRequestData) error {
	return actor.Wait(context.TODO(), a.commandChan, func() error {
		for _, p := range a.allProviders {
			p.Clear()
		}

		for _, reqData := range lockRequestData {
			err := a.lockLockRequest(reqData)
			if err != nil {
				return fmt.Errorf("lock by lock request data failed, err: %w", err)
			}
		}

		a.localLockReqIndex = index

		// 检查是否有等待同步进度的需求
		a.wakeUpIndexWaiter()

		return nil
	})
}

func (a *ProvidersActor) wakeUpIndexWaiter() {
	var resetWaiters []indexWaiter
	for _, waiter := range a.indexWaiters {
		if waiter.Index <= a.localLockReqIndex {
			waiter.Future.SetVoid()
		} else {
			resetWaiters = append(resetWaiters, waiter)
		}
	}
	a.indexWaiters = resetWaiters
}

func (a *ProvidersActor) Serve() error {
	cmdChan := a.commandChan.BeginChanReceive()
	defer a.commandChan.CloseChanReceive()

	for {
		select {
		case cmd, ok := <-cmdChan:
			if !ok {
				return fmt.Errorf("command channel closed")
			}

			cmd()
		}
	}
}
