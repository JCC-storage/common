package ops

import (
	"fmt"
	"io"

	"gitlink.org.cn/cloudream/common/pkgs/future"
	"gitlink.org.cn/cloudream/common/pkgs/ioswitch/dag"
	"gitlink.org.cn/cloudream/common/pkgs/ioswitch/exec"
	"gitlink.org.cn/cloudream/common/pkgs/ioswitch/utils"
)

func init() {
	exec.UseOp[*OnStreamBegin]()
	exec.UseOp[*OnStreamEnd]()
	exec.UseOp[*HoldUntil]()
	exec.UseOp[*HangUntil]()
	exec.UseOp[*Broadcast]()
}

type OnStreamBegin struct {
	Raw    exec.VarID     `json:"raw"`
	New    exec.VarID     `json:"new"`
	Signal exec.SignalVar `json:"signal"`
}

func (o *OnStreamBegin) Execute(ctx *exec.ExecContext, e *exec.Executor) error {
	raw, err := exec.BindVar[*exec.StreamValue](e, ctx.Context, o.Raw)
	if err != nil {
		return err
	}

	e.PutVar(o.New, &exec.StreamValue{Stream: raw.Stream}).
		PutVar(o.Signal.ID, o.Signal.Value)
	return nil
}

func (o *OnStreamBegin) String() string {
	return fmt.Sprintf("OnStreamBegin %v->%v S:%v", o.Raw, o.New, o.Signal.ID)
}

type OnStreamEnd struct {
	Raw    exec.VarID      `json:"raw"`
	New    exec.VarID      `json:"new"`
	Signal *exec.SignalVar `json:"signal"`
}

type onStreamEnd struct {
	inner    io.ReadCloser
	callback *future.SetVoidFuture
}

func (o *onStreamEnd) Read(p []byte) (n int, err error) {
	n, err = o.inner.Read(p)
	if err == io.EOF {
		o.callback.SetVoid()
	} else if err != nil {
		o.callback.SetError(err)
	}
	return n, err
}

func (o *onStreamEnd) Close() error {
	o.callback.SetError(fmt.Errorf("stream closed early"))
	return o.inner.Close()
}

func (o *OnStreamEnd) Execute(ctx *exec.ExecContext, e *exec.Executor) error {
	raw, err := exec.BindVar[*exec.StreamValue](e, ctx.Context, o.Raw)
	if err != nil {
		return err
	}

	cb := future.NewSetVoid()

	e.PutVar(o.New, &exec.StreamValue{Stream: &onStreamEnd{
		inner:    raw.Stream,
		callback: cb,
	}})

	err = cb.Wait(ctx.Context)
	if err != nil {
		return err
	}

	e.PutVar(o.Signal.ID, o.Signal.Value)
	return nil
}

func (o *OnStreamEnd) String() string {
	return fmt.Sprintf("OnStreamEnd %v->%v S:%v", o.Raw, o.New, o.Signal.ID)
}

type HoldUntil struct {
	Waits []exec.VarID `json:"waits"`
	Holds []exec.VarID `json:"holds"`
	Emits []exec.VarID `json:"emits"`
}

func (w *HoldUntil) Execute(ctx *exec.ExecContext, e *exec.Executor) error {
	holds, err := exec.BindArray[exec.VarValue](e, ctx.Context, w.Holds)
	if err != nil {
		return err
	}

	_, err = exec.BindArray[exec.VarValue](e, ctx.Context, w.Waits)
	if err != nil {
		return err
	}

	exec.PutArray(e, w.Emits, holds)
	return nil
}

func (w *HoldUntil) String() string {
	return fmt.Sprintf("HoldUntil Waits: %v, (%v) -> (%v)", utils.FormatVarIDs(w.Waits), utils.FormatVarIDs(w.Holds), utils.FormatVarIDs(w.Emits))
}

type HangUntil struct {
	Waits []exec.VarID `json:"waits"`
	Op    exec.Op      `json:"op"`
}

func (h *HangUntil) Execute(ctx *exec.ExecContext, e *exec.Executor) error {
	_, err := exec.BindArray[exec.VarValue](e, ctx.Context, h.Waits)
	if err != nil {
		return err
	}

	return h.Op.Execute(ctx, e)
}

func (h *HangUntil) String() string {
	return "HangUntil"
}

type Broadcast struct {
	Source  exec.VarID   `json:"source"`
	Targets []exec.VarID `json:"targets"`
}

func (b *Broadcast) Execute(ctx *exec.ExecContext, e *exec.Executor) error {
	src, err := exec.BindVar[*exec.SignalValue](e, ctx.Context, b.Source)
	if err != nil {
		return err
	}

	targets := make([]exec.VarValue, len(b.Targets))
	for i := 0; i < len(b.Targets); i++ {
		targets[i] = src.Clone()
	}

	exec.PutArray(e, b.Targets, targets)
	return nil
}

func (b *Broadcast) String() string {
	return "Broadcast"
}

type HoldUntilNode struct {
	dag.NodeBase
}

func (b *GraphNodeBuilder) NewHoldUntil() *HoldUntilNode {
	node := &HoldUntilNode{}
	b.AddNode(node)
	return node
}

func (t *HoldUntilNode) SetSignal(s *dag.Var) {
	t.InputValues().EnsureSize(1)
	s.ValueTo(t, 0)
}

func (t *HoldUntilNode) HoldStream(str *dag.Var) *dag.Var {
	str.StreamTo(t, t.InputStreams().EnlargeOne())
	output := t.Graph().NewVar()
	t.OutputStreams().SetupNew(t, output)
	return output
}

func (t *HoldUntilNode) HoldVar(v *dag.Var) *dag.Var {
	v.ValueTo(t, t.InputValues().EnlargeOne())
	output := t.Graph().NewVar()
	t.OutputValues().SetupNew(t, output)
	return output
}

func (t *HoldUntilNode) GenerateOp() (exec.Op, error) {
	o := &HoldUntil{
		Waits: []exec.VarID{t.InputValues().Get(0).VarID},
	}

	for i := 0; i < t.OutputValues().Len(); i++ {
		o.Holds = append(o.Holds, t.InputValues().Get(i+1).VarID)
		o.Emits = append(o.Emits, t.OutputValues().Get(i).VarID)
	}

	for i := 0; i < t.OutputStreams().Len(); i++ {
		o.Holds = append(o.Holds, t.InputStreams().Get(i).VarID)
		o.Emits = append(o.Emits, t.OutputStreams().Get(i).VarID)
	}

	return o, nil
}

// func (t *HoldUntilType) String() string {
// 	return fmt.Sprintf("HoldUntil[]%v%v", formatStreamIO(node), formatValueIO(node))
// }
