package tickevent

import (
	"sync/atomic"
	"time"
)

type StartOption struct {
	RandomStartDelayMs int // 第一次任务启动前增加一段随机时长的延迟，随机的最长延迟时间由这个参数指定。如果为0，则不进行延迟。
}

type ExecuteContext[TArgs any] struct {
	Executor *Executor[TArgs]
	Self     *EventTicker[TArgs]
	Args     TArgs
}

type EventTicker[TArgs any] struct {
	event      TickEvent[TArgs]
	intervalMs int
	doneChan   chan int
	done       atomic.Bool
}

type Executor[TArgs any] struct {
	execArgs TArgs
}

func NewExecutor[TArgs any](args TArgs) Executor[TArgs] {
	return Executor[TArgs]{
		execArgs: args,
	}
}

func (e *Executor[TArgs]) Start(event TickEvent[TArgs], intervalMs int, opts ...StartOption) EventTicker[TArgs] {
	opt := StartOption{}
	if len(opts) > 0 {
		opt = opts[0]
	}

	ticker := EventTicker[TArgs]{
		event:      event,
		intervalMs: intervalMs,
		doneChan:   make(chan int),
		done:       atomic.Bool{},
	}
	ticker.done.Store(false)

	go func() {
		if opt.RandomStartDelayMs > 0 {
			<-time.After(time.Duration(opt.RandomStartDelayMs) * time.Millisecond)
		}

		timeTicker := time.NewTicker(time.Duration(intervalMs) * time.Millisecond)

	loop:
		for {
			select {
			case <-timeTicker.C:
				if ticker.done.Load() {
					break loop
				}

				execCtx := ExecuteContext[TArgs]{
					Executor: e,
					Self:     &ticker,
					Args:     e.execArgs,
				}
				event.Execute(execCtx)

			case <-ticker.doneChan:
				break loop
			}
		}

		timeTicker.Stop()
	}()

	return ticker
}

func (e *Executor[TArgs]) Stop(ticker EventTicker[TArgs]) {
	ticker.done.Store(true)
	close(ticker.doneChan)
	// 保证在调用此函数结束后，事件不会再被调用
}
