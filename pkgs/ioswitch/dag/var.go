package dag

import (
	"gitlink.org.cn/cloudream/common/pkgs/ioswitch/exec"
	"gitlink.org.cn/cloudream/common/utils/lo2"
)

type Var interface {
	ID() int
	From() *EndPoint
	To() *EndPointSlots
}

type EndPoint struct {
	Node      Node
	SlotIndex int // 所连接的Node的Output或Input数组的索引
}

type EndPointSlots []EndPoint

func (s *EndPointSlots) Len() int {
	return len(*s)
}

func (s *EndPointSlots) Get(idx int) *EndPoint {
	return &(*s)[idx]
}

func (s *EndPointSlots) Add(ed EndPoint) int {
	(*s) = append((*s), ed)
	return len(*s) - 1
}

func (s *EndPointSlots) Remove(ed EndPoint) {
	for i, e := range *s {
		if e == ed {
			(*s) = lo2.RemoveAt((*s), i)
			return
		}
	}
}

func (s *EndPointSlots) RemoveAt(idx int) {
	lo2.RemoveAt((*s), idx)
}

func (s *EndPointSlots) Resize(size int) {
	if s.Len() < size {
		(*s) = append((*s), make([]EndPoint, size-s.Len())...)
	} else if s.Len() > size {
		(*s) = (*s)[:size]
	}
}

func (s *EndPointSlots) RawArray() []EndPoint {
	return *s
}

type VarBase struct {
	id   int
	from EndPoint
	to   EndPointSlots
}

func (v *VarBase) ID() int {
	return v.id
}

func (v *VarBase) From() *EndPoint {
	return &v.from
}

func (v *VarBase) To() *EndPointSlots {
	return &v.to
}

type StreamVar struct {
	VarBase
	Var *exec.StreamVar
}

func (v *StreamVar) Connect(to Node, slotIdx int) {
	v.To().Add(EndPoint{Node: to, SlotIndex: slotIdx})
	to.InputStreams().Set(slotIdx, v)
}

func (v *StreamVar) Disconnect(node Node, slotIdx int) {
	v.to.Remove(EndPoint{Node: node, SlotIndex: slotIdx})
	node.InputStreams().Set(slotIdx, nil)
}

func (v *StreamVar) DisconnectAll() {
	for _, ed := range v.to {
		ed.Node.InputStreams().Set(ed.SlotIndex, nil)
	}
	v.to = nil
}

type ValueVarType int

const (
	UnknownValueVar ValueVarType = iota
	StringValueVar
	SignalValueVar
)

type ValueVar struct {
	VarBase
	Type ValueVarType
	Var  exec.Var
}

func (v *ValueVar) Connect(to Node, slotIdx int) {
	v.To().Add(EndPoint{Node: to, SlotIndex: slotIdx})
	to.InputValues().Set(slotIdx, v)
}

func (v *ValueVar) Disconnect(node Node, slotIdx int) {
	v.to.Remove(EndPoint{Node: node, SlotIndex: slotIdx})
	node.InputValues().Set(slotIdx, nil)
}
