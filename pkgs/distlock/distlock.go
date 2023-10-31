package distlock

import (
	"fmt"

	"gitlink.org.cn/cloudream/common/pkgs/distlock/internal"
)

type Lock = internal.Lock

type LockRequest = internal.LockRequest

type LockProvider = internal.LockProvider

type Config = internal.Config

type LockTargetBusyError struct {
	lockName string
}

func (e *LockTargetBusyError) Error() string {
	return fmt.Sprintf("the lock object is locked by %s", e.lockName)
}

func NewLockTargetBusyError(lockName string) *LockTargetBusyError {
	return &LockTargetBusyError{
		lockName: lockName,
	}
}
