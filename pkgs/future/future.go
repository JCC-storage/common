package future

import (
	"fmt"
	"time"
)

var ErrWaitTimeout = fmt.Errorf("wait timeout")

type Future interface {
	Error() error
	IsComplete() bool

	Wait() error
	WaitTimeout(timeout time.Duration) error
}

type ValueFuture[T any] interface {
	Future

	Value() T

	WaitValue() (T, error)
	WaitValueTimeout(timeout time.Duration) (T, error)
}
