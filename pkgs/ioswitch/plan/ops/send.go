package ops

import (
	"context"
	"fmt"
	"io"

	"gitlink.org.cn/cloudream/common/pkgs/future"
	"gitlink.org.cn/cloudream/common/pkgs/ioswitch/dag"
	"gitlink.org.cn/cloudream/common/pkgs/ioswitch/exec"
	"gitlink.org.cn/cloudream/common/pkgs/logger"
	"gitlink.org.cn/cloudream/common/utils/io2"
)

func init() {
	exec.UseOp[*SendStream]()
	exec.UseOp[*GetStream]()
	exec.UseOp[*SendVar]()
	exec.UseOp[*GetVar]()
}

type SendStream struct {
	Input  *exec.StreamVar `json:"input"`
	Send   *exec.StreamVar `json:"send"`
	Worker exec.WorkerInfo `json:"worker"`
}

func (o *SendStream) Execute(ctx context.Context, e *exec.Executor) error {
	err := e.BindVars(ctx, o.Input)
	if err != nil {
		return err
	}
	defer o.Input.Stream.Close()

	cli, err := o.Worker.NewClient()
	if err != nil {
		return fmt.Errorf("new worker %v client: %w", o.Worker, err)
	}
	defer cli.Close()

	logger.Debugf("sending stream %v as %v to worker %v", o.Input.ID, o.Send.ID, o.Worker)

	// 发送后流的ID不同
	err = cli.SendStream(ctx, e.Plan().ID, o.Send, o.Input.Stream)
	if err != nil {
		return fmt.Errorf("sending stream: %w", err)
	}

	return nil
}

func (o *SendStream) String() string {
	return fmt.Sprintf("SendStream %v->%v@%v", o.Input.ID, o.Send.ID, o.Worker)
}

type GetStream struct {
	Signal *exec.SignalVar `json:"signal"`
	Target *exec.StreamVar `json:"target"`
	Output *exec.StreamVar `json:"output"`
	Worker exec.WorkerInfo `json:"worker"`
}

func (o *GetStream) Execute(ctx context.Context, e *exec.Executor) error {
	cli, err := o.Worker.NewClient()
	if err != nil {
		return fmt.Errorf("new worker %v client: %w", o.Worker, err)
	}
	defer cli.Close()

	logger.Debugf("getting stream %v as %v from worker %v", o.Target.ID, o.Output.ID, o.Worker)

	str, err := cli.GetStream(ctx, e.Plan().ID, o.Target, o.Signal)
	if err != nil {
		return fmt.Errorf("getting stream: %w", err)
	}

	fut := future.NewSetVoid()
	// 获取后送到本地的流ID是不同的
	o.Output.Stream = io2.AfterReadClosedOnce(str, func(closer io.ReadCloser) {
		fut.SetVoid()
	})
	e.PutVars(o.Output)

	return fut.Wait(ctx)
}

func (o *GetStream) String() string {
	return fmt.Sprintf("GetStream %v(S:%v)<-%v@%v", o.Output.ID, o.Signal.ID, o.Target.ID, o.Worker)
}

type SendVar struct {
	Input  exec.Var        `json:"input"`
	Send   exec.Var        `json:"send"`
	Worker exec.WorkerInfo `json:"worker"`
}

func (o *SendVar) Execute(ctx context.Context, e *exec.Executor) error {
	err := e.BindVars(ctx, o.Input)
	if err != nil {
		return err
	}

	cli, err := o.Worker.NewClient()
	if err != nil {
		return fmt.Errorf("new worker %v client: %w", o.Worker, err)
	}
	defer cli.Close()

	logger.Debugf("sending var %v as %v to worker %v", o.Input.GetID(), o.Send.GetID(), o.Worker)

	exec.AssignVar(o.Input, o.Send)
	err = cli.SendVar(ctx, e.Plan().ID, o.Send)
	if err != nil {
		return fmt.Errorf("sending var: %w", err)
	}

	return nil
}

func (o *SendVar) String() string {
	return fmt.Sprintf("SendVar %v->%v@%v", o.Input.GetID(), o.Send.GetID(), o.Worker)
}

type GetVar struct {
	Signal *exec.SignalVar `json:"signal"`
	Target exec.Var        `json:"target"`
	Output exec.Var        `json:"output"`
	Worker exec.WorkerInfo `json:"worker"`
}

