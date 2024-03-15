package task

import (
	"sync"
	"sync/atomic"
	"time"

	"gitlink.org.cn/cloudream/common/utils/lo2"
)

type CompleteOption struct {
	// 在Task调用complete函数时调用。调用时被Manager的锁保护。
	Completing func()
	// 延迟删除Manager中的任务，为0时没有延迟，即在Task调用complete函数时立刻删除。
	RemovingDelay time.Duration
}

type CompleteFn = func(err error, opts ...CompleteOption)

type TaskBody[TCtx any] interface {
	Execute(task *Task[TCtx], ctx TCtx, complete CompleteFn)
}

type ComparableTaskBody[TCtx any] interface {
	TaskBody[TCtx]
	Compare(other *Task[TCtx]) bool
}

type Task[TCtx any] struct {
	id          string
	body        TaskBody[TCtx]
	isCompleted atomic.Bool
	waiters     []chan any
	onCompleted []func(task *Task[TCtx])
	waiterLock  sync.Mutex
	err         error
}

func (t *Task[TCtx]) ID() string {
	return t.id
}

func (t *Task[TCtx]) Body() TaskBody[TCtx] {
	return t.body
}

func (t *Task[TCtx]) IsCompleted() bool {
	// 设置err是在Store之前，所以isCompleted为true时一定能获得最新的err
	return t.isCompleted.Load()
}

func (t *Task[TCtx]) Error() error {
	return t.err
}

func (t *Task[TCtx]) Wait() {
	t.waiterLock.Lock()
	if t.isCompleted.Load() {
		t.waiterLock.Unlock()
		return
	}

	waiter := make(chan any)
	t.waiters = append(t.waiters, waiter)
	t.waiterLock.Unlock()

	<-waiter
}

// 限时等待，返回true代表等待成功，返回false代表等待超时
func (t *Task[TCtx]) WaitTimeout(timeout time.Duration) bool {
	t.waiterLock.Lock()
	if t.isCompleted.Load() {
		t.waiterLock.Unlock()
		return true
	}

	waiter := make(chan any)
	t.waiters = append(t.waiters, waiter)
	t.waiterLock.Unlock()

	select {
	case <-time.After(timeout):
		t.waiterLock.Lock()
		t.waiters = lo2.Remove(t.waiters, waiter)
		t.waiterLock.Unlock()

		return false

	case <-waiter:
		return true
	}
}

func (t *Task[TCtx]) OnCompleted(callback func(task *Task[TCtx])) {
	t.waiterLock.Lock()
	if t.isCompleted.Load() {
		t.waiterLock.Unlock()
		callback(t)
		return
	}

	t.onCompleted = append(t.onCompleted, callback)
	t.waiterLock.Unlock()
}
