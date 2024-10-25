package ops

import (
	"gitlink.org.cn/cloudream/common/pkgs/ioswitch/exec"
)

func init() {
	exec.UseOp[*ConstVar]()
}

type ConstVar struct {
	ID    exec.VarID    `json:"id"`
	Value exec.VarValue `json:"value"`
}

func (o *ConstVar) Execute(ctx *exec.ExecContext, e *exec.Executor) error {
	e.PutVar(o.ID, o.Value)
	return nil
}

func (o *ConstVar) String() string {
	return "ConstVar"
}
