package service

import (
	"fmt"

	"gitlink.org.cn/cloudream/common/pkg/actor"
	"gitlink.org.cn/cloudream/common/pkg/distlock"
	"gitlink.org.cn/cloudream/common/pkg/future"
	"gitlink.org.cn/cloudream/common/pkg/trie"
)

type indexWaiter struct {
	Index  int64
	Future *future.SetVoidFuture
}

type lockRequestDataUpdateOp struct {
	Data   lockRequestData
	IsLock bool
}

type providersActor struct {
	localLockReqIndex int64
	provdersTrie      trie.Trie[distlock.LockProvider]
	allProviders      []distlock.LockProvider

	indexWaiters []indexWaiter

	commandChan *actor.CommandChannel
}

func newProvidersActor() *providersActor {
	return &providersActor{
		commandChan: actor.NewCommandChannel(),
	}
}

func (a *providersActor) Init() {
}

func (a *providersActor) WaitIndexUpdated(index int64) error {
	fut := future.NewSetVoid()

	a.commandChan.Send(func() {
		if a.localLockReqIndex <= index {
			fut.SetVoid()
		} else {
			a.indexWaiters = append(a.indexWaiters, indexWaiter{
				Index:  index,
				Future: fut,
			})
		}
	})

	return fut.Wait()
}

func (a *providersActor) BatchUpdateByLockRequestData(ops []lockRequestDataUpdateOp) error {
	return actor.Wait(a.commandChan, func() error {
		for _, op := range ops {
			if op.IsLock {
				err := a.lockLockRequest(op.Data)
				if err != nil {
					return fmt.Errorf("lock by lock request data failed, err: %w", err)
				}

			} else {
				err := a.unlockLockRequest(op.Data)
				if err != nil {
					return fmt.Errorf("unlock by lock request data failed, err: %w", err)
				}
			}
		}

		// 处理了多少事件，Index就往后移动多少个
		a.localLockReqIndex += int64(len(ops))

		// 检查是否有等待同步进度的需求
		a.checkIndexWaiter()

		return nil
	})
}

func (svc *providersActor) lockLockRequest(reqData lockRequestData) error {
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

func (svc *providersActor) unlockLockRequest(reqData lockRequestData) error {
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

// TestLockRequestAndMakeData 判断锁能否锁成功，并生成锁数据的字符串表示。注：不会生成请求ID
func (a *providersActor) TestLockRequestAndMakeData(req distlock.LockRequest) (lockRequestData, error) {
	return actor.WaitValue[lockRequestData](a.commandChan, func() (lockRequestData, error) {
		reqData := lockRequestData{}

		for _, lock := range req.Locks {
			n, ok := a.provdersTrie.WalkEnd(lock.Path)
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
	})
}

// ResetState 重置内部状态
func (a *providersActor) ResetState(index int64, lockRequestData []lockRequestData) error {
	return actor.Wait(a.commandChan, func() error {
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
		a.checkIndexWaiter()

		return nil
	})
}

func (a *providersActor) checkIndexWaiter() {
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

func (a *providersActor) Serve() error {
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
