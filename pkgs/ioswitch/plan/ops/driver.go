package ops

import (
	"gitlink.org.cn/cloudream/common/pkgs/ioswitch/dag"
	"gitlink.org.cn/cloudream/common/pkgs/ioswitch/exec"
)

type FromDriverNode struct {
	dag.NodeBase
	Handle *exec.DriverWriteStream
}

func (b *GraphNodeBuilder) NewFromDriver(handle *exec.DriverWriteStream) *FromDriverNode {
	node := &FromDriverNode{
		Handle: handle,
	}
	b.AddNode(node)

	node.OutputStreams().SetupNew(node, b.NewStreamVar())

	return node
}

func (t *FromDriverNode) Output() dag.StreamSlot {
	return dag.StreamSlot{
		Var:   t.OutputStreams().Get(0),
		Index: 0,
	}
}

func (t *FromDriverNode) GenerateOp() (exec.Op, error) {
	t.Handle.Var = t.OutputStreams().Get(0).Var
	return nil, nil
}

// func (t *FromDriverType) String() string {
// 	return fmt.Sprintf("FromDriver[]%v%v", formatStreamIO(node), formatValueIO(node))
// }

type ToDriverNode struct {
	dag.NodeBase
	Handle *exec.DriverReadStream
	Range  exec.Range
}

func (b *GraphNodeBuilder) NewToDriver(handle *exec.DriverReadStream) *ToDriverNode {
	node := &ToDriverNode{
		Handle: handle,
	}
	b.AddNode(node)

	return node
}

func (t *ToDriverNode) SetInput(v *dag.StreamVar) {
	t.InputStreams().EnsureSize(1)
	v.Connect(t, 0)
}

func (t *ToDriverNode) Input() dag.StreamSlot {
	return dag.StreamSlot{
		Var:   t.InputStreams().Get(0),
		Index: 0,
	}
}

func (t *ToDriverNode) GenerateOp() (exec.Op, error) {
	t.Handle.Var = t.InputStreams().Get(0).Var
	return nil, nil
}

// func (t *ToDriverType) String() string {
// 	return fmt.Sprintf("ToDriver[%v+%v]%v%v", t.Range.Offset, t.Range.Length, formatStreamIO(node), formatValueIO(node))
// }
