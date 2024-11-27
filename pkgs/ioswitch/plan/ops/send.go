package ops

import (
	"fmt"
	"io"

	"gitlink.org.cn/cloudream/common/pkgs/future"
	"gitlink.org.cn/cloudream/common/pkgs/ioswitch/dag"
	"gitlink.org.cn/cloudream/common/pkgs/ioswitch/exec"
	"gitlink.org.cn/cloudream/common/utils/io2"
)

func init() {
	exec.UseOp[*SendStream]()
	exec.UseOp[*GetStream]()
	exec.UseOp[*SendVar]()
	exec.UseOp[*GetVar]()
}

type SendStream struct {
	Input  exec.VarID      `json:"input"`
	Send   exec.VarID      `json:"send"`
	Worker exec.WorkerInfo `json:"worker"`
}

func (o *SendStream) Execute(ctx *exec.ExecContext, e *exec.Executor) error {
	inputStr, err := exec.BindVar[*exec.StreamValue](e, ctx.Context, o.Input)
	if err != nil {
		return err
	}
	defer inputStr.Stream.Close()

	cli, err := o.Worker.NewClient()
	if err != nil {
		return fmt.Errorf("new worker %v client: %w", o.Worker, err)
	}
	defer cli.Close()

	// 发送后流的ID不同
	err = cli.SendStream(ctx.Context, e.Plan().ID, o.Send, inputStr.Stream)
	if err != nil {
		return fmt.Errorf("sending stream: %w", err)
	}

	return nil
}

func (o *SendStream) String() string {
	return fmt.Sprintf("SendStream %v->%v@%v", o.Input, o.Send, o.Worker)
}

type GetStream struct {
	Signal exec.SignalVar  `json:"signal"`
	Target exec.VarID      `json:"target"`
	Output exec.VarID      `json:"output"`
	Worker exec.WorkerInfo `json:"worker"`
}

func (o *GetStream) Execute(ctx *exec.ExecContext, e *exec.Executor) error {
	cli, err := o.Worker.NewClient()
	if err != nil {
		return fmt.Errorf("new worker %v client: %w", o.Worker, err)
	}
	defer cli.Close()

	str, err := cli.GetStream(ctx.Context, e.Plan().ID, o.Target, o.Signal.ID, o.Signal.Value)
	if err != nil {
		return fmt.Errorf("getting stream: %w", err)
	}

	fut := future.NewSetVoid()
	// 获取后送到本地的流ID是不同的
	str = io2.AfterReadClosedOnce(str, func(closer io.ReadCloser) {
		fut.SetVoid()
	})
	e.PutVar(o.Output, &exec.StreamValue{Stream: str})

	return fut.Wait(ctx.Context)
}

func (o *GetStream) String() string {
	return fmt.Sprintf("GetStream %v(S:%v)<-%v@%v", o.Output, o.Signal.ID, o.Target, o.Worker)
}

type SendVar struct {
	Input  exec.VarID      `json:"input"`
	Send   exec.VarID      `json:"send"`
	Worker exec.WorkerInfo `json:"worker"`
}

func (o *SendVar) Execute(ctx *exec.ExecContext, e *exec.Executor) error {
	input, err := e.BindVar(ctx.Context, o.Input)
	if err != nil {
		return err
	}

	cli, err := o.Worker.NewClient()
	if err != nil {
		return fmt.Errorf("new worker %v client: %w", o.Worker, err)
	}
	defer cli.Close()

	err = cli.SendVar(ctx.Context, e.Plan().ID, o.Send, input)
	if err != nil {
		return fmt.Errorf("sending var: %w", err)
	}

	return nil
}

func (o *SendVar) String() string {
	return fmt.Sprintf("SendVar %v->%v@%v", o.Input, o.Send, o.Worker)
}

type GetVar struct {
	Signal exec.SignalVar  `json:"signal"`
	Target exec.VarID      `json:"target"`
	Output exec.VarID      `json:"output"`
	Worker exec.WorkerInfo `json:"worker"`
}

func (o *GetVar) Execute(ctx *exec.ExecContext, e *exec.Executor) error {
	cli, err := o.Worker.NewClient()
	if err != nil {
		return fmt.Errorf("new worker %v client: %w", o.Worker, err)
	}
	defer cli.Close()

	get, err := cli.GetVar(ctx.Context, e.Plan().ID, o.Target, o.Signal.ID, o.Signal.Value)
	if err != nil {
		return fmt.Errorf("getting var: %w", err)
	}

	e.PutVar(o.Output, get)

	return nil
}

func (o *GetVar) String() string {
	return fmt.Sprintf("GetVar %v(S:%v)<-%v@%v", o.Output, o.Signal.ID, o.Target, o.Worker)
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

	node.InputStreams().Init(1)
	node.OutputStreams().Init(node, 1)
	return node
}

func (t *SendStreamNode) Send(v *dag.StreamVar) *dag.StreamVar {
	v.To(t, 0)
	return t.OutputStreams().Get(0)
}

func (t *SendStreamNode) GenerateOp() (exec.Op, error) {
	return &SendStream{
		Input:  t.InputStreams().Get(0).VarID,
		Send:   t.OutputStreams().Get(0).VarID,
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

	node.InputValues().Init(1)
	node.OutputValues().Init(node, 1)
	return node
}

func (t *SendValueNode) Send(v *dag.ValueVar) *dag.ValueVar {
	v.To(t, 0)
	return t.OutputValues().Get(0)
}

func (t *SendValueNode) GenerateOp() (exec.Op, error) {
	return &SendVar{
		Input:  t.InputValues().Get(0).VarID,
		Send:   t.OutputValues().Get(0).VarID,
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

	node.InputStreams().Init(1)
	node.OutputValues().Init(node, 1)
	node.OutputStreams().Init(node, 1)
	return node
}

func (t *GetStreamNode) Get(v *dag.StreamVar) *dag.StreamVar {
	v.To(t, 0)
	return t.OutputStreams().Get(0)
}

func (t *GetStreamNode) SignalVar() *dag.ValueVar {
	return t.OutputValues().Get(0)
}

func (t *GetStreamNode) GenerateOp() (exec.Op, error) {
	return &GetStream{
		Signal: exec.NewSignalVar(t.OutputValues().Get(0).VarID),
		Output: t.OutputStreams().Get(0).VarID,
		Target: t.InputStreams().Get(0).VarID,
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

	node.InputValues().Init(1)
	node.OutputValues().Init(node, 2)
	return node
}

func (t *GetValueNode) Get(v *dag.ValueVar) *dag.ValueVar {
	v.To(t, 0)
	return t.OutputValues().Get(1)
}

func (t *GetValueNode) SignalVar() *dag.ValueVar {
	return t.OutputValues().Get(0)
}

func (t *GetValueNode) GenerateOp() (exec.Op, error) {
	return &GetVar{
		Signal: exec.NewSignalVar(t.OutputValues().Get(0).VarID),
		Output: t.OutputValues().Get(1).VarID,
		Target: t.InputValues().Get(0).VarID,
		Worker: t.FromWorker,
	}, nil
}

// func (t *GetVaType) String() string {
// 	return fmt.Sprintf("GetVar[]%v%v", formatStreamIO(node), formatValueIO(node))
// }
