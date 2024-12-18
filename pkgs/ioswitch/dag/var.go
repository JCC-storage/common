package dag

import (
	"gitlink.org.cn/cloudream/common/pkgs/ioswitch/exec"
	"gitlink.org.cn/cloudream/common/utils/lo2"
)

type Var interface {
	GetVarID() exec.VarID
}

type StreamVar struct {
	VarID exec.VarID
	Src   Node
	Dst   DstList
}

func (v *StreamVar) GetVarID() exec.VarID {
	return v.VarID
}

func (v *StreamVar) IndexAtSrc() int {
	return v.Src.OutputStreams().IndexOf(v)
}

func (v *StreamVar) To(to Node, slotIdx int) {
	v.Dst.Add(to)
	to.InputStreams().Slots.Set(slotIdx, v)
}

func (v *StreamVar) ToSlot(slot StreamInputSlot) {
	v.Dst.Add(slot.Node)
	slot.Node.InputStreams().Slots.Set(slot.Index, v)
}

func (v *StreamVar) NotTo(node Node) {
	v.Dst.Remove(node)
	node.InputStreams().Slots.Clear(v)
}

func (v *StreamVar) ClearAllDst() {
	for _, n := range v.Dst {
		n.InputStreams().Slots.Clear(v)
	}
	v.Dst = nil
}

type ValueVar struct {
	VarID exec.VarID
	Src   Node
	Dst   DstList
}

func (v *ValueVar) GetVarID() exec.VarID {
	return v.VarID
}

func (v *ValueVar) IndexAtSrc() int {
	return v.Src.InputValues().IndexOf(v)
}

func (v *ValueVar) To(to Node, slotIdx int) {
	v.Dst.Add(to)
	to.InputValues().Slots.Set(slotIdx, v)
}

func (v *ValueVar) ToSlot(slot ValueInputSlot) {
	v.Dst.Add(slot.Node)
	slot.Node.InputValues().Slots.Set(slot.Index, v)
}

func (v *ValueVar) NotTo(node Node) {
	v.Dst.Remove(node)
	node.InputValues().Slots.Clear(v)
}

func (v *ValueVar) ClearAllDst() {
	for _, n := range v.Dst {
		n.InputValues().Slots.Clear(v)
	}
	v.Dst = nil
}

type DstList []Node

func (s *DstList) Len() int {
	return len(*s)
}

func (s *DstList) Get(idx int) Node {
	return (*s)[idx]
}

func (s *DstList) Add(n Node) int {
	(*s) = append((*s), n)
	return len(*s) - 1
}

func (s *DstList) Remove(n Node) {
	for i, e := range *s {
		if e == n {
			(*s) = lo2.RemoveAt((*s), i)
			return
		}
	}
}

func (s *DstList) RemoveAt(idx int) {
	lo2.RemoveAt((*s), idx)
}

func (s *DstList) Resize(size int) {
	if s.Len() < size {
		(*s) = append((*s), make([]Node, size-s.Len())...)
	} else if s.Len() > size {
		(*s) = (*s)[:size]
	}
}

func (s *DstList) RawArray() []Node {
	return *s
}
