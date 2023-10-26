package distlock

type Config struct {
	EtcdAddress  string `json:"etcdAddress"`
	EtcdUsername string `json:"etcdUsername"`
	EtcdPassword string `json:"etcdPassword"`

	EtcdLockLeaseTimeSec   int64 `json:"etcdLockLeaseTimeSec"`   // 全局锁的租约时间。锁服务会在这个时间内自动续约锁，但如果服务崩溃，则其他服务在租约到期后能重新获得锁。
	RandomReleasingDelayMs int64 `json:"randomReleasingDelayMs"` // 释放锁失败，随机延迟之后再次尝试。延迟时间=random(0, RandomReleasingDelayMs) + 最少延迟时间(1000ms)
}
