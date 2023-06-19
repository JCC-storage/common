package distlock

type Config struct {
	EtcdAddress  string `json:"etcdAddress"`
	EtcdUsername string `json:"etcdUsername"`
	EtcdPassword string `json:"etcdPassword"`

	EtcdLockAcquireTimeoutMs int   `json:"etcdLockAcquireTimeoutMs"` // 获取Etcd全局锁的超时时间
	EtcdLockLeaseTimeSec     int64 `json:"etcdLockLeaseTimeSec"`     // 全局锁的租约时间。锁服务会在这个时间内自动续约锁，但如果服务崩溃，则其他服务在租约到期后能重新获得锁。

	// 写入锁请求数据到的ETCD的时候，不设置租约。开启此选项之后，请求锁的服务崩溃，
	// 锁请求数据会依然留在ETCD中。仅供调试使用。
	SubmitLockRequestWithoutLease bool `json:"submitLockRequestWithoutLease"`
}
