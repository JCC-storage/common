package distlock

import "fmt"

type Lock struct {
	Path   []string // 锁路径，存储的是路径的每一部分
	Name   string   // 锁名
	Target any      // 锁对象，由具体的Provider去解析
}

type LockRequest struct {
	Locks []Lock
}

func (b *LockRequest) Add(lock Lock) {
	b.Locks = append(b.Locks, lock)
}

type LockProvider interface {
	// CanLock 判断这个锁能否锁定成功
	CanLock(lock Lock) error

	// 锁定。在内部可以不用判断能否加锁，外部需要保证调用此函数前调用了CanLock进行检查
	Lock(reqID string, lock Lock) error

	// 解锁
	Unlock(reqID string, lock Lock) error

	// GetTargetString 将锁对象序列化为字符串，方便存储到ETCD
	GetTargetString(target any) (string, error)

	// ParseTargetString 解析字符串格式的锁对象数据
	ParseTargetString(targetStr string) (any, error)

	// Clear 清除内部所有状态
	Clear()
}

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
