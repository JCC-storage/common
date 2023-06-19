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

type SetValueFuture2[T1 any, T2 any] struct {
	value1       T1
	value2       T2
	err          error
	isCompleted  bool
	completeChan chan any
}

func NewSetValue2[T1 any, T2 any]() *SetValueFuture2[T1, T2] {
	return &SetValueFuture2[T1, T2]{
		completeChan: make(chan any),
	}
}

func (f *SetValueFuture2[T1, T2]) SetComplete(val1 T1, val2 T2, err error) {
	f.value1 = val1
	f.value2 = val2
	f.err = err
	f.isCompleted = true
	close(f.completeChan)
}

func (f *SetValueFuture2[T1, T2]) SetValue(val1 T1, val2 T2) {
	f.value1 = val1
	f.value2 = val2
	f.isCompleted = true
	close(f.completeChan)
}

func (f *SetValueFuture2[T1, T2]) SetError(err error) {
	f.err = err
	f.isCompleted = true
	close(f.completeChan)
}

func (f *SetValueFuture2[T1, T2]) Error() error {
	return f.err
}

func (f *SetValueFuture2[T1, T2]) Value() (T1, T2) {
	return f.value1, f.value2
}

func (f *SetValueFuture2[T1, T2]) IsComplete() bool {
	return f.isCompleted
}

func (f *SetValueFuture2[T1, T2]) Wait() error {
	<-f.completeChan
	return f.err
}

func (f *SetValueFuture2[T1, T2]) WaitTimeout(timeout time.Duration) error {
	select {
	case <-f.completeChan:
		return f.err

	case <-time.After(timeout):
		return ErrWaitTimeout
	}
}

func (f *SetValueFuture2[T1, T2]) WaitValue() (T1, T2, error) {
	<-f.completeChan
	return f.value1, f.value2, f.err
}

func (f *SetValueFuture2[T1, T2]) WaitValueTimeout(timeout time.Duration) (T1, T2, error) {
	select {
	case <-f.completeChan:
		return f.value1, f.value2, f.err

	case <-time.After(timeout):
		var ret1 T1
		var ret2 T2
		return ret1, ret2, ErrWaitTimeout
	}
}
