package exec

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/hashicorp/go-multierror"
	"gitlink.org.cn/cloudream/common/pkgs/future"
	"gitlink.org.cn/cloudream/common/utils/io2"
	"gitlink.org.cn/cloudream/common/utils/math2"
)

type Driver struct {
	planID     PlanID
	planBlder  *PlanBuilder
	callback   *future.SetValueFuture[map[string]VarValue]
	ctx        *ExecContext
	cancel     context.CancelFunc
	driverExec *Executor
}

// 开始写入一个流。此函数会将输入视为一个完整的流，因此会给流包装一个Range来获取只需要的部分。
func (e *Driver) BeginWrite(str io.ReadCloser, handle *DriverWriteStream) {
	e.driverExec.PutVar(handle.ID, &StreamValue{Stream: io2.NewRange(str, handle.RangeHint.Offset, handle.RangeHint.Length)})
}

// 开始写入一个流。此函数默认输入流已经是Handle的RangeHint锁描述的范围，因此不会做任何其他处理
func (e *Driver) BeginWriteRanged(str io.ReadCloser, handle *DriverWriteStream) {
	e.driverExec.PutVar(handle.ID, &StreamValue{Stream: str})
}

func (e *Driver) BeginRead(handle *DriverReadStream) (io.ReadCloser, error) {
	str, err := BindVar[*StreamValue](e.driverExec, e.ctx.Context, handle.ID)
	if err != nil {
		return nil, fmt.Errorf("bind vars: %w", err)
	}

	return str.Stream, nil
}

func (e *Driver) Signal(signal *DriverSignalVar) {
	e.driverExec.PutVar(signal.ID, &SignalValue{})
}

func (e *Driver) Wait(ctx context.Context) (map[string]VarValue, error) {
	stored, err := e.callback.Wait(ctx)
	if err != nil {
		return nil, err
	}

	return stored, nil
}

func (e *Driver) execute() {
	wg := sync.WaitGroup{}

	errLock := sync.Mutex{}
	var execErr error
	for _, p := range e.planBlder.WorkerPlans {
		wg.Add(1)

		go func(p *WorkerPlanBuilder, ctx context.Context, cancel context.CancelFunc) {
			defer wg.Done()

			plan := Plan{
				ID:  e.planID,
				Ops: p.Ops,
			}

			cli, err := p.Worker.NewClient()
			if err != nil {
				errLock.Lock()
				execErr = multierror.Append(execErr, fmt.Errorf("worker %v: new client: %w", p.Worker, err))
				errLock.Unlock()
				cancel()
				return
			}
			defer cli.Close()

			err = cli.ExecutePlan(ctx, plan)
			if err != nil {
				errLock.Lock()
				execErr = multierror.Append(execErr, fmt.Errorf("worker %v: execute plan: %w", p.Worker, err))
				errLock.Unlock()
				cancel()
				return
			}
		}(p, e.ctx.Context, e.cancel)
	}

	stored, err := e.driverExec.Run(e.ctx)
	if err != nil {
		errLock.Lock()
		execErr = multierror.Append(execErr, fmt.Errorf("driver: execute plan: %w", err))
		errLock.Unlock()
		e.cancel()
	}

	wg.Wait()

	e.callback.SetComplete(stored, execErr)
}

type DriverWriteStream struct {
	ID        VarID
	RangeHint *math2.Range
}

type DriverReadStream struct {
	ID VarID
}

type DriverSignalVar struct {
	ID     VarID
	Signal SignalValue
}
