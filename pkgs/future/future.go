package future

import (
	"context"
	"fmt"
)

var ErrContextCancelled = fmt.Errorf("context cancelled")

type Future interface {
	Error() error
	IsComplete() bool

	Wait(ctx context.Context) error
}

type ValueFuture[T any] interface {
	Future

	Value() T

	WaitValue(ctx context.Context) (T, error)
}
