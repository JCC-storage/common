package future

import (
	"time"
)

type SetValueFuture[T any] struct {
	result       T
	err          error
	isCompleted  bool
	completeChan chan any
}

func NewSetValue[T any]() *SetValueFuture[T] {
	return &SetValueFuture[T]{
		completeChan: make(chan any),
	}
}

func (f *SetValueFuture[T]) SetValue(val T) {
	f.result = val
	f.isCompleted = true
	close(f.completeChan)
}

func (f *SetValueFuture[T]) SetError(err error) {
	f.err = err
	f.isCompleted = true
	close(f.completeChan)
}

func (f *SetValueFuture[T]) IsComplete() bool {
	return f.isCompleted
}

func (f *SetValueFuture[T]) Wait() (T, error) {
	<-f.completeChan
	return f.result, f.err
}

func (f *SetValueFuture[T]) WaitTimeout(timeout time.Duration) (T, error) {
	select {
	case <-f.completeChan:
		return f.result, f.err

	case <-time.After(timeout):
		var ret T
		return ret, ErrWaitTimeout
	}
}
