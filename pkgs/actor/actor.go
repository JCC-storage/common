package actor

import (
	"sync"
	"sync/atomic"

	"github.com/zyedidia/generic/list"

	"gitlink.org.cn/cloudream/common/pkgs/future"
	mysync "gitlink.org.cn/cloudream/common/utils/sync"
)

type CommandFn func()

type CommandChannel struct {
	cmds        *list.List[CommandFn]
	cmdsLock    sync.Mutex
	cmdsCounter *mysync.CounterCond

	chanReceive     chan CommandFn
	chanReceiveDone atomic.Bool
	chanReceiveLock sync.Mutex

	cmdChan chan CommandFn
}

func NewCommandChannel() *CommandChannel {
	return &CommandChannel{
		cmds:        list.New[CommandFn](),
		cmdsCounter: mysync.NewCounterCond(0),
		cmdChan:     make(chan CommandFn),
	}
}

func (c *CommandChannel) Send(cmd CommandFn) {
	c.cmdsLock.Lock()
	defer c.cmdsLock.Unlock()

	c.cmds.PushBack(cmd)
	c.cmdsCounter.Release()
}

func (c *CommandChannel) Receive() CommandFn {
	for !c.cmdsCounter.Wait() {
	}

	c.cmdsLock.Lock()
	defer c.cmdsLock.Unlock()

	val := c.cmds.Front.Value
	c.cmds.Remove(c.cmds.Front)

	return val
}

func (c *CommandChannel) BeginChanReceive() <-chan CommandFn {
	c.chanReceiveLock.Lock()
	defer c.chanReceiveLock.Unlock()

	if c.chanReceive == nil {
		// Stop的时候会设置c.chanReceive为nil，所以需要使用额外的变量
		chanRecv := make(chan CommandFn)
		c.chanReceive = chanRecv
		c.chanReceiveDone = atomic.Bool{}
		go func() {
			for {
				ok := c.cmdsCounter.Wait()
				if c.chanReceiveDone.Load() {
					// 已取消通过channel获取命令，那么如果被唤醒的时候
					// 消耗了队列中的一个命令，那么就要将它还回去。此处只要把计数还回去就行
					if ok {
						c.cmdsCounter.Release()
					}
					close(chanRecv)
					return
				}

				if !ok {
					continue
				}

				c.cmdsLock.Lock()
				val := c.cmds.Front.Value
				c.cmds.Remove(c.cmds.Front)
				c.cmdsLock.Unlock()

				chanRecv <- val
			}
		}()
	}

	return c.chanReceive
}

func (c *CommandChannel) CloseChanReceive() {
	c.chanReceiveLock.Lock()
	defer c.chanReceiveLock.Unlock()

	c.chanReceive = nil
	c.chanReceiveDone.Store(true)
	// 让chan receive线程能够退出
	c.cmdsCounter.WakeupAll()
}

func Wait(c *CommandChannel, cmd func() error) error {
	fut := future.NewSetVoid()

	c.Send(func() {
		err := cmd()
		if err != nil {
			fut.SetError(err)
		} else {
			fut.SetVoid()
		}
	})

	return fut.Wait()
}

func WaitValue[T any](c *CommandChannel, cmd func() (T, error)) (T, error) {
	fut := future.NewSetValue[T]()

	c.Send(func() {
		val, err := cmd()
		fut.SetComplete(val, err)
	})

	return fut.WaitValue()
}

func WaitValue2[T1 any, T2 any](c *CommandChannel, cmd func() (T1, T2, error)) (T1, T2, error) {
	fut := future.NewSetValue2[T1, T2]()

	c.Send(func() {
		val1, val2, err := cmd()
		fut.SetComplete(val1, val2, err)
	})

	return fut.WaitValue()
}
