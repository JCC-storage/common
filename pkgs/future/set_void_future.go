package future

import (
	"context"
	"sync"
)

type SetVoidFuture struct {
	isCompleted  bool
	ch           chan error
	completeOnce sync.Once
}

func NewSetVoid() *SetVoidFuture {
	return &SetVoidFuture{
		ch: make(chan error, 1),
	}
}

func (f *SetVoidFuture) SetVoid() {
	f.completeOnce.Do(func() {
		f.ch <- nil
		close(f.ch)
		f.isCompleted = true
	})
}

func (f *SetVoidFuture) SetError(err error) {
	f.completeOnce.Do(func() {
		f.ch <- err
		close(f.ch)
		f.isCompleted = true
	})
}

func (f *SetVoidFuture) IsComplete() bool {
	return f.isCompleted
}

func (f *SetVoidFuture) Wait(ctx context.Context) error {
	select {
	case v, ok := <-f.ch:
		if !ok {
			return ErrCompleted
		}
		return v

	case <-ctx.Done():
		return context.Canceled
	}
}

func (f *SetVoidFuture) Chan() <-chan error {
	return f.ch
}
