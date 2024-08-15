package dag

import (
	"fmt"

	"gitlink.org.cn/cloudream/common/pkgs/ioswitch/exec"
)

type NodeType[NP any, VP any] interface {
	InitNode(node *Node[NP, VP])
	String(node *Node[NP, VP]) string
	GenerateOp(node *Node[NP, VP], blder *exec.PlanBuilder) error
}

type WorkerInfo interface {
	// 获取连接到这个worker的GRPC服务的地址
	GetAddress() string
	// 判断两个worker是否相同
	Equals(worker WorkerInfo) bool
}

type NodeEnvType string

const (
	EnvUnknown  NodeEnvType = ""
	EnvExecutor NodeEnvType = "Executor"
	EnvWorker   NodeEnvType = "Worker"
)

type NodeEnv struct {
	Type   NodeEnvType
	Worker WorkerInfo
}

func (e *NodeEnv) ToEnvUnknown() {
	e.Type = EnvUnknown
	e.Worker = nil
}

func (e *NodeEnv) ToEnvExecutor() {
	e.Type = EnvExecutor
	e.Worker = nil
}

func (e *NodeEnv) ToEnvWorker(worker WorkerInfo) {
	e.Type = EnvWorker
	e.Worker = worker
}

func (e *NodeEnv) Equals(other NodeEnv) bool {
	if e.Type != other.Type {
		return false
	}

	if e.Type != EnvWorker {
		return true
	}

	return e.Worker.Equals(other.Worker)
}

type Node[NP any, VP any] struct {
	Type          NodeType[NP, VP]
	Env           NodeEnv
	Props         NP
	InputStreams  []*StreamVar[NP, VP]
	OutputStreams []*StreamVar[NP, VP]
	InputValues   []*ValueVar[NP, VP]
	OutputValues  []*ValueVar[NP, VP]
	Graph         *Graph[NP, VP]
}

func (n *Node[NP, VP]) String() string {
	return fmt.Sprintf("%v", n.Type.String(n))
}
