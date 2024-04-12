package event

import (
	"sync"

	"github.com/zyedidia/generic/list"
	"gitlink.org.cn/cloudream/common/utils/sync2"
)

type ExecuteOption struct {
	IsEmergency bool
	DontMerge   bool
}

type ExecuteContext[TArgs any] struct {
	Executor *Executor[TArgs]
	Option   ExecuteOption
	Args     TArgs
}

type postedEvent[TArgs any] struct {
	Event  Event[TArgs]
	Option ExecuteOption
}

type Executor[TArgs any] struct {
	events    *list.List[postedEvent[TArgs]]
	locker    sync.Mutex
	eventCond *sync2.CounterCond
	execArgs  TArgs
}

func NewExecutor[TArgs any](args TArgs) Executor[TArgs] {
	return Executor[TArgs]{
		events:    list.New[postedEvent[TArgs]](),
		locker:    sync.Mutex{},
		eventCond: sync2.NewCounterCond(0),
		execArgs:  args,
	}

}

func (e *Executor[TArgs]) Post(event Event[TArgs], opts ...ExecuteOption) {

	opt := ExecuteOption{
		IsEmergency: false,
		DontMerge:   false,
	}

	if len(opts) > 0 {
		opt = opts[0]
	}

	e.locker.Lock()
	defer e.locker.Unlock()

	// 紧急任务直接插入到队头，不进行合并

	if opt.IsEmergency {
		e.events.PushFront(postedEvent[TArgs]{
			Event:  event,
			Option: opt,
		})
		e.eventCond.Release()
		return
	}

	// 合并任务

	if opt.DontMerge {
		ptr := e.events.Front
		for ptr != nil {
			// 只与非紧急任务，且允许合并的任务进行合并
			if !ptr.Value.Option.IsEmergency && !ptr.Value.Option.DontMerge {
				if ptr.Value.Event.TryMerge(event) {
					return
				}
			}
			ptr = ptr.Next
		}
	}

	e.events.PushBack(postedEvent[TArgs]{
		Event:  event,
		Option: opt,
	})

	e.eventCond.Release()
}

// Execute 开始执行任务

func (e *Executor[TArgs]) Execute() error {

	for {
		e.eventCond.Wait()
		event := e.popFrontEvent()
		if event == nil {
			continue
		}

		ctx := ExecuteContext[TArgs]{
			Executor: e,
			Option:   event.Option,
			Args:     e.execArgs,
		}

		event.Event.Execute(ctx)
	}

}

func (e *Executor[TArgs]) popFrontEvent() *postedEvent[TArgs] {
	e.locker.Lock()
	defer e.locker.Unlock()

	if e.events.Front == nil {
		return nil
	}

	val := &e.events.Front.Value
	e.events.Remove(e.events.Front)
	return val

}
