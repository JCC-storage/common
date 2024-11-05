package exec

import (
	"context"
	"fmt"
	"sync"

	"github.com/hashicorp/go-multierror"
	"gitlink.org.cn/cloudream/common/pkgs/future"
	"gitlink.org.cn/cloudream/common/utils/lo2"
)

type binding struct {
	ID       VarID
	Callback *future.SetValueFuture[VarValue]
}

type freeVar struct {
	ID    VarID
	Value VarValue
}

type Executor struct {
	plan     Plan
	vars     map[VarID]freeVar
	bindings []*binding
	lock     sync.Mutex
	store    map[string]VarValue
}

func NewExecutor(plan Plan) *Executor {
	planning := Executor{
		plan:  plan,
		vars:  make(map[VarID]freeVar),
		store: make(map[string]VarValue),
	}

	return &planning
}

func (s *Executor) Plan() *Plan {
	return &s.plan
}

func (s *Executor) Run(ctx *ExecContext) (map[string]VarValue, error) {
	c, cancel := context.WithCancel(ctx.Context)
	ctx.Context = c

	defer cancel()

	err := s.runOps(s.plan.Ops, ctx, cancel)
	if err != nil {
		return nil, err
	}

	return s.store, nil
}

func (s *Executor) runOps(ops []Op, ctx *ExecContext, cancel context.CancelFunc) error {
	lock := sync.Mutex{}
	var err error

	var wg sync.WaitGroup
	wg.Add(len(ops))
	for i, arg := range ops {
		go func(arg Op, index int) {
			defer wg.Done()

			if e := arg.Execute(ctx, s); e != nil {
				lock.Lock()
				// 尽量不记录 Canceled 错误，除非没有其他错误
				if err == nil {
					err = e
				} else if err == context.Canceled {
					err = e
				} else if e != context.Canceled {
					err = multierror.Append(err, e)
				}
				lock.Unlock()

				cancel()
			}
		}(arg, i)
	}
	wg.Wait()

	return err
}

func (s *Executor) BindVar(ctx context.Context, id VarID) (VarValue, error) {
	s.lock.Lock()

	gv, ok := s.vars[id]
	if ok {
		delete(s.vars, id)
		s.lock.Unlock()
		return gv.Value, nil
	}

	callback := future.NewSetValue[VarValue]()
	s.bindings = append(s.bindings, &binding{
		ID:       id,
		Callback: callback,
	})

	s.lock.Unlock()
	return callback.Wait(ctx)
}

func (s *Executor) PutVar(id VarID, value VarValue) *Executor {
	s.lock.Lock()
	defer s.lock.Unlock()

	for ib, b := range s.bindings {
		if b.ID != id {
			continue
		}

		b.Callback.SetValue(value)
		s.bindings = lo2.RemoveAt(s.bindings, ib)

		return s
	}

	// 如果没有绑定，则直接放入变量表中
	s.vars[id] = freeVar{ID: id, Value: value}
	return s
}

func (s *Executor) Store(key string, val VarValue) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.store[key] = val
}

func BindVar[T VarValue](e *Executor, ctx context.Context, id VarID) (T, error) {
	v, err := e.BindVar(ctx, id)
	if err != nil {
		var def T
		return def, err
	}

	ret, ok := v.(T)
	if !ok {
		var def T
		return def, fmt.Errorf("binded var %v is %T, not %T", id, v, def)
	}

	return ret, nil
}

func BindArray[T VarValue](e *Executor, ctx context.Context, ids []VarID) ([]T, error) {
	ret := make([]T, len(ids))
	for i := range ids {
		v, err := e.BindVar(ctx, ids[i])
		if err != nil {
			return nil, err
		}

		v2, ok := v.(T)
		if !ok {
			var def T
			return nil, fmt.Errorf("binded var %v is %T, not %T", ids[i], v, def)
		}

		ret[i] = v2
	}
	return ret, nil
}

func PutArray[T VarValue](e *Executor, ids []VarID, values []T) {
	for i := range ids {
		e.PutVar(ids[i], values[i])
	}
}
