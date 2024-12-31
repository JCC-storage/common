package ops

import (
	"gitlink.org.cn/cloudream/common/pkgs/ioswitch/dag"
	"gitlink.org.cn/cloudream/common/pkgs/ioswitch/exec"
	"gitlink.org.cn/cloudream/common/utils/math2"
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

	node.OutputStreams().Init(node, 1)

	return node
}

func (t *FromDriverNode) Output() dag.StreamOutputSlot {
	return dag.StreamOutputSlot{
		Node:  t,
		Index: 0,
	}
}

func (t *FromDriverNode) GenerateOp() (exec.Op, error) {
	t.Handle.ID = t.OutputStreams().Get(0).VarID
	return nil, nil
}

// func (t *FromDriverType) String() string {
// 	return fmt.Sprintf("FromDriver[]%v%v", formatStreamIO(node), formatValueIO(node))
// }

type ToDriverNode struct {
	dag.NodeBase
	Handle *exec.DriverReadStream
	Range  math2.Range
}

func (b *GraphNodeBuilder) NewToDriver(handle *exec.DriverReadStream) *ToDriverNode {
	node := &ToDriverNode{
		Handle: handle,
	}
	b.AddNode(node)

	node.InputStreams().Init(1)
	return node
}

func (t *ToDriverNode) SetInput(v *dag.StreamVar) {
	v.To(t, 0)
}

func (t *ToDriverNode) Input() dag.StreamOutputSlot {
	return dag.StreamOutputSlot{
		Node:  t,
		Index: 0,
	}
}

func (t *ToDriverNode) GenerateOp() (exec.Op, error) {
	t.Handle.ID = t.InputStreams().Get(0).VarID
	return nil, nil
}

// func (t *ToDriverType) String() string {
// 	return fmt.Sprintf("ToDriver[%v+%v]%v%v", t.Range.Offset, t.Range.Length, formatStreamIO(node), formatValueIO(node))
// }
