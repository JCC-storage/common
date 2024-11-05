package future

import (
	"context"
	"sync"
)

type SetValueFuture[T any] struct {
	isCompleted  bool
	ch           chan ChanValue1[T]
	completeOnce sync.Once
}

func NewSetValue[T any]() *SetValueFuture[T] {
	return &SetValueFuture[T]{
		ch: make(chan ChanValue1[T], 1),
	}
}

func (f *SetValueFuture[T]) SetComplete(val T, err error) {
	f.completeOnce.Do(func() {
		f.ch <- ChanValue1[T]{
			Err:   err,
			Value: val,
		}
		close(f.ch)
		f.isCompleted = true
	})
}

func (f *SetValueFuture[T]) SetValue(val T) {
	f.completeOnce.Do(func() {
		f.ch <- ChanValue1[T]{
			Value: val,
		}
		close(f.ch)
		f.isCompleted = true
	})
}

func (f *SetValueFuture[T]) SetError(err error) {
	f.completeOnce.Do(func() {
		f.ch <- ChanValue1[T]{
			Err: err,
		}
		close(f.ch)
		f.isCompleted = true
	})
}

func (f *SetValueFuture[T]) IsComplete() bool {
	return f.isCompleted
}

func (f *SetValueFuture[T]) Chan() <-chan ChanValue1[T] {
	return f.ch
}

// 等待直到Complete或者ctx被取消。
// 注：返回context.Canceled不代表产生结果的过程没有执行过，甚至不代表Future没有Complete
func (f *SetValueFuture[T]) Wait(ctx context.Context) (T, error) {
	select {
	case cv, ok := <-f.ch:
		if !ok {
			var ret T
			return ret, cv.Err
		}
		return cv.Value, cv.Err

	case <-ctx.Done():
		var ret T
		return ret, context.Canceled
	}
}

type SetValueFuture2[T1 any, T2 any] struct {
	isCompleted  bool
	ch           chan ChanValue2[T1, T2]
	completeOnce sync.Once
}

func NewSetValue2[T1 any, T2 any]() *SetValueFuture2[T1, T2] {
	return &SetValueFuture2[T1, T2]{
		ch: make(chan ChanValue2[T1, T2], 1),
	}
}

func (f *SetValueFuture2[T1, T2]) SetComplete(val1 T1, val2 T2, err error) {
	f.completeOnce.Do(func() {
		f.ch <- ChanValue2[T1, T2]{
			Value1: val1,
			Value2: val2,
			Err:    err,
		}
		close(f.ch)
		f.isCompleted = true
	})
}

func (f *SetValueFuture2[T1, T2]) SetValue(val1 T1, val2 T2) {
	f.completeOnce.Do(func() {
		f.ch <- ChanValue2[T1, T2]{
			Value1: val1,
			Value2: val2,
		}
		close(f.ch)
		f.isCompleted = true
	})
}

func (f *SetValueFuture2[T1, T2]) SetError(err error) {
	f.completeOnce.Do(func() {
		f.ch <- ChanValue2[T1, T2]{
			Err: err,
		}
		close(f.ch)
		f.isCompleted = true
	})
}

func (f *SetValueFuture2[T1, T2]) IsComplete() bool {
	return f.isCompleted
}

func (f *SetValueFuture2[T1, T2]) Wait(ctx context.Context) (T1, T2, error) {
	select {
	case cv, ok := <-f.ch:
		if !ok {
			return cv.Value1, cv.Value2, cv.Err
		}
		return cv.Value1, cv.Value2, cv.Err

	case <-ctx.Done():
		var ret1 T1
		var ret2 T2
		return ret1, ret2, context.Canceled
	}
}

func (f *SetValueFuture2[T1, T2]) Chan() <-chan ChanValue2[T1, T2] {
	return f.ch
}
