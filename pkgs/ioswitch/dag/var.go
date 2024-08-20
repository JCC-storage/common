package dag

import (
	"gitlink.org.cn/cloudream/common/pkgs/ioswitch/exec"
	"gitlink.org.cn/cloudream/common/utils/lo2"
)

type EndPoint struct {
	Node      *Node
	SlotIndex int // 所连接的Node的Output或Input数组的索引
}

type StreamVar struct {
	ID    int
	From  EndPoint
	Toes  []EndPoint
	Props any
	Var   *exec.StreamVar
}

func (v *StreamVar) To(to *Node, slotIdx int) int {
	v.Toes = append(v.Toes, EndPoint{Node: to, SlotIndex: slotIdx})
	to.InputStreams[slotIdx] = v
	return len(v.Toes) - 1
}

// func (v *StreamVar) NotTo(toIdx int) EndPoint {
// 	ed := v.Toes[toIdx]
// 	lo2.RemoveAt(v.Toes, toIdx)
// 	ed.Node.InputStreams[ed.SlotIndex] = nil
// 	return ed
// }

func (v *StreamVar) NotTo(node *Node) (EndPoint, bool) {
	for i, ed := range v.Toes {
		if ed.Node == node {
			v.Toes = lo2.RemoveAt(v.Toes, i)
			ed.Node.InputStreams[ed.SlotIndex] = nil
			return ed, true
		}
	}

	return EndPoint{}, false
}

func (v *StreamVar) NotToWhere(pred func(to EndPoint) bool) []EndPoint {
	var newToes []EndPoint
	var rmed []EndPoint
	for _, ed := range v.Toes {
		if pred(ed) {
			ed.Node.InputStreams[ed.SlotIndex] = nil
			rmed = append(rmed, ed)
		} else {
			newToes = append(newToes, ed)
		}
	}
	v.Toes = newToes
	return rmed
}

func (v *StreamVar) NotToAll() []EndPoint {
	for _, ed := range v.Toes {
		ed.Node.InputStreams[ed.SlotIndex] = nil
	}
	toes := v.Toes
	v.Toes = nil
	return toes
}

func NodeNewOutputStream(node *Node, props any) *StreamVar {
	str := &StreamVar{
		ID:    node.Graph.genVarID(),
		From:  EndPoint{Node: node, SlotIndex: len(node.OutputStreams)},
		Props: props,
	}
	node.OutputStreams = append(node.OutputStreams, str)
	return str
}

func NodeDeclareInputStream(node *Node, cnt int) {
	node.InputStreams = make([]*StreamVar, cnt)
}

type ValueVarType int

const (
	StringValueVar ValueVarType = iota
	SignalValueVar
)

type ValueVar struct {
	ID    int
	Type  ValueVarType
	From  EndPoint
	Toes  []EndPoint
	Props any
	Var   exec.Var
}

func (v *ValueVar) To(to *Node, slotIdx int) int {
	v.Toes = append(v.Toes, EndPoint{Node: to, SlotIndex: slotIdx})
	to.InputValues[slotIdx] = v
	return len(v.Toes) - 1
}

func NodeNewOutputValue(node *Node, props any) *ValueVar {
	val := &ValueVar{
		ID:    node.Graph.genVarID(),
		From:  EndPoint{Node: node, SlotIndex: len(node.OutputStreams)},
		Props: props,
	}
	node.OutputValues = append(node.OutputValues, val)
	return val
}

func NodeDeclareInputValue(node *Node, cnt int) {
	node.InputValues = make([]*ValueVar, cnt)
}
