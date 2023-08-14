package future

import (
	"time"
)

type SetVoidFuture struct {
	err          error
	isCompleted  bool
	completeChan chan any
}

func NewSetVoid() *SetVoidFuture {
	return &SetVoidFuture{
		completeChan: make(chan any),
	}
}

func (f *SetVoidFuture) SetVoid() {
	f.isCompleted = true
	close(f.completeChan)
}

func (f *SetVoidFuture) SetError(err error) {
	f.err = err
	f.isCompleted = true
	close(f.completeChan)
}

func (f *SetVoidFuture) Error() error {
	return f.err
}

func (f *SetVoidFuture) IsComplete() bool {
	return f.isCompleted
}

func (f *SetVoidFuture) Wait() error {
	<-f.completeChan
	return f.err
}

func (f *SetVoidFuture) WaitTimeout(timeout time.Duration) error {
	select {
	case <-f.completeChan:
		return f.err

	case <-time.After(timeout):
		return ErrWaitTimeout
	}
}