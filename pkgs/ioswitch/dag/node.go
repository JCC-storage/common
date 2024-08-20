package dag

import (
	"fmt"

	"gitlink.org.cn/cloudream/common/pkgs/ioswitch/exec"
)

type NodeType interface {
	InitNode(node *Node)
	String(node *Node) string
	GenerateOp(node *Node) (exec.Op, error)
}

type NodeEnvType string

const (
	EnvUnknown NodeEnvType = ""
	EnvDriver  NodeEnvType = "Driver"
	EnvWorker  NodeEnvType = "Worker"
)

type NodeEnv struct {
	Type   NodeEnvType
	Worker exec.WorkerInfo
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

func (e *NodeEnv) Equals(other NodeEnv) bool {
	if e.Type != other.Type {
		return false
	}

	if e.Type != EnvWorker {
		return true
	}

	return e.Worker.Equals(other.Worker)
}

type Node struct {
	Type          NodeType
	Env           NodeEnv
	Props         any
	InputStreams  []*StreamVar
	OutputStreams []*StreamVar
	InputValues   []*ValueVar
	OutputValues  []*ValueVar
	Graph         *Graph
}

func (n *Node) String() string {
	return fmt.Sprintf("%v", n.Type.String(n))
}
