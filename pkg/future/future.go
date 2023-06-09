package future

import (
	"fmt"
	"time"
)

var ErrWaitTimeout = fmt.Errorf("wait timeout")

type Future[T any] interface {
	IsComplete() bool

	Wait() (T, error)
	WaitTimeout(timeout time.Duration) (T, error)
}
