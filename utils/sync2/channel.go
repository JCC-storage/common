package sync2

import (
	"context"
	"errors"
	"sync"
)

var ErrChannelClosed = errors.New("channel is closed")

type Channel[T any] struct {
	ch        chan T
	closed    chan any
	closeOnce sync.Once
	err       error
}

func NewChannel[T any]() *Channel[T] {
	return &Channel[T]{
		ch:     make(chan T),
		closed: make(chan any),
	}
}

func (c *Channel[T]) Error() error {
	return c.err
}

func (c *Channel[T]) Send(val T) error {
	select {
	case c.ch <- val:
		return nil
	case <-c.closed:
		return c.err
	}
}

func (c *Channel[T]) Receive(ctx context.Context) (T, error) {
	select {
	case val := <-c.ch:
		return val, nil
	case <-c.closed:
		var t T
		return t, c.err
	case <-ctx.Done():
		var t T
		return t, ctx.Err()
	}
}

// 获取channel的发送端，需要与Closed一起使用，防止错过关闭信号
func (c *Channel[T]) Sender() chan<- T {
	return c.ch
}

// 获取channel的接收端，需要与Closed一起使用，防止错过关闭信号
func (c *Channel[T]) Receiver() <-chan T {
	return c.ch
}

// 获取channel的关闭信号，用于通知接收端和发送端关闭
func (c *Channel[T]) Closed() <-chan any {
	return c.closed
}

// 关闭channel。注：此操作不会关闭Sender和Receiver返回的channel
func (c *Channel[T]) Close() {
	c.closeOnce.Do(func() {
		close(c.closed)
		c.err = ErrChannelClosed
	})
}

// 关闭channel并设置error。注：此操作不会关闭Sender和Receiver返回的channel
func (c *Channel[T]) CloseWithError(err error) {
	c.closeOnce.Do(func() {
		close(c.closed)
		c.err = err
	})
}
