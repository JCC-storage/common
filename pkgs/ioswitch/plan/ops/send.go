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

type SendStreamNode struct {
	dag.NodeBase
	ToWorker exec.WorkerInfo
}

func (b *GraphNodeBuilder) NewSendStream(to exec.WorkerInfo) *SendStreamNode {
	node := &SendStreamNode{
		ToWorker: to,
	}
	b.AddNode(node)
	return node
}

func (t *SendStreamNode) Send(v *dag.StreamVar) *dag.StreamVar {
	t.InputStreams().EnsureSize(1)
	v.Connect(t, 0)
	output := t.Graph().NewStreamVar()
	t.OutputStreams().Setup(t, output, 0)
	return output
}

func (t *SendStreamNode) GenerateOp() (exec.Op, error) {
	return &SendStream{
		Input:  t.InputStreams().Get(0).Var,
		Send:   t.OutputStreams().Get(0).Var,
		Worker: t.ToWorker,
	}, nil
}

// func (t *SendStreamType) String() string {
// 	return fmt.Sprintf("SendStream[]%v%v", formatStreamIO(node), formatValueIO(node))
// }

type SendValueNode struct {
	dag.NodeBase
	ToWorker exec.WorkerInfo
}

func (b *GraphNodeBuilder) NewSendValue(to exec.WorkerInfo) *SendValueNode {
	node := &SendValueNode{
		ToWorker: to,
	}
	b.AddNode(node)
	return node
}

func (t *SendValueNode) Send(v *dag.ValueVar) *dag.ValueVar {
	t.InputValues().EnsureSize(1)
	v.Connect(t, 0)
	output := t.Graph().NewValueVar(v.Type)
	t.OutputValues().Setup(t, output, 0)
	return output
}

func (t *SendValueNode) GenerateOp() (exec.Op, error) {
	return &SendVar{
		Input:  t.InputValues().Get(0).Var,
		Send:   t.OutputValues().Get(0).Var,
		Worker: t.ToWorker,
	}, nil
}

// func (t *SendVarType) String() string {
// 	return fmt.Sprintf("SendVar[]%v%v", formatStreamIO(node), formatValueIO(node))
// }

type GetStreamNode struct {
	dag.NodeBase
	FromWorker exec.WorkerInfo
}

func (b *GraphNodeBuilder) NewGetStream(from exec.WorkerInfo) *GetStreamNode {
	node := &GetStreamNode{
		FromWorker: from,
	}
	b.AddNode(node)
	node.OutputValues().Setup(node, node.Graph().NewValueVar(dag.SignalValueVar), 0)
	return node
}

func (t *GetStreamNode) Get(v *dag.StreamVar) *dag.StreamVar {
	t.InputStreams().EnsureSize(1)
	v.Connect(t, 0)
	output := t.Graph().NewStreamVar()
	t.OutputStreams().Setup(t, output, 0)
	return output
}

func (t *GetStreamNode) SignalVar() *dag.ValueVar {
	return t.OutputValues().Get(0)
}

func (t *GetStreamNode) GenerateOp() (exec.Op, error) {
	return &GetStream{
		Signal: t.OutputValues().Get(0).Var.(*exec.SignalVar),
		Output: t.OutputStreams().Get(0).Var,
		Target: t.InputStreams().Get(0).Var,
		Worker: t.FromWorker,
	}, nil
}

// func (t *GetStreamType) String() string {
// 	return fmt.Sprintf("GetStream[]%v%v", formatStreamIO(node), formatValueIO(node))
// }

type GetValueNode struct {
	dag.NodeBase
	FromWorker exec.WorkerInfo
}

func (b *GraphNodeBuilder) NewGetValue(from exec.WorkerInfo) *GetValueNode {
	node := &GetValueNode{
		FromWorker: from,
	}
	b.AddNode(node)
	node.OutputValues().Setup(node, node.Graph().NewValueVar(dag.SignalValueVar), 0)
	return node
}

func (t *GetValueNode) Get(v *dag.ValueVar) *dag.ValueVar {
	t.InputValues().EnsureSize(1)
	v.Connect(t, 0)
	output := t.Graph().NewValueVar(v.Type)
	t.OutputValues().Setup(t, output, 1)
	return output
}

func (t *GetValueNode) SignalVar() *dag.ValueVar {
	return t.OutputValues().Get(0)
}

func (t *GetValueNode) GenerateOp() (exec.Op, error) {
	return &GetVar{
		Signal: t.OutputValues().Get(0).Var.(*exec.SignalVar),
		Output: t.OutputValues().Get(1).Var,
		Target: t.InputValues().Get(0).Var,
		Worker: t.FromWorker,
	}, nil
}

// func (t *GetVaType) String() string {
// 	return fmt.Sprintf("GetVar[]%v%v", formatStreamIO(node), formatValueIO(node))
// }
