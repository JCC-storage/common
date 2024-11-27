package dag

import (
	"github.com/samber/lo"
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
	InputStreams() *StreamInputSlots
	OutputStreams() *StreamOutputSlots
	InputValues() *ValueInputSlots
	OutputValues() *ValueOutputSlots
	GenerateOp() (exec.Op, error)
	// String() string
}

type VarSlots[T any] []*T

func (s *VarSlots[T]) Len() int {
	return len(*s)
}

func (s *VarSlots[T]) Get(idx int) *T {
	return (*s)[idx]
}

func (s *VarSlots[T]) Set(idx int, val *T) *T {
	old := (*s)[idx]
	(*s)[idx] = val
	return old
}

func (s *VarSlots[T]) IndexOf(v *T) int {
	return lo.IndexOf(*s, v)
}

func (s *VarSlots[T]) Append(val *T) int {
	*s = append(*s, val)
	return s.Len() - 1
}

func (s *VarSlots[T]) Clear(val *T) {
	for i := 0; i < s.Len(); i++ {
		if (*s)[i] == val {
			(*s)[i] = nil
		}
	}
}

func (s *VarSlots[T]) RemoveAt(idx int) {
	(*s) = lo2.RemoveAt(*s, idx)
}

func (s *VarSlots[T]) Resize(size int) {
	if s.Len() < size {
		*s = append(*s, make([]*T, size-s.Len())...)
	} else if s.Len() > size {
		*s = (*s)[:size]
	}
}

func (s *VarSlots[T]) SetRawArray(arr []*T) {
	*s = arr
}

func (s *VarSlots[T]) RawArray() []*T {
	return *s
}

type StreamInputSlots struct {
	Slots VarSlots[StreamVar]
}

func (s *StreamInputSlots) Len() int {
	return s.Slots.Len()
}

func (s *StreamInputSlots) Get(idx int) *StreamVar {
	return s.Slots.Get(idx)
}

func (s *StreamInputSlots) IndexOf(v *StreamVar) int {
	return s.Slots.IndexOf(v)
}

// 初始化输入流槽。调用者应该保证没有正在使用的槽位（即Slots的每一个元素都为nil）
func (s *StreamInputSlots) Init(cnt int) {
	s.Slots.Resize(cnt)
}

func (s *StreamInputSlots) EnlargeOne() int {
	s.Slots.Append(nil)
	return s.Len() - 1
}

func (s *StreamInputSlots) ClearInputAt(my Node, idx int) {
	v := s.Get(idx)
	if v == nil {
		return
	}
	s.Slots.Set(idx, nil)

	v.Dst.Remove(my)
}

func (s *StreamInputSlots) ClearAllInput(my Node) {
	for i := 0; i < s.Len(); i++ {
		v := s.Get(i)
		if v == nil {
			continue
		}
		s.Slots.Set(i, nil)

		v.Dst.Remove(my)
	}
}

func (s *StreamInputSlots) GetVarIDs() []exec.VarID {
	var ids []exec.VarID
	for _, v := range s.Slots.RawArray() {
		if v == nil {
			continue
		}
		ids = append(ids, v.VarID)
	}

	return ids
}

func (s *StreamInputSlots) GetVarIDsRanged(start, end int) []exec.VarID {
	var ids []exec.VarID
	for i := start; i < end; i++ {
		v := s.Get(i)
		if v == nil {
			continue
		}
		ids = append(ids, v.VarID)
	}

	return ids
}

type ValueInputSlots struct {
	Slots VarSlots[ValueVar]
}

func (s *ValueInputSlots) Len() int {
	return s.Slots.Len()
}

func (s *ValueInputSlots) Get(idx int) *ValueVar {
	return s.Slots.Get(idx)
}

func (s *ValueInputSlots) IndexOf(v *ValueVar) int {
	return s.Slots.IndexOf(v)
}

// 初始化输入流槽。调用者应该保证没有正在使用的槽位（即Slots的每一个元素都为nil）
func (s *ValueInputSlots) Init(cnt int) {
	if s.Len() < cnt {
		s.Slots = append(s.Slots, make([]*ValueVar, cnt-s.Len())...)
	}
}

func (s *ValueInputSlots) EnlargeOne() int {
	s.Slots.Append(nil)
	return s.Len() - 1
}

func (s *ValueInputSlots) ClearInputAt(my Node, idx int) {
	v := s.Get(idx)
	if v == nil {
		return
	}
	s.Slots.Set(idx, nil)

	v.Dst.Remove(my)
}

func (s *ValueInputSlots) GetVarIDs() []exec.VarID {
	var ids []exec.VarID
	for _, v := range s.Slots.RawArray() {
		if v == nil {
			continue
		}
		ids = append(ids, v.VarID)
	}

	return ids
}

func (s *ValueInputSlots) GetVarIDsRanged(start, end int) []exec.VarID {
	var ids []exec.VarID
	for i := start; i < end; i++ {
		v := s.Get(i)
		if v == nil {
			continue
		}
		ids = append(ids, v.VarID)
	}

	return ids
}

type StreamOutputSlots struct {
	Slots VarSlots[StreamVar]
}

