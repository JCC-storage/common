package future

import (
	"context"
	"sync"
)

type SetVoidFuture struct {
	err          error
	isCompleted  bool
	completeChan chan any
	completeOnce sync.Once
}

func NewSetVoid() *SetVoidFuture {
	return &SetVoidFuture{
		completeChan: make(chan any),
	}
}

func (f *SetVoidFuture) SetVoid() {
	f.completeOnce.Do(func() {
		f.isCompleted = true
		close(f.completeChan)
	})
}

func (f *SetVoidFuture) SetError(err error) {
	f.completeOnce.Do(func() {
		f.err = err
		f.isCompleted = true
		close(f.completeChan)
	})
}

func (f *SetVoidFuture) Error() error {
	return f.err
}

func (f *SetVoidFuture) IsComplete() bool {
	return f.isCompleted
}

func (f *SetVoidFuture) Wait(ctx context.Context) error {
	select {
	case <-f.completeChan:
		return f.err

	case <-ctx.Done():
		return ErrContextCancelled
	}
}
