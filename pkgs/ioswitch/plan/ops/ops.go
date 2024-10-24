package ops

import "gitlink.org.cn/cloudream/common/pkgs/ioswitch/dag"

type GraphNodeBuilder struct {
	*dag.Graph
}

func NewGraphNodeBuilder() *GraphNodeBuilder {
	return &GraphNodeBuilder{
		Graph: dag.NewGraph(),
	}
}
