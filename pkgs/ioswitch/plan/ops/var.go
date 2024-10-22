package ops

import (
	"gitlink.org.cn/cloudream/common/pkgs/ioswitch/exec"
)

func init() {
	exec.UseOp[*ConstVar]()
}

type ConstVar struct {
	Var *exec.StringVar `json:"var"`
}

func (o *ConstVar) Execute(ctx *exec.ExecContext, e *exec.Executor) error {
	e.PutVars(o.Var)
	return nil
}

func (o *ConstVar) String() string {
	return "ConstVar"
}
