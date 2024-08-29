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

type DropType struct{}

func (t *DropType) InitNode(node *dag.Node) {
	dag.NodeDeclareInputStream(node, 1)
}

func (t *DropType) GenerateOp(op *dag.Node) (exec.Op, error) {
	return &DropStream{
		Input: op.InputStreams[0].Var,
	}, nil
}

func (t *DropType) String(node *dag.Node) string {
	return fmt.Sprintf("Drop[]%v%v", formatStreamIO(node), formatValueIO(node))
}
