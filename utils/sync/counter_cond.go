package sync

import "sync"

type CounterCond struct {
	count int
	cond  *sync.Cond
}

func NewCounterCond(initCount int) *CounterCond {
	return &CounterCond{
		count: initCount,
		cond:  sync.NewCond(&sync.Mutex{}),
	}
}

func (c *CounterCond) Wait() {
	c.cond.L.Lock()

	for c.count == 0 {
		c.cond.Wait()
	}

	c.count--

	c.cond.L.Unlock()
}

func (c *CounterCond) Release() {
	c.count++
	c.cond.Signal()
}
