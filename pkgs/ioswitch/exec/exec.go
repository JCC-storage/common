package exec

import (
	"context"
)

type PlanID string

type Plan struct {
	ID  PlanID `json:"id"`
	Ops []Op   `json:"ops"`
}

type Op interface {
	Execute(ctx context.Context, sw *Executor) error
}
