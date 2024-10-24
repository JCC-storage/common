package ops

import (
	"fmt"

	"gitlink.org.cn/cloudream/common/pkgs/ioswitch/dag"
	"gitlink.org.cn/cloudream/common/pkgs/ioswitch/exec"
)

type Store struct {
	Var exec.Var
	Key string
}

func (o *Store) Execute(ctx *exec.ExecContext, e *exec.Executor) error {
	err := e.BindVars(ctx.Context, o.Var)
	if err != nil {
		return err
	}

	switch v := o.Var.(type) {
	case *exec.IntVar:
		e.Store(o.Key, v.Value)
	case *exec.StringVar:
		e.Store(o.Key, v.Value)
	}

	return nil
}

func (o *Store) String() string {
	return fmt.Sprintf("Store %v: %v", o.Key, o.Var.GetID())
}

type StoreNode struct {
	dag.NodeBase
	Key string
}

func (b *GraphNodeBuilder) NewStore() *StoreNode {
	node := &StoreNode{}
	b.AddNode(node)
	return node
}

func (t *StoreNode) Store(key string, v *dag.ValueVar) {
	t.Key = key
	t.InputValues().EnsureSize(1)
	v.Connect(t, 0)
}

func (t *StoreNode) GenerateOp() (exec.Op, error) {
	return &Store{
		Var: t.InputValues().Get(0).Var,
		Key: t.Key,
	}, nil
}

// func (t *StoreType) String() string {
// 	return fmt.Sprintf("Store[%s]%v%v", t.StoreKey, formatStreamIO(node), formatValueIO(node))
// }
