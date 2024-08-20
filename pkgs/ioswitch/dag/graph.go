package dag

import (
	"gitlink.org.cn/cloudream/common/utils/lo2"
)

type Graph struct {
	Nodes     []*Node
	isWalking bool
	nextVarID int
}

func NewGraph() *Graph {
	return &Graph{}
}

func (g *Graph) NewNode(typ NodeType, props any) *Node {
	n := &Node{
		Type:  typ,
		Props: props,
		Graph: g,
	}
	typ.InitNode(n)
	g.Nodes = append(g.Nodes, n)
	return n
}

func (g *Graph) RemoveNode(node *Node) {
	for i, n := range g.Nodes {
		if n == node {
			if g.isWalking {
				g.Nodes[i] = nil
			} else {
				g.Nodes = lo2.RemoveAt(g.Nodes, i)
			}
			break
		}
	}
}

func (g *Graph) Walk(cb func(node *Node) bool) {
	g.isWalking = true
	for i := 0; i < len(g.Nodes); i++ {
		if g.Nodes[i] == nil {
			continue
		}

		if !cb(g.Nodes[i]) {
			break
		}
	}
	g.isWalking = false

	g.Nodes = lo2.RemoveAllDefault(g.Nodes)
}

func (g *Graph) genVarID() int {
	g.nextVarID++
	return g.nextVarID
}

func NewNode[N NodeType](graph *Graph, typ N, props any) (*Node, N) {
	return graph.NewNode(typ, props), typ
}

func WalkOnlyType[N NodeType](g *Graph, cb func(node *Node, typ N) bool) {
	g.Walk(func(node *Node) bool {
		typ, ok := node.Type.(N)
		if ok {
			return cb(node, typ)
		}
		return true
	})
}
