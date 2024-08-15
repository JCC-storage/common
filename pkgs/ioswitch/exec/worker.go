package exec

import (
	"context"
	"sync"

	"github.com/samber/lo"
	"gitlink.org.cn/cloudream/common/pkgs/future"
	"gitlink.org.cn/cloudream/common/utils/lo2"
)

type finding struct {
	PlanID   PlanID
	Callback *future.SetValueFuture[*Executor]
}

type Worker struct {
	lock      sync.Mutex
	executors map[PlanID]*Executor
	findings  []*finding
}

func NewWorker() Worker {
	return Worker{
		executors: make(map[PlanID]*Executor),
	}
}

func (s *Worker) Add(exe *Executor) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.executors[exe.Plan().ID] = exe

	s.findings = lo.Reject(s.findings, func(f *finding, idx int) bool {
		if f.PlanID != exe.Plan().ID {
			return false
		}

		f.Callback.SetValue(exe)
		return true
	})
}

func (s *Worker) Remove(sw *Executor) {
	s.lock.Lock()
	defer s.lock.Unlock()

	delete(s.executors, sw.Plan().ID)
}

func (s *Worker) FindByID(id PlanID) *Executor {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.executors[id]
}

func (s *Worker) FindByIDContexted(ctx context.Context, id PlanID) *Executor {
	s.lock.Lock()

	sw := s.executors[id]
	if sw != nil {
		s.lock.Unlock()
		return sw
	}

	cb := future.NewSetValue[*Executor]()
	f := &finding{
		PlanID:   id,
		Callback: cb,
	}
	s.findings = append(s.findings, f)

	s.lock.Unlock()

	sw, _ = cb.WaitValue(ctx)

	s.lock.Lock()
	defer s.lock.Unlock()

	s.findings = lo2.Remove(s.findings, f)

	return sw
}
