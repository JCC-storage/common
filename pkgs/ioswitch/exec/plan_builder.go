package exec

import (
	"context"
	"sync"

	"gitlink.org.cn/cloudream/common/pkgs/future"
	cdssdk "gitlink.org.cn/cloudream/common/sdks/storage"
	"gitlink.org.cn/cloudream/common/utils/lo2"
)

type PlanBuilder struct {
	Vars        []Var
	WorkerPlans map[cdssdk.NodeID]*WorkerPlanBuilder
	DriverPlan  DriverPlanBuilder
}

func NewPlanBuilder() *PlanBuilder {
	bld := &PlanBuilder{
		WorkerPlans: make(map[cdssdk.NodeID]*WorkerPlanBuilder),
		DriverPlan: DriverPlanBuilder{
			StoreMap: &sync.Map{},
		},
	}

	return bld
}

func (b *PlanBuilder) AtExecutor() *DriverPlanBuilder {
	return &b.DriverPlan
}

func (b *PlanBuilder) AtAgent(node cdssdk.Node) *WorkerPlanBuilder {
	agtPlan, ok := b.WorkerPlans[node.NodeID]
	if !ok {
		agtPlan = &WorkerPlanBuilder{
			Node: node,
		}
		b.WorkerPlans[node.NodeID] = agtPlan
	}

	return agtPlan
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

func (b *PlanBuilder) Execute() *Driver {
	ctx, cancel := context.WithCancel(context.Background())
	planID := genRandomPlanID()

	execPlan := Plan{
		ID:  planID,
		Ops: b.DriverPlan.Ops,
	}

	exec := Driver{
		planID:     planID,
		planBlder:  b,
		callback:   future.NewSetVoid(),
		ctx:        ctx,
		cancel:     cancel,
		driverExec: NewExecutor(execPlan),
	}
	go exec.execute()

	return &exec
}

type WorkerPlanBuilder struct {
	Node cdssdk.Node
	Ops  []Op
}

func (b *WorkerPlanBuilder) AddOp(op Op) {
	b.Ops = append(b.Ops, op)
}

func (b *WorkerPlanBuilder) RemoveOp(op Op) {
	b.Ops = lo2.Remove(b.Ops, op)
}

type DriverPlanBuilder struct {
	Ops      []Op
	StoreMap *sync.Map
}

func (b *DriverPlanBuilder) AddOp(op Op) {
	b.Ops = append(b.Ops, op)
}

func (b *DriverPlanBuilder) RemoveOp(op Op) {
	b.Ops = lo2.Remove(b.Ops, op)
}
