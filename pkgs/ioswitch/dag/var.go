package dag

import (
	"gitlink.org.cn/cloudream/common/pkgs/ioswitch/exec"
	"gitlink.org.cn/cloudream/common/utils/lo2"
)

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

type Var struct {
	VarID exec.VarID
	from  EndPoint
	to    EndPointSlots
}

func (v *Var) From() *EndPoint {
	return &v.from
}

func (v *Var) To() *EndPointSlots {
	return &v.to
}

func (v *Var) Connect(to Node, slotIdx int) {
	v.To().Add(EndPoint{Node: to, SlotIndex: slotIdx})
	to.InputValues().Set(slotIdx, v)
}

func (v *Var) Disconnect(node Node, slotIdx int) {
	v.to.Remove(EndPoint{Node: node, SlotIndex: slotIdx})
	node.InputValues().Set(slotIdx, nil)
}

func (v *Var) DisconnectAll() {
	for _, ed := range v.to {
		ed.Node.InputStreams().Set(ed.SlotIndex, nil)
	}
	v.to = nil
}
