package sync2

import "context"

type SafeChannel[T any] struct {
	ch     chan T
	done   context.Context
	cacnel func()
}

func NewSafeChannel[T any]() SafeChannel[T] {
	ctx, cancel := context.WithCancel(context.Background())
	return SafeChannel[T]{
		ch:     make(chan T),
		done:   ctx,
		cacnel: cancel,
	}
}

func NewChannelWithCapacity[T any](cap int) SafeChannel[T] {
	ctx, cancel := context.WithCancel(context.Background())
	return SafeChannel[T]{
		ch:     make(chan T, cap),
		done:   ctx,
		cacnel: cancel,
	}
}

func (c *SafeChannel[T]) Send(val T) bool {
	select {
	case <-c.done.Done():
		return false
	case c.ch <- val:
		return true
	}
}

func (c *SafeChannel[T]) Receive() (T, bool) {
	select {
	case <-c.done.Done():
		var ret T
		return ret, false

	case v := <-c.ch:
		return v, true
	}
}

// 需要与Closed函数一起使用
func (c *SafeChannel[T]) Sender() chan<- T {
	return c.ch
}

// 需要与Closed函数一起使用
func (c *SafeChannel[T]) Receiver() <-chan T {
	return c.ch
}

// 如果返回的chan被关闭，则代表此SafeChannel已经关闭
func (c *SafeChannel[T]) Closed() <-chan struct{} {
	return c.done.Done()
}

func (c *SafeChannel[T]) Close() {
	c.cacnel()
}
