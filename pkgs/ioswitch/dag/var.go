package dag

import (
	"github.com/samber/lo"
	"gitlink.org.cn/cloudream/common/pkgs/ioswitch/exec"
	"gitlink.org.cn/cloudream/common/utils/lo2"
)

type EndPoint struct {
	Node      Node
	SlotIndex int // 所连接的Node的Output或Input数组的索引
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

type Var struct {
	VarID exec.VarID
	src   Node
	dst   DstList
}

func (v *Var) From() Node {
	return v.src
}

func (v *Var) To() *DstList {
	return &v.dst
}

func (v *Var) StreamIndexOfFrom() int {
	return lo.IndexOf(v.src.OutputStreams().RawArray(), v)
}

func (v *Var) ValueIndexOfFrom() int {
	return lo.IndexOf(v.src.InputValues().RawArray(), v)
}

func (v *Var) ValueTo(to Node, slotIdx int) {
	v.To().Add(to)
	to.InputValues().Set(slotIdx, v)
}

func (v *Var) ValueNotTo(node Node, slotIdx int) {
	v.dst.Remove(node)
	node.InputValues().Set(slotIdx, nil)
}

func (v *Var) StreamTo(to Node, slotIdx int) {
	v.To().Add(to)
	to.InputStreams().Set(slotIdx, v)
}

func (v *Var) StreamNotTo(node Node, slotIdx int) {
	v.dst.Remove(node)
	node.InputStreams().Set(slotIdx, nil)
}

func (v *Var) NoInputAllValue() {
	for _, n := range v.dst {
		n.InputValues().ClearInput(v)
	}
	v.dst = nil
}

func (v *Var) NoInputAllStream() {
	for _, n := range v.dst {
		n.InputStreams().ClearInput(v)
	}
	v.dst = nil
}
