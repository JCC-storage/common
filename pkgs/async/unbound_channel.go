package async

import (
	"container/list"
	"errors"
	"gitlink.org.cn/cloudream/common/pkgs/future"
	"sync"
)

var ErrChannelClosed = errors.New("channel is closed")

type UnboundChannel[T any] struct {
	values  *list.List
	waiters []*future.SetValueFuture[T]
	lock    sync.Mutex
	err     error
}

func NewUnboundChannel[T any]() *UnboundChannel[T] {
	return &UnboundChannel[T]{
		values: list.New(),
	}
}

func (c *UnboundChannel[T]) Error() error {
	return c.err
}

func (c *UnboundChannel[T]) Send(val T) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.err != nil {
		return c.err
	}

	c.values.PushBack(val)

	for len(c.waiters) > 0 && c.values.Len() > 0 {
		waiter := c.waiters[0]
		waiter.SetValue(c.values.Front().Value.(T))
		c.values.Remove(c.values.Front())
		c.waiters = c.waiters[1:]
		return nil
	}

	return nil
}

func (c *UnboundChannel[T]) Receive() future.Future1[T] {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.err != nil {
		return future.NewReadyError1[T](c.err)
	}

	if c.values.Len() > 0 {
		ret := c.values.Front().Value.(T)
		c.values.Remove(c.values.Front())
		return future.NewReadyValue1[T](ret)
	}

	fut := future.NewSetValue[T]()
	c.waiters = append(c.waiters, fut)

	return fut
}

func (c *UnboundChannel[T]) Close() {
	c.CloseWithError(ErrChannelClosed)
}

func (c *UnboundChannel[T]) CloseWithError(err error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.err != nil {
		return
	}
	c.err = err

	for i := 0; i < len(c.waiters); i++ {
		c.waiters[i].SetError(c.err)
	}

	c.waiters = nil
	c.values = nil
}
