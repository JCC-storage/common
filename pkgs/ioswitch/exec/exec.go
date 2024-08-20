package exec

import (
	"context"

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
	Execute(ctx context.Context, e *Executor) error
}

func UseOp[T Op]() {
	opUnion.Add(reflect2.TypeOf[T]())
}
