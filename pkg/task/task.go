package task

import "sync"

type TaskBody[TCtx any] interface {
	Execute(ctx TCtx, complete func(completing func()))
}

type Task[TCtx any] struct {
	body        TaskBody[TCtx]
	isCompleted bool
	waiters     []chan any
	waiterLock  sync.Mutex
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
