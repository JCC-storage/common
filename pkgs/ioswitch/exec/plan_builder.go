package exec

import (
	"context"
	"strings"

	"gitlink.org.cn/cloudream/common/pkgs/future"
	"gitlink.org.cn/cloudream/common/utils/lo2"
)

type PlanBuilder struct {
	Vars        []Var
	WorkerPlans []*WorkerPlanBuilder
	DriverPlan  DriverPlanBuilder
}

func NewPlanBuilder() *PlanBuilder {
	bld := &PlanBuilder{
		DriverPlan: DriverPlanBuilder{},
	}

	return bld
}

func (b *PlanBuilder) AtDriver() *DriverPlanBuilder {
	return &b.DriverPlan
}

func (b *PlanBuilder) AtWorker(worker WorkerInfo) *WorkerPlanBuilder {
	for _, p := range b.WorkerPlans {
		if p.Worker.Equals(worker) {
			return p
		}
	}

	p := &WorkerPlanBuilder{
		Worker: worker,
	}
	b.WorkerPlans = append(b.WorkerPlans, p)

	return p
}

func (b *PlanBuilder) NewStreamVar() *StreamVar {
	v := &StreamVar{
		ID: VarID(len(b.Vars)),
	}
	b.Vars = append(b.Vars, v)

	return v
}

func (b *PlanBuilder) NewIntVar() *IntVar {
	v := &IntVar{
		ID: VarID(len(b.Vars)),
	}
	b.Vars = append(b.Vars, v)

	return v
}

func (b *PlanBuilder) NewStringVar() *StringVar {
	v := &StringVar{
		ID: VarID(len(b.Vars)),
	}
	b.Vars = append(b.Vars, v)

	return v
}
func (b *PlanBuilder) NewSignalVar() *SignalVar {
	v := &SignalVar{
		ID: VarID(len(b.Vars)),
	}
	b.Vars = append(b.Vars, v)

	return v
}

func (b *PlanBuilder) Execute(ctx *ExecContext) *Driver {
	c, cancel := context.WithCancel(ctx.Context)
	ctx.Context = c

	planID := genRandomPlanID()

	execPlan := Plan{
		ID:  planID,
		Ops: b.DriverPlan.Ops,
	}

	exec := Driver{
		planID:     planID,
		planBlder:  b,
		callback:   future.NewSetValue[map[string]any](),
		ctx:        ctx,
		cancel:     cancel,
		driverExec: NewExecutor(execPlan),
	}
	go exec.execute()

	return &exec
}

func (b *PlanBuilder) String() string {
	sb := strings.Builder{}
	sb.WriteString("Driver:\n")
	for _, op := range b.DriverPlan.Ops {
		sb.WriteString(op.String())
		sb.WriteRune('\n')
	}
	sb.WriteRune('\n')

	for _, w := range b.WorkerPlans {
		sb.WriteString("Worker(")
		sb.WriteString(w.Worker.String())
		sb.WriteString("):\n")
		for _, op := range w.Ops {
			sb.WriteString(op.String())
			sb.WriteRune('\n')
		}
		sb.WriteRune('\n')
	}

	return sb.String()
}

type WorkerPlanBuilder struct {
	Worker WorkerInfo
	Ops    []Op
}

func (b *WorkerPlanBuilder) AddOp(op Op) {
	b.Ops = append(b.Ops, op)
}

func (b *WorkerPlanBuilder) RemoveOp(op Op) {
	b.Ops = lo2.Remove(b.Ops, op)
}

type DriverPlanBuilder struct {
	Ops []Op
}

func (b *DriverPlanBuilder) AddOp(op Op) {
	b.Ops = append(b.Ops, op)
}

func (b *DriverPlanBuilder) RemoveOp(op Op) {
	b.Ops = lo2.Remove(b.Ops, op)
}
