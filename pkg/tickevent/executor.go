package tickevent

import (
	"sync/atomic"
	"time"
)

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

func (e *Executor[TArgs]) Start(event TickEvent[TArgs], intervalMs int) EventTicker[TArgs] {
	ticker := EventTicker[TArgs]{
		event:      event,
		intervalMs: intervalMs,
		doneChan:   make(chan int),
		done:       atomic.Bool{},
	}
	ticker.done.Store(false)

	go func() {
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
