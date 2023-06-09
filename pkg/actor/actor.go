package actor

import "gitlink.org.cn/cloudream/common/pkg/future"

type CommandFn func()

type CommandChannel struct {
	cmdChan chan CommandFn
}

func NewCommandChannel() *CommandChannel {
	return &CommandChannel{
		cmdChan: make(chan CommandFn),
	}
}

func (c *CommandChannel) Send(cmd CommandFn) {
	c.cmdChan <- cmd
}

func (c *CommandChannel) Receive() (CommandFn, bool) {
	cmd, ok := <-c.cmdChan
	return cmd, ok
}

func (c *CommandChannel) ChanReceive() <-chan CommandFn {
	return c.cmdChan
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
