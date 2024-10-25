package ops

import (
	"fmt"
	"io"

	"gitlink.org.cn/cloudream/common/pkgs/ioswitch/dag"
	"gitlink.org.cn/cloudream/common/pkgs/ioswitch/exec"
)

func init() {
	exec.UseOp[*DropStream]()
}

type DropStream struct {
	Input exec.VarID `json:"input"`
}

func (o *DropStream) Execute(ctx *exec.ExecContext, e *exec.Executor) error {
	str, err := exec.BindVar[*exec.StreamValue](e, ctx.Context, o.Input)
	if err != nil {
		return err
	}
	defer str.Stream.Close()

	for {
		buf := make([]byte, 1024*8)
		_, err = str.Stream.Read(buf)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

func (o *DropStream) String() string {
	return fmt.Sprintf("DropStream %v", o.Input)
}

type DropNode struct {
	dag.NodeBase
}

func (b *GraphNodeBuilder) NewDropStream() *DropNode {
	node := &DropNode{}
	b.AddNode(node)
	return node
}

func (t *DropNode) SetInput(v *dag.Var) {
	t.InputStreams().EnsureSize(1)
	v.Connect(t, 0)
}

func (t *DropNode) GenerateOp() (exec.Op, error) {
	return &DropStream{
		Input: t.InputStreams().Get(0).VarID,
	}, nil
}

// func (t *DropType) String(node *dag.Node) string {
// 	return fmt.Sprintf("Drop[]%v%v", formatStreamIO(node), formatValueIO(node))
// }
