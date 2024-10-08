package exec

import (
	"context"
	"fmt"
	"io"
	"sync"

	"gitlink.org.cn/cloudream/common/pkgs/future"
	"gitlink.org.cn/cloudream/common/utils/io2"
)

type Driver struct {
	planID     PlanID
	planBlder  *PlanBuilder
	callback   *future.SetValueFuture[map[string]any]
	ctx        context.Context
	cancel     context.CancelFunc
	driverExec *Executor
}

// 开始写入一个流。此函数会将输入视为一个完整的流，因此会给流包装一个Range来获取只需要的部分。
func (e *Driver) BeginWrite(str io.ReadCloser, handle *DriverWriteStream) {
	handle.Var.Stream = io2.NewRange(str, handle.RangeHint.Offset, handle.RangeHint.Length)
	e.driverExec.PutVars(handle.Var)
}

// 开始写入一个流。此函数默认输入流已经是Handle的RangeHint锁描述的范围，因此不会做任何其他处理
func (e *Driver) BeginWriteRanged(str io.ReadCloser, handle *DriverWriteStream) {
	handle.Var.Stream = str
	e.driverExec.PutVars(handle.Var)
}

func (e *Driver) BeginRead(handle *DriverReadStream) (io.ReadCloser, error) {
	err := e.driverExec.BindVars(e.ctx, handle.Var)
	if err != nil {
		return nil, fmt.Errorf("bind vars: %w", err)
	}

	return handle.Var.Stream, nil
}

func (e *Driver) Signal(signal *DriverSignalVar) {
	e.driverExec.PutVars(signal.Var)
}

func (e *Driver) Wait(ctx context.Context) (map[string]any, error) {
	stored, err := e.callback.Wait(ctx)
	if err != nil {
		return nil, err
	}

	return stored, nil
}

func (e *Driver) execute() {
	wg := sync.WaitGroup{}

	for _, p := range e.planBlder.WorkerPlans {
		wg.Add(1)

		go func(p *WorkerPlanBuilder) {
			defer wg.Done()

			plan := Plan{
				ID:  e.planID,
				Ops: p.Ops,
			}

			cli, err := p.Worker.NewClient()
			if err != nil {
				e.stopWith(fmt.Errorf("new client to worker %v: %w", p.Worker, err))
				return
			}
			defer cli.Close()

			err = cli.ExecutePlan(e.ctx, plan)
			if err != nil {
				e.stopWith(fmt.Errorf("execute plan at worker %v: %w", p.Worker, err))
				return
			}
		}(p)
	}

	stored, err := e.driverExec.Run(e.ctx)
	if err != nil {
		e.stopWith(fmt.Errorf("run executor switch: %w", err))
		return
	}

	wg.Wait()

	e.callback.SetValue(stored)
}

func (e *Driver) stopWith(err error) {
	e.callback.SetError(err)
	e.cancel()
}

type DriverWriteStream struct {
	Var       *StreamVar
	RangeHint *Range
}

type DriverReadStream struct {
	Var *StreamVar
}

type DriverSignalVar struct {
	Var *SignalVar
}
