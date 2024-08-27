package ops

import (
	"context"
	"fmt"

	"gitlink.org.cn/cloudream/common/pkgs/ioswitch/dag"
	"gitlink.org.cn/cloudream/common/pkgs/ioswitch/exec"
)

type Store struct {
	Var exec.Var
	Key string
}

func (o *Store) Execute(ctx context.Context, e *exec.Executor) error {
	err := e.BindVars(ctx, o.Var)
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

type StoreType struct {
	StoreKey string
}

func (t *StoreType) Store(node *dag.Node, v *dag.ValueVar) {
	v.To(node, 0)
}

func (t *StoreType) InitNode(node *dag.Node) {
	dag.NodeDeclareInputValue(node, 1)
}

func (t *StoreType) GenerateOp(op *dag.Node) (exec.Op, error) {
	return &Store{
		Var: op.InputValues[0].Var,
		Key: t.StoreKey,
	}, nil
}

func (t *StoreType) String(node *dag.Node) string {
	return fmt.Sprintf("Store[%s]%v%v", t.StoreKey, formatStreamIO(node), formatValueIO(node))
}
