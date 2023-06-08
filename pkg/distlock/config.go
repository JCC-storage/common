package distlock

type Config struct {
	EtcdAddress  string `json:"etcdAddress"`
	EtcdUsername string `json:"etcdUsername"`
	EtcdPassword string `json:"etcdPassword"`

	LockRequestDataConfig LockRequestDataConfig `json:"lockRequestDataConfig"`
}

type LockRequestDataConfig struct {
	AcquireTimeoutMs int   `json:"acquireTimeoutMs"` // 获取Etcd全局锁的超时时间
	LeaseTimeSec     int64 `json:"leaseTimeSec"`     // 全局锁的租约时间。锁服务会在这个时间内自动续约锁，但如果服务崩溃，则其他服务在租约到期后能重新获得锁。
}
