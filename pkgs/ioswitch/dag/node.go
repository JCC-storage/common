package dag

import (
	"gitlink.org.cn/cloudream/common/pkgs/ioswitch/exec"
	"gitlink.org.cn/cloudream/common/utils/lo2"
)

type NodeEnvType string

const (
	EnvUnknown NodeEnvType = ""
	EnvDriver  NodeEnvType = "Driver"
	EnvWorker  NodeEnvType = "Worker"
)

type NodeEnv struct {
	Type   NodeEnvType
	Worker exec.WorkerInfo
	Pinned bool // 如果为true，则不应该改变这个节点的执行环境
}

func (e *NodeEnv) ToEnvUnknown() {
	e.Type = EnvUnknown
	e.Worker = nil
}

func (e *NodeEnv) ToEnvDriver() {
	e.Type = EnvDriver
	e.Worker = nil
}

func (e *NodeEnv) ToEnvWorker(worker exec.WorkerInfo) {
	e.Type = EnvWorker
	e.Worker = worker
}

func (e *NodeEnv) Equals(other *NodeEnv) bool {
	if e.Type != other.Type {
		return false
	}

	if e.Type != EnvWorker {
		return true
	}

	return e.Worker.Equals(other.Worker)
}

type Node interface {
	Graph() *Graph
	SetGraph(graph *Graph)
	Env() *NodeEnv
	InputStreams() *InputSlots
	OutputStreams() *OutputSlots
	InputValues() *InputSlots
	OutputValues() *OutputSlots
	GenerateOp() (exec.Op, error)
	// String() string
}

type VarSlots []*Var

func (s *VarSlots) Len() int {
	return len(*s)
}

func (s *VarSlots) Get(idx int) *Var {
	return (*s)[idx]
}

func (s *VarSlots) Set(idx int, val *Var) *Var {
	old := (*s)[idx]
	(*s)[idx] = val
	return old
}

func (s *VarSlots) Append(val *Var) int {
	*s = append(*s, val)
	return s.Len() - 1
}

func (s *VarSlots) RemoveAt(idx int) {
	(*s) = lo2.RemoveAt(*s, idx)
}

func (s *VarSlots) Resize(size int) {
	if s.Len() < size {
		*s = append(*s, make([]*Var, size-s.Len())...)
	} else if s.Len() > size {
		*s = (*s)[:size]
	}
}

func (s *VarSlots) SetRawArray(arr []*Var) {
	*s = arr
}

func (s *VarSlots) RawArray() []*Var {
	return *s
}

type InputSlots struct {
	VarSlots
}

func (s *InputSlots) EnsureSize(cnt int) {
	if s.Len() < cnt {
		s.VarSlots = append(s.VarSlots, make([]*Var, cnt-s.Len())...)
	}
}

func (s *InputSlots) EnlargeOne() int {
	s.Append(nil)
	return s.Len() - 1
}

type OutputSlots struct {
	VarSlots
}

func (s *OutputSlots) Setup(my Node, v *Var, slotIdx int) {
	if s.Len() <= slotIdx {
		s.VarSlots = append(s.VarSlots, make([]*Var, slotIdx-s.Len()+1)...)
	}

	s.Set(slotIdx, v)
	*v.From() = EndPoint{
		Node:      my,
		SlotIndex: slotIdx,
	}
}

func (s *OutputSlots) SetupNew(my Node, v *Var) {
	s.Append(v)
	*v.From() = EndPoint{
		Node:      my,
		SlotIndex: s.Len() - 1,
	}
}

type Slot struct {
	Var   *Var
	Index int
}

type NodeBase struct {
	env           NodeEnv
	inputStreams  InputSlots
	outputStreams OutputSlots
	inputValues   InputSlots
	outputValues  OutputSlots
	graph         *Graph
}

func (n *NodeBase) Graph() *Graph {
	return n.graph
}

func (n *NodeBase) SetGraph(graph *Graph) {
	n.graph = graph
}

func (n *NodeBase) Env() *NodeEnv {
	return &n.env
}

func (n *NodeBase) InputStreams() *InputSlots {
	return &n.inputStreams
}

func (n *NodeBase) OutputStreams() *OutputSlots {
	return &n.outputStreams
}

func (n *NodeBase) InputValues() *InputSlots {
	return &n.inputValues
}

func (n *NodeBase) OutputValues() *OutputSlots {
	return &n.outputValues
}
