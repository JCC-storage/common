package exec

import (
	"gitlink.org.cn/cloudream/common/pkgs/types"
	"gitlink.org.cn/cloudream/common/utils/reflect2"
	"gitlink.org.cn/cloudream/common/utils/serder"
)

type PlanID string

type Plan struct {
	ID  PlanID `json:"id"`
	Ops []Op   `json:"ops"`
}

var opUnion = serder.UseTypeUnionExternallyTagged(types.Ref(types.NewTypeUnion[Op]()))

type Op interface {
	Execute(ctx *ExecContext, e *Executor) error
	String() string
}

func UseOp[T Op]() {
	opUnion.Add(reflect2.TypeOf[T]())
}
