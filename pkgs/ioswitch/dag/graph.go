package dag

import (
	"gitlink.org.cn/cloudream/common/utils/lo2"
)

type Graph struct {
	Nodes     []Node
	isWalking bool
	nextVarID int
}

func NewGraph() *Graph {
	return &Graph{}
}

func (g *Graph) AddNode(node Node) {
	g.Nodes = append(g.Nodes, node)
	node.SetGraph(g)
}

func (g *Graph) RemoveNode(node Node) {
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

func (g *Graph) Walk(cb func(node Node) bool) {
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

func (g *Graph) NewStreamVar() *StreamVar {
	str := &StreamVar{
		VarBase: VarBase{
			id: g.genVarID(),
		},
	}
	return str
}

func (g *Graph) NewValueVar(valType ValueVarType) *ValueVar {
	val := &ValueVar{
		VarBase: VarBase{
			id: g.genVarID(),
		},
		Type: valType,
	}
	return val
}

func (g *Graph) genVarID() int {
	g.nextVarID++
	return g.nextVarID
}

func AddNode[N Node](graph *Graph, typ N) N {
	graph.AddNode(typ)
	return typ
}

func WalkOnlyType[N Node](g *Graph, cb func(node N) bool) {
	g.Walk(func(n Node) bool {
		node, ok := n.(N)
		if ok {
			return cb(node)
		}
		return true
	})
}
