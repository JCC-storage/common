package sync2

import (
	"container/list"
	"sync"
)

type UnboundChannel[T any] struct {
	values *list.List
	cond   *sync.Cond
	err    error
}

func NewUnboundChannel[T any]() *UnboundChannel[T] {
	return &UnboundChannel[T]{
		values: list.New(),
		cond:   sync.NewCond(&sync.Mutex{}),
	}
}

func (c *UnboundChannel[T]) Error() error {
	return c.err
}

func (c *UnboundChannel[T]) Send(val T) error {
	c.cond.L.Lock()
	if c.err != nil {
		c.cond.L.Unlock()
		return c.err
	}
	c.values.PushBack(val)
	c.cond.L.Unlock()

	c.cond.Signal()
	return nil
}

func (c *UnboundChannel[T]) Receive() (T, error) {
	c.cond.L.Lock()
	defer c.cond.L.Unlock()

	if c.values.Len() == 0 {
		c.cond.Wait()
	}

	if c.values.Len() == 0 {
		var ret T
		return ret, c.err
	}

	ret := c.values.Front().Value.(T)
	c.values.Remove(c.values.Front())
	return ret, nil
}

func (c *UnboundChannel[T]) Close() {
	c.cond.L.Lock()
	if c.err != nil {
		return
	}
	c.err = ErrChannelClosed
	c.cond.L.Unlock()
	c.cond.Broadcast()
}

func (c *UnboundChannel[T]) CloseWithError(err error) {
	c.cond.L.Lock()
	if c.err != nil {
		return
	}
	c.err = err
	c.cond.L.Unlock()
	c.cond.Broadcast()
}
