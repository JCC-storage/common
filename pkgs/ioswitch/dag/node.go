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
	InputStreams() *InputSlots[*StreamVar]
	OutputStreams() *OutputSlots[*StreamVar]
	InputValues() *InputSlots[*ValueVar]
	OutputValues() *OutputSlots[*ValueVar]
	GenerateOp() (exec.Op, error)
	// String() string
}

type VarSlots[T Var] []T

func (s *VarSlots[T]) Len() int {
	return len(*s)
}

func (s *VarSlots[T]) Get(idx int) T {
	return (*s)[idx]
}

func (s *VarSlots[T]) Set(idx int, val T) T {
	old := (*s)[idx]
	(*s)[idx] = val
	return old
}

func (s *VarSlots[T]) Append(val T) int {
	*s = append(*s, val)
	return s.Len() - 1
}

func (s *VarSlots[T]) RemoveAt(idx int) {
	(*s) = lo2.RemoveAt(*s, idx)
}

func (s *VarSlots[T]) Resize(size int) {
	if s.Len() < size {
		*s = append(*s, make([]T, size-s.Len())...)
	} else if s.Len() > size {
		*s = (*s)[:size]
	}
}

func (s *VarSlots[T]) SetRawArray(arr []T) {
	*s = arr
}

func (s *VarSlots[T]) RawArray() []T {
	return *s
}

type InputSlots[T Var] struct {
	VarSlots[T]
}

func (s *InputSlots[T]) EnsureSize(cnt int) {
	if s.Len() < cnt {
		s.VarSlots = append(s.VarSlots, make([]T, cnt-s.Len())...)
	}
}

func (s *InputSlots[T]) EnlargeOne() int {
	var t T
	s.Append(t)
	return s.Len() - 1
}

type OutputSlots[T Var] struct {
	VarSlots[T]
}

func (s *OutputSlots[T]) Setup(my Node, v T, slotIdx int) {
	if s.Len() <= slotIdx {
		s.VarSlots = append(s.VarSlots, make([]T, slotIdx-s.Len()+1)...)
	}

	s.Set(slotIdx, v)
	*v.From() = EndPoint{
		Node:      my,
		SlotIndex: slotIdx,
	}
}

func (s *OutputSlots[T]) SetupNew(my Node, v T) {
	s.Append(v)
	*v.From() = EndPoint{
		Node:      my,
		SlotIndex: s.Len() - 1,
	}
}

type Slot[T Var] struct {
	Var   T
	Index int
}

type StreamSlot = Slot[*StreamVar]

type ValueSlot = Slot[*ValueVar]

type NodeBase struct {
	env           NodeEnv
	inputStreams  InputSlots[*StreamVar]
	outputStreams OutputSlots[*StreamVar]
	inputValues   InputSlots[*ValueVar]
	outputValues  OutputSlots[*ValueVar]
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

func (n *NodeBase) InputStreams() *InputSlots[*StreamVar] {
	return &n.inputStreams
}

func (n *NodeBase) OutputStreams() *OutputSlots[*StreamVar] {
	return &n.outputStreams
}

func (n *NodeBase) InputValues() *InputSlots[*ValueVar] {
	return &n.inputValues
}

func (n *NodeBase) OutputValues() *OutputSlots[*ValueVar] {
	return &n.outputValues
}
