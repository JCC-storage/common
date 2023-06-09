package distlock

type Config struct {
	EtcdAddress  string `json:"etcdAddress"`
	EtcdUsername string `json:"etcdUsername"`
	EtcdPassword string `json:"etcdPassword"`

	EtcdLockAcquireTimeoutMs int   `json:"etcdLockAcquireTimeoutMs"` // 获取Etcd全局锁的超时时间
	EtcdLockLeaseTimeSec     int64 `json:"etcdLockLeaseTimeSec"`     // 全局锁的租约时间。锁服务会在这个时间内自动续约锁，但如果服务崩溃，则其他服务在租约到期后能重新获得锁。

	LockRequestLeaseTimeSec int64 `json:"lockRequestLeaseTimeSec"` // 锁请求的租约时间。调用方必须在这个时间内调用Renew续约。
}
