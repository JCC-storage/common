package future

import (
	"context"
	"sync"
)

type SetValueFuture[T any] struct {
	value        T
	err          error
	isCompleted  bool
	completeChan chan any
	completeOnce sync.Once
}

func NewSetValue[T any]() *SetValueFuture[T] {
	return &SetValueFuture[T]{
		completeChan: make(chan any),
	}
}

func (f *SetValueFuture[T]) SetComplete(val T, err error) {
	f.completeOnce.Do(func() {
		f.value = val
		f.err = err
		f.isCompleted = true
		close(f.completeChan)
	})
}

func (f *SetValueFuture[T]) SetValue(val T) {
	f.completeOnce.Do(func() {
		f.value = val
		f.isCompleted = true
		close(f.completeChan)
	})
}

func (f *SetValueFuture[T]) SetError(err error) {
	f.completeOnce.Do(func() {
		f.err = err
		f.isCompleted = true
		close(f.completeChan)
	})
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

// 等待直到Complete或者ctx被取消。
// 注：返回ErrContextCancelled不代表产生结果的过程没有执行过，甚至不代表Future没有Complete
func (f *SetValueFuture[T]) Wait(ctx context.Context) error {
	select {
	case <-f.completeChan:
		return f.err

	case <-ctx.Done():
		return ErrContextCancelled
	}
}

func (f *SetValueFuture[T]) WaitValue(ctx context.Context) (T, error) {
	select {
	case <-f.completeChan:
		return f.value, f.err

	case <-ctx.Done():
		var ret T
		return ret, ErrContextCancelled
	}
}

type SetValueFuture2[T1 any, T2 any] struct {
	value1       T1
	value2       T2
	err          error
	isCompleted  bool
	completeChan chan any
	completeOnce sync.Once
}

func NewSetValue2[T1 any, T2 any]() *SetValueFuture2[T1, T2] {
	return &SetValueFuture2[T1, T2]{
		completeChan: make(chan any),
	}
}

func (f *SetValueFuture2[T1, T2]) SetComplete(val1 T1, val2 T2, err error) {
	f.completeOnce.Do(func() {
		f.value1 = val1
		f.value2 = val2
		f.err = err
		f.isCompleted = true
		close(f.completeChan)
	})
}

func (f *SetValueFuture2[T1, T2]) SetValue(val1 T1, val2 T2) {
	f.completeOnce.Do(func() {
		f.value1 = val1
		f.value2 = val2
		f.isCompleted = true
		close(f.completeChan)
	})
}

func (f *SetValueFuture2[T1, T2]) SetError(err error) {
	f.completeOnce.Do(func() {
		f.err = err
		f.isCompleted = true
		close(f.completeChan)
	})
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

func (f *SetValueFuture2[T1, T2]) WaitValue(ctx context.Context) (T1, T2, error) {
	select {
	case <-f.completeChan:
		return f.value1, f.value2, f.err

	case <-ctx.Done():
		var ret1 T1
		var ret2 T2
		return ret1, ret2, ErrContextCancelled
	}
}
