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

func (c *CounterCond) Wait() bool {
	c.cond.L.Lock()
	defer c.cond.L.Unlock()

	for c.count == 0 {
		c.cond.Wait()

		if c.count == 0 {
			return false
		}
	}

	c.count--

	return true
}

func (c *CounterCond) Release() {
	c.count++
	c.cond.Signal()
}

// WakeupAll 不改变计数状态，唤醒所有等待线程
func (c *CounterCond) WakeupAll() {
	c.cond.Broadcast()
}
