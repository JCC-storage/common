package ops

import (
	"fmt"

	"gitlink.org.cn/cloudream/common/pkgs/ioswitch/dag"
	"gitlink.org.cn/cloudream/common/pkgs/ioswitch/exec"
)

type FromDriverType struct {
	Handle *exec.DriverWriteStream
}

func (t *FromDriverType) InitNode(node *dag.Node) {
	dag.NodeNewOutputStream(node, nil)
}

func (t *FromDriverType) GenerateOp(op *dag.Node) (exec.Op, error) {
	t.Handle.Var = op.OutputStreams[0].Var
	return nil, nil
}

func (t *FromDriverType) String(node *dag.Node) string {
	return fmt.Sprintf("FromDriver[]%v%v", formatStreamIO(node), formatValueIO(node))
}

type ToDriverType struct {
	Handle *exec.DriverReadStream
	Range  exec.Range
}

func (t *ToDriverType) InitNode(node *dag.Node) {
	dag.NodeDeclareInputStream(node, 1)
}

func (t *ToDriverType) GenerateOp(op *dag.Node) (exec.Op, error) {
	t.Handle.Var = op.InputStreams[0].Var
	return nil, nil
}

func (t *ToDriverType) String(node *dag.Node) string {
	return fmt.Sprintf("ToDriver[%v+%v]%v%v", t.Range.Offset, t.Range.Length, formatStreamIO(node), formatValueIO(node))
}
