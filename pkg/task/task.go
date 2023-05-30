package task

import "sync"

type CompleteFn = func(err error, completing func())

type TaskBody[TCtx any] interface {
	Execute(ctx TCtx, complete CompleteFn)
}

type ComparableTaskBody[TCtx any] interface {
	TaskBody[TCtx]
	Compare(other TaskBody[TCtx]) bool
}

type Task[TCtx any] struct {
	body        TaskBody[TCtx]
	isCompleted bool
	waiters     []chan any
	onCompleted []func(task *Task[TCtx])
	waiterLock  sync.Mutex
	err         error
}

func (t *Task[TCtx]) Body() TaskBody[TCtx] {
	return t.body
}

func (t *Task[TCtx]) IsCompleted() bool {
	return t.isCompleted
}

func (t *Task[TCtx]) Error() error {
	return t.err
}

func (t *Task[TCtx]) Wait() {
	t.waiterLock.Lock()
	if t.isCompleted {
		t.waiterLock.Unlock()
		return
	}

	waiter := make(chan any)
	t.waiters = append(t.waiters, waiter)
	t.waiterLock.Unlock()

	<-waiter
}

func (t *Task[TCtx]) OnCompleted(callback func(task *Task[TCtx])) {
	t.waiterLock.Lock()
	if t.isCompleted {
		t.waiterLock.Unlock()
		callback(t)
		return
	}

	t.onCompleted = append(t.onCompleted, callback)
	t.waiterLock.Unlock()
}