func (s *StreamOutputSlots) Len() int {
	return s.Slots.Len()
}

func (s *StreamOutputSlots) Get(idx int) *StreamVar {
	return s.Slots.Get(idx)
}

func (s *StreamOutputSlots) IndexOf(v *StreamVar) int {
	return s.Slots.IndexOf(v)
}

// 设置Slots大小，并为每个Slot创建一个StreamVar。
// 调用者应该保证没有正在使用的输出流，即每一个输出流的Dst都为空。
func (s *StreamOutputSlots) Init(my Node, size int) {
	s.Slots.Resize(size)
	for i := 0; i < size; i++ {
		v := my.Graph().NewStreamVar()
		v.Src = my
		s.Slots.Set(i, v)
	}
}

// 在Slots末尾增加一个StreamVar，并返回它的索引
func (s *StreamOutputSlots) SetupNew(my Node) StreamSlot {
	v := my.Graph().NewStreamVar()
	v.Src = my
	s.Slots.Append(v)
	return StreamSlot{Var: v, Index: s.Len() - 1}
}

// 断开指定位置的输出流到指定节点的连接
func (s *StreamOutputSlots) ClearOutputAt(idx int, dst Node) {
	v := s.Get(idx)
	v.Dst.Remove(dst)
	dst.InputStreams().Slots.Clear(v)
}

// 断开所有输出流的所有连接，完全清空所有输出流。但会保留流变量
func (s *StreamOutputSlots) ClearAllOutput(my Node) {
	for i := 0; i < s.Len(); i++ {
		v := s.Get(i)
		v.ClearAllDst()
	}
}

func (s *StreamOutputSlots) GetVarIDs() []exec.VarID {
	var ids []exec.VarID
	for _, v := range s.Slots.RawArray() {
		if v == nil {
			continue
		}
		ids = append(ids, v.VarID)
	}

	return ids
}

func (s *StreamOutputSlots) GetVarIDsRanged(start, end int) []exec.VarID {
	var ids []exec.VarID
	for i := start; i < end; i++ {
		v := s.Get(i)
		if v == nil {
			continue
		}
		ids = append(ids, v.VarID)
	}

	return ids
}

type ValueOutputSlots struct {
	Slots VarSlots[ValueVar]
}

func (s *ValueOutputSlots) Len() int {
	return s.Slots.Len()
}

func (s *ValueOutputSlots) Get(idx int) *ValueVar {
	return s.Slots.Get(idx)
}

func (s *ValueOutputSlots) IndexOf(v *ValueVar) int {
	return s.Slots.IndexOf(v)
}

// 设置Slots大小，并为每个Slot创建一个StreamVar
// 调用者应该保证没有正在使用的输出流，即每一个输出流的Dst都为空。
func (s *ValueOutputSlots) Init(my Node, size int) {
	s.Slots.Resize(size)
	for i := 0; i < size; i++ {
		v := my.Graph().NewValueVar()
		v.Src = my
		s.Slots.Set(i, v)
	}
}

// 在Slots末尾增加一个StreamVar，并返回它的索引
func (s *ValueOutputSlots) AppendNew(my Node) ValueSlot {
	v := my.Graph().NewValueVar()
	v.Src = my
	s.Slots.Append(v)
	return ValueSlot{Var: v, Index: s.Len() - 1}
}

// 断开指定位置的输出流到指定节点的连接
func (s *ValueOutputSlots) ClearOutputAt(idx int, dst Node) {
	v := s.Get(idx)
	v.Dst.Remove(dst)
	dst.InputValues().Slots.Clear(v)
}

// 断开所有输出流的所有连接，完全清空所有输出流。但会保留流变量
func (s *ValueOutputSlots) ClearAllOutput(my Node) {
	for i := 0; i < s.Len(); i++ {
		v := s.Get(i)
		v.ClearAllDst()
	}
}

func (s *ValueOutputSlots) GetVarIDs() []exec.VarID {
	var ids []exec.VarID
	for _, v := range s.Slots.RawArray() {
		if v == nil {
			continue
		}
		ids = append(ids, v.VarID)
	}

	return ids
}

func (s *ValueOutputSlots) GetVarIDsRanged(start, end int) []exec.VarID {
	var ids []exec.VarID
	for i := start; i < end; i++ {
		v := s.Get(i)
		if v == nil {
			continue
		}
		ids = append(ids, v.VarID)
	}

	return ids
}

type StreamSlot struct {
	Var   *StreamVar
	Index int
}

type ValueSlot struct {
	Var   *ValueVar
	Index int
}

type NodeBase struct {
	env           NodeEnv
	inputStreams  StreamInputSlots
	outputStreams StreamOutputSlots
	inputValues   ValueInputSlots
	outputValues  ValueOutputSlots
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

func (n *NodeBase) InputStreams() *StreamInputSlots {
	return &n.inputStreams
}

func (n *NodeBase) OutputStreams() *StreamOutputSlots {
	return &n.outputStreams
}

func (n *NodeBase) InputValues() *ValueInputSlots {
	return &n.inputValues
}

func (n *NodeBase) OutputValues() *ValueOutputSlots {
	return &n.outputValues
}
