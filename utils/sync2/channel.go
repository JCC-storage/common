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

func (c *Channel[T]) Sender() chan<- T {
	return c.ch
}

func (c *Channel[T]) Receiver() <-chan T {
	return c.ch
}

func (c *Channel[T]) Close() {
	c.closeOnce.Do(func() {
		close(c.closed)
		close(c.ch)
		c.err = ErrChannelClosed
	})
}

func (c *Channel[T]) CloseWithError(err error) {
	c.closeOnce.Do(func() {
		close(c.closed)
		close(c.ch)
		c.err = err
	})
}

func (c *Channel[T]) Closed() <-chan any {
	return c.closed
}
