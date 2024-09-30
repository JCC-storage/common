package ops

import (
	"context"
	"fmt"
	"io"

	"gitlink.org.cn/cloudream/common/pkgs/ioswitch/dag"
	"gitlink.org.cn/cloudream/common/pkgs/ioswitch/exec"
)

func init() {
	exec.UseOp[*DropStream]()
}

type DropStream struct {
	Input *exec.StreamVar `json:"input"`
}

func (o *DropStream) Execute(ctx context.Context, e *exec.Executor) error {
	err := e.BindVars(ctx, o.Input)
	if err != nil {
		return err
	}
	defer o.Input.Stream.Close()

	for {
		buf := make([]byte, 1024*8)
		_, err = o.Input.Stream.Read(buf)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

func (o *DropStream) String() string {
	return fmt.Sprintf("DropStream %v", o.Input.ID)
}

type DropNode struct {
	dag.NodeBase
}

func (b *GraphNodeBuilder) NewDropStream() *DropNode {
	node := &DropNode{}
	b.AddNode(node)
	return node
}

func (t *DropNode) SetInput(v *dag.StreamVar) {
	t.InputStreams().EnsureSize(1)
	v.Connect(t, 0)
}

func (t *DropNode) GenerateOp() (exec.Op, error) {
	return &DropStream{
		Input: t.InputStreams().Get(0).Var,
	}, nil
}

// func (t *DropType) String(node *dag.Node) string {
// 	return fmt.Sprintf("Drop[]%v%v", formatStreamIO(node), formatValueIO(node))
// }