func (o *GetVar) Execute(ctx context.Context, e *exec.Executor) error {
	cli, err := o.Worker.NewClient()
	if err != nil {
		return fmt.Errorf("new worker %v client: %w", o.Worker, err)
	}
	defer cli.Close()

	logger.Debugf("getting var %v as %v from worker %v", o.Target.GetID(), o.Output.GetID(), o.Worker)

	err = cli.GetVar(ctx, e.Plan().ID, o.Target, o.Signal)
	if err != nil {
		return fmt.Errorf("getting var: %w", err)
	}
	exec.AssignVar(o.Target, o.Output)
	e.PutVars(o.Output)

	return nil
}

func (o *GetVar) String() string {
	return fmt.Sprintf("GetVar %v(S:%v)<-%v@%v", o.Output.GetID(), o.Signal.ID, o.Target.GetID(), o.Worker)
}

type SendStreamType struct {
	ToWorker exec.WorkerInfo
}

func (t *SendStreamType) Send(n *dag.Node, v *dag.StreamVar) *dag.StreamVar {
	v.To(n, 0)
	return n.OutputStreams[0]
}

func (t *SendStreamType) InitNode(node *dag.Node) {
	dag.NodeDeclareInputStream(node, 1)
	dag.NodeNewOutputStream(node, nil)
}

func (t *SendStreamType) GenerateOp(op *dag.Node) (exec.Op, error) {
	return &SendStream{
		Input:  op.InputStreams[0].Var,
		Send:   op.OutputStreams[0].Var,
		Worker: t.ToWorker,
	}, nil
}

func (t *SendStreamType) String(node *dag.Node) string {
	return fmt.Sprintf("SendStream[]%v%v", formatStreamIO(node), formatValueIO(node))
}

type SendVarType struct {
	ToWorker exec.WorkerInfo
}

func (t *SendVarType) Send(n *dag.Node, v *dag.ValueVar) *dag.ValueVar {
	v.To(n, 0)
	n.OutputValues[0].Type = v.Type
	return n.OutputValues[0]
}

func (t *SendVarType) InitNode(node *dag.Node) {
	dag.NodeDeclareInputValue(node, 1)
	dag.NodeNewOutputValue(node, 0, nil)
}

func (t *SendVarType) GenerateOp(op *dag.Node) (exec.Op, error) {
	return &SendVar{
		Input:  op.InputValues[0].Var,
		Send:   op.OutputValues[0].Var,
		Worker: t.ToWorker,
	}, nil
}

func (t *SendVarType) String(node *dag.Node) string {
	return fmt.Sprintf("SendVar[]%v%v", formatStreamIO(node), formatValueIO(node))
}

type GetStreamType struct {
	FromWorker exec.WorkerInfo
}

func (t *GetStreamType) Get(n *dag.Node, v *dag.StreamVar) *dag.StreamVar {
	v.To(n, 0)
	return n.OutputStreams[0]
}

func (t *GetStreamType) SignalVar(n *dag.Node) *dag.ValueVar {
	return n.OutputValues[0]
}

func (t *GetStreamType) InitNode(node *dag.Node) {
	dag.NodeDeclareInputStream(node, 1)
	dag.NodeNewOutputValue(node, dag.SignalValueVar, nil)
	dag.NodeNewOutputStream(node, nil)
}

func (t *GetStreamType) GenerateOp(op *dag.Node) (exec.Op, error) {
	return &GetStream{
		Signal: op.OutputValues[0].Var.(*exec.SignalVar),
		Output: op.OutputStreams[0].Var,
		Target: op.InputStreams[0].Var,
		Worker: t.FromWorker,
	}, nil
}

func (t *GetStreamType) String(node *dag.Node) string {
	return fmt.Sprintf("GetStream[]%v%v", formatStreamIO(node), formatValueIO(node))
}

type GetVaType struct {
	FromWorker exec.WorkerInfo
}

func (t *GetVaType) Get(n *dag.Node, v *dag.ValueVar) *dag.ValueVar {
	v.To(n, 0)
	n.OutputValues[1].Type = v.Type
	return n.OutputValues[1]
}

func (t *GetVaType) SignalVar(n *dag.Node) *dag.ValueVar {
	return n.OutputValues[0]
}

func (t *GetVaType) InitNode(node *dag.Node) {
	dag.NodeDeclareInputValue(node, 1)
	dag.NodeNewOutputValue(node, dag.SignalValueVar, nil)
	dag.NodeNewOutputValue(node, 0, nil)
}

func (t *GetVaType) GenerateOp(op *dag.Node) (exec.Op, error) {
	return &GetVar{
		Signal: op.OutputValues[0].Var.(*exec.SignalVar),
		Output: op.OutputValues[1].Var,
		Target: op.InputValues[0].Var,
		Worker: t.FromWorker,
	}, nil
}

func (t *GetVaType) String(node *dag.Node) string {
	return fmt.Sprintf("GetVar[]%v%v", formatStreamIO(node), formatValueIO(node))
}
