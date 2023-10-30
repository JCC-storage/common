package internal

import "strings"

const (
	EtcdLockRequestDataPrefix = "/distlock/lockRequest/data"
	EtcdLockRequestIndex      = "/distlock/lockRequest/index"
	EtcdLockRequestLock       = "/distlock/lockRequest/lock"
	EtcdServiceInfoPrefix     = "/distlock/services"
	EtcdWatchPrefix           = "/distlock"
)

type Lock struct {
	Path   []string // 锁路径，存储的是路径的每一部分
	Name   string   // 锁名
	Target any      // 锁对象，由具体的Provider去解析
}

type LockRequest struct {
	Reason string
	Locks  []Lock
}

func (b *LockRequest) Add(lock Lock) {
	b.Locks = append(b.Locks, lock)
}

type LockProvider interface {
	// CanLock 判断这个锁能否锁定成功
	CanLock(lock Lock) error

	// Lock 锁定。由于同一个锁请求内的锁不检查冲突，因此这个函数必须支持有冲突的锁进行锁定。
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

type lockData struct {
	Path   []string `json:"path"`
	Name   string   `json:"name"`
	Target string   `json:"target"`
}

type LockRequestData struct {
	ID        string     `json:"id"`
	SerivceID string     `json:"serviceID"`
	Reason    string     `json:"reason"`
	Timestamp int64      `json:"timestamp"`
	Locks     []lockData `json:"locks"`
}

func MakeEtcdLockRequestKey(reqID string) string {
	return EtcdLockRequestDataPrefix + "/" + reqID
}

func GetLockRequestID(key string) string {
	return strings.TrimPrefix(key, EtcdLockRequestDataPrefix+"/")
}

func MakeServiceInfoKey(svcID string) string {
	return EtcdServiceInfoPrefix + "/" + svcID
}

type ServiceInfo struct {
	ID string `json:"id"`
}
