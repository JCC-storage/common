package future

import (
	"time"
)

type SetValueFuture[T any] struct {
	value        T
	err          error
	isCompleted  bool
	completeChan chan any
}

func NewSetValue[T any]() *SetValueFuture[T] {
	return &SetValueFuture[T]{
		completeChan: make(chan any),
	}
}

func (f *SetValueFuture[T]) SetComplete(val T, err error) {
	f.value = val
	f.err = err
	f.isCompleted = true
	close(f.completeChan)
}

func (f *SetValueFuture[T]) SetValue(val T) {
	f.value = val
	f.isCompleted = true
	close(f.completeChan)
}

func (f *SetValueFuture[T]) SetError(err error) {
	f.err = err
	f.isCompleted = true
	close(f.completeChan)
}

func (f *SetValueFuture[T]) Error() error {
	return f.err
}

func (f *SetValueFuture[T]) Value() T {
	return f.value
}

func (f *SetValueFuture[T]) IsComplete() bool {
	return f.isCompleted
}

func (f *SetValueFuture[T]) Wait() error {
	<-f.completeChan
	return f.err
}

func (f *SetValueFuture[T]) WaitTimeout(timeout time.Duration) error {
	select {
	case <-f.completeChan:
		return f.err

	case <-time.After(timeout):
		return ErrWaitTimeout
	}
}

func (f *SetValueFuture[T]) WaitValue() (T, error) {
	<-f.completeChan
	return f.value, f.err
}

func (f *SetValueFuture[T]) WaitValueTimeout(timeout time.Duration) (T, error) {
	select {
	case <-f.completeChan:
		return f.value, f.err

	case <-time.After(timeout):
		var ret T
		return ret, ErrWaitTimeout
	}
}
