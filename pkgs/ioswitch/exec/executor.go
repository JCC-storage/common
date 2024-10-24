package exec

import (
	"context"
	"fmt"
	"sync"

	"gitlink.org.cn/cloudream/common/pkgs/future"
	"gitlink.org.cn/cloudream/common/utils/lo2"
	"gitlink.org.cn/cloudream/common/utils/sync2"
)

type bindingVars struct {
	Waittings []Var
	Bindeds   []Var
	Callback  *future.SetVoidFuture
}

type Executor struct {
	plan     Plan
	vars     map[VarID]Var
	bindings []*bindingVars
	lock     sync.Mutex
	store    map[string]any
}

func NewExecutor(plan Plan) *Executor {
	planning := Executor{
		plan:  plan,
		vars:  make(map[VarID]Var),
		store: make(map[string]any),
	}

	return &planning
}

func (s *Executor) Plan() *Plan {
	return &s.plan
}

func (s *Executor) Run(ctx *ExecContext) (map[string]any, error) {
	c, cancel := context.WithCancel(ctx.Context)
	ctx.Context = c

	defer cancel()

	err := sync2.ParallelDo(s.plan.Ops, func(o Op, idx int) error {
		err := o.Execute(ctx, s)

		s.lock.Lock()
		defer s.lock.Unlock()

		if err != nil {
			cancel()
			return fmt.Errorf("%T: %w", o, err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return s.store, nil
}

func (s *Executor) BindVars(ctx context.Context, vs ...Var) error {
	s.lock.Lock()

	callback := future.NewSetVoid()
	binding := &bindingVars{
		Callback: callback,
	}

	for _, v := range vs {
		v2 := s.vars[v.GetID()]
		if v2 == nil {
			binding.Waittings = append(binding.Waittings, v)
			continue
		}

		if err := AssignVar(v2, v); err != nil {
			s.lock.Unlock()
			return fmt.Errorf("assign var %v to %v: %w", v2.GetID(), v.GetID(), err)
		}

		binding.Bindeds = append(binding.Bindeds, v)
	}

	if len(binding.Waittings) == 0 {
		s.lock.Unlock()
		return nil
	}

	s.bindings = append(s.bindings, binding)
	s.lock.Unlock()

	err := callback.Wait(ctx)

	s.lock.Lock()
	defer s.lock.Unlock()

	s.bindings = lo2.Remove(s.bindings, binding)

	return err
}

func (s *Executor) PutVars(vs ...Var) {
	s.lock.Lock()
	defer s.lock.Unlock()

loop:
	for _, v := range vs {
		for ib, b := range s.bindings {
			for iw, w := range b.Waittings {
				if w.GetID() != v.GetID() {
					continue
				}

				if err := AssignVar(v, w); err != nil {
					b.Callback.SetError(fmt.Errorf("assign var %v to %v: %w", v.GetID(), w.GetID(), err))
					// 绑定类型不对，说明生成的执行计划有问题，怎么处理都可以，因为最终会执行失败
					continue loop
				}

				b.Bindeds = append(b.Bindeds, w)
				b.Waittings = lo2.RemoveAt(b.Waittings, iw)
				if len(b.Waittings) == 0 {
					b.Callback.SetVoid()
					s.bindings = lo2.RemoveAt(s.bindings, ib)
				}

				// 绑定成功，继续最外层循环
				continue loop
			}

		}

		// 如果没有绑定，则直接放入变量表中
		s.vars[v.GetID()] = v
	}
}

func (s *Executor) Store(key string, val any) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.store[key] = val
}

func BindArrayVars[T Var](sw *Executor, ctx context.Context, vs []T) error {
	var vs2 []Var
	for _, v := range vs {
		vs2 = append(vs2, v)
	}

	return sw.BindVars(ctx, vs2...)
}

func PutArrayVars[T Var](sw *Executor, vs []T) {
	var vs2 []Var
	for _, v := range vs {
		vs2 = append(vs2, v)
	}

	sw.PutVars(vs2...)
}
