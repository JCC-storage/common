package future

import (
	"context"
	"fmt"
)

var ErrContextCancelled = fmt.Errorf("context cancelled")
var ErrCompleted = fmt.Errorf("context cancelled")

type Future interface {
	IsComplete() bool

	Chan() <-chan error

	Wait(ctx context.Context) error
}

type ChanValue1[T any] struct {
	Value T
	Err   error
}

type ChanValue2[T1 any, T2 any] struct {
	Value1 T1
	Value2 T2
	Err    error
}

type Future1[T any] interface {
	IsComplete() bool

	Chan() <-chan ChanValue1[T]

	Wait(ctx context.Context) (T, error)
}
