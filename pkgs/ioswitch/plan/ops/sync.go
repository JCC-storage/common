package ops

import (
	"context"
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
	Raw    *exec.StreamVar `json:"raw"`
	New    *exec.StreamVar `json:"new"`
	Signal *exec.SignalVar `json:"signal"`
}

func (o *OnStreamBegin) Execute(ctx context.Context, e *exec.Executor) error {
	err := e.BindVars(ctx, o.Raw)
	if err != nil {
		return err
	}

	o.New.Stream = o.Raw.Stream

	e.PutVars(o.New, o.Signal)
	return nil
}

func (o *OnStreamBegin) String() string {
	return fmt.Sprintf("OnStreamBegin %v->%v S:%v", o.Raw.ID, o.New.ID, o.Signal.ID)
}

type OnStreamEnd struct {
	Raw    *exec.StreamVar `json:"raw"`
	New    *exec.StreamVar `json:"new"`
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

func (o *OnStreamEnd) Execute(ctx context.Context, e *exec.Executor) error {
	err := e.BindVars(ctx, o.Raw)
	if err != nil {
		return err
	}

	cb := future.NewSetVoid()

	o.New.Stream = &onStreamEnd{
		inner:    o.Raw.Stream,
		callback: cb,
	}
	e.PutVars(o.New)

	err = cb.Wait(ctx)
	if err != nil {
		return err
	}

	e.PutVars(o.Signal)
	return nil
}

func (o *OnStreamEnd) String() string {
	return fmt.Sprintf("OnStreamEnd %v->%v S:%v", o.Raw.ID, o.New.ID, o.Signal.ID)
}

type HoldUntil struct {
	Waits []*exec.SignalVar `json:"waits"`
	Holds []exec.Var        `json:"holds"`
	Emits []exec.Var        `json:"emits"`
}

func (w *HoldUntil) Execute(ctx context.Context, e *exec.Executor) error {
	err := e.BindVars(ctx, w.Holds...)
	if err != nil {
		return err
	}

	err = exec.BindArrayVars(e, ctx, w.Waits)
	if err != nil {
		return err
	}

	for i := 0; i < len(w.Holds); i++ {
		err := exec.AssignVar(w.Holds[i], w.Emits[i])
		if err != nil {
			return err
		}
	}

	e.PutVars(w.Emits...)
	return nil
}

func (w *HoldUntil) String() string {
	return fmt.Sprintf("HoldUntil Waits: %v, (%v) -> (%v)", utils.FormatVarIDs(w.Waits), utils.FormatVarIDs(w.Holds), utils.FormatVarIDs(w.Emits))
}

type HangUntil struct {
	Waits []*exec.SignalVar `json:"waits"`
	Op    exec.Op           `json:"op"`
}

func (h *HangUntil) Execute(ctx context.Context, e *exec.Executor) error {
	err := exec.BindArrayVars(e, ctx, h.Waits)
	if err != nil {
		return err
	}

	return h.Op.Execute(ctx, e)
}

func (h *HangUntil) String() string {
	return "HangUntil"
}

type Broadcast struct {
	Source  *exec.SignalVar   `json:"source"`
	Targets []*exec.SignalVar `json:"targets"`
}

func (b *Broadcast) Execute(ctx context.Context, e *exec.Executor) error {
	err := e.BindVars(ctx, b.Source)
	if err != nil {
		return err
	}

	exec.PutArrayVars(e, b.Targets)
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

func (t *HoldUntilNode) SetSignal(s *dag.ValueVar) {
	t.InputValues().EnsureSize(1)
	s.Connect(t, 0)
}

func (t *HoldUntilNode) HoldStream(str *dag.StreamVar) *dag.StreamVar {
	str.Connect(t, t.InputStreams().EnlargeOne())
	output := t.Graph().NewStreamVar()
	t.OutputStreams().SetupNew(t, output)
	return output
}

func (t *HoldUntilNode) HoldVar(v *dag.ValueVar) *dag.ValueVar {
	v.Connect(t, t.InputValues().EnlargeOne())
	output := t.Graph().NewValueVar(v.Type)
	t.OutputValues().SetupNew(t, output)
	return output
}

func (t *HoldUntilNode) GenerateOp() (exec.Op, error) {
	o := &HoldUntil{
		Waits: []*exec.SignalVar{t.InputValues().Get(0).Var.(*exec.SignalVar)},
	}

	for i := 0; i < t.OutputValues().Len(); i++ {
		o.Holds = append(o.Holds, t.InputValues().Get(i+1).Var)
		o.Emits = append(o.Emits, t.OutputValues().Get(i).Var)
	}

	for i := 0; i < t.OutputStreams().Len(); i++ {
		o.Holds = append(o.Holds, t.InputStreams().Get(i).Var)
		o.Emits = append(o.Emits, t.OutputStreams().Get(i).Var)
	}

	return o, nil
}

// func (t *HoldUntilType) String() string {
// 	return fmt.Sprintf("HoldUntil[]%v%v", formatStreamIO(node), formatValueIO(node))
// }
