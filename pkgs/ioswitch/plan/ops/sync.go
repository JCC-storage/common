package ops

import (
	"context"
	"fmt"
	"io"

	"gitlink.org.cn/cloudream/common/pkgs/future"
	"gitlink.org.cn/cloudream/common/pkgs/ioswitch/dag"
	"gitlink.org.cn/cloudream/common/pkgs/ioswitch/exec"
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

type HoldUntilType struct {
}

func (t *HoldUntilType) InitNode(node *dag.Node) {
	dag.NodeDeclareInputValue(node, 1)
}

func (t *HoldUntilType) GenerateOp(op *dag.Node) (exec.Op, error) {
	o := &HoldUntil{
		Waits: []*exec.SignalVar{op.InputValues[0].Var.(*exec.SignalVar)},
	}

	for i := 0; i < len(op.OutputValues); i++ {
		o.Holds = append(o.Holds, op.InputValues[i+1].Var)
		o.Emits = append(o.Emits, op.OutputValues[i].Var)
	}

	for i := 0; i < len(op.OutputStreams); i++ {
		o.Holds = append(o.Holds, op.InputStreams[i].Var)
		o.Emits = append(o.Emits, op.OutputStreams[i].Var)
	}

	return o, nil
}

func (t *HoldUntilType) Signal(n *dag.Node, s *dag.ValueVar) {
	s.To(n, 0)
}

func (t *HoldUntilType) HoldStream(n *dag.Node, str *dag.StreamVar) *dag.StreamVar {
	n.InputStreams = append(n.InputStreams, nil)
	str.To(n, len(n.InputStreams)-1)

	return dag.NodeNewOutputStream(n, nil)
}

func (t *HoldUntilType) HoldVar(n *dag.Node, v *dag.ValueVar) *dag.ValueVar {
	n.InputValues = append(n.InputValues, nil)
	v.To(n, len(n.InputValues)-1)

	return dag.NodeNewOutputValue(n, v.Type, nil)
}

func (t *HoldUntilType) String(node *dag.Node) string {
	return fmt.Sprintf("HoldUntil[]%v%v", formatStreamIO(node), formatValueIO(node))
}
