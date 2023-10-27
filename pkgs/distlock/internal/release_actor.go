package internal

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"gitlink.org.cn/cloudream/common/pkgs/logger"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/clientv3util"
)

const (
	DefaultMaxReleaseingDelayMs = 4000
	BaseReleaseingDelayMs       = 1000
)

type ReleaseActor struct {
	cfg     *Config
	etcdCli *clientv3.Client

	releasingLockRequestIDs map[string]bool
	timer                   *time.Timer
	timerSetuped            bool
	lock                    sync.Mutex
}

func NewReleaseActor(cfg *Config, etcdCli *clientv3.Client) *ReleaseActor {
	return &ReleaseActor{
		cfg:                     cfg,
		etcdCli:                 etcdCli,
		releasingLockRequestIDs: make(map[string]bool),
	}
}

// 立刻尝试释放这些锁。一般用于在用户主动释放了一个锁之后
func (a *ReleaseActor) Release(reqIDs []string) {
	a.lock.Lock()
	defer a.lock.Unlock()

	for _, id := range reqIDs {
		a.releasingLockRequestIDs[id] = true
	}

	// TODO 处理错误
	err := a.doReleasing()
	if err != nil {
		logger.Std.Debugf("doing releasing: %s", err.Error())
	}

	a.setupTimer()
}

// 延迟释放锁。一般用于清理崩溃的锁服务遗留下来的锁
func (a *ReleaseActor) DelayRelease(reqIDs []string) {
	a.lock.Lock()
	defer a.lock.Unlock()

	for _, id := range reqIDs {
		a.releasingLockRequestIDs[id] = true
	}

	a.setupTimer()
}

func (a *ReleaseActor) doReleasing() error {
	ctx := context.TODO()

	if len(a.releasingLockRequestIDs) == 0 {
		return nil
	}

	// TODO 根据不同的错误设置不同的错误类型，方便上层进行后续处理
	unlock, err := acquireEtcdRequestDataLock(ctx, a.etcdCli, a.cfg.EtcdLockLeaseTimeSec)
	if err != nil {
		return fmt.Errorf("acquire etcd request data lock failed, err: %w", err)
	}
	defer unlock()

	index, err := getEtcdLockRequestIndex(ctx, a.etcdCli)
	if err != nil {
		return err
	}

	// TODO 可以考虑优化成一次性删除多个锁
	for id := range a.releasingLockRequestIDs {
		lockReqKey := makeEtcdLockRequestKey(id)

		txResp, err := a.etcdCli.Txn(ctx).
			If(clientv3util.KeyExists(lockReqKey)).
			Then(clientv3.OpDelete(lockReqKey), clientv3.OpPut(EtcdLockRequestIndex, strconv.FormatInt(index+1, 10))).Commit()
		if err != nil {
			return fmt.Errorf("updating lock request data: %w", err)
		}
		// 只有确实删除了锁数据，才更新index
		if txResp.Succeeded {
			index++
		}
		delete(a.releasingLockRequestIDs, id)
	}

	return nil
}

func (a *ReleaseActor) setupTimer() {
	if len(a.releasingLockRequestIDs) == 0 {
		return
	}

	if a.timerSetuped {
		return
	}
	a.timerSetuped = true

	delay := int64(0)
	if a.cfg.RandomReleasingDelayMs == 0 {
		delay = rand.Int63n(DefaultMaxReleaseingDelayMs)
	} else {
		delay = rand.Int63n(a.cfg.RandomReleasingDelayMs)
	}

	if a.timer == nil {
		a.timer = time.NewTimer(time.Duration(delay+BaseReleaseingDelayMs) * time.Millisecond)
	} else {
		a.timer.Reset(time.Duration(delay+BaseReleaseingDelayMs) * time.Millisecond)
	}

	go func() {
		<-a.timer.C
		a.timerSetuped = false

		// TODO 处理错误
		err := a.doReleasing()
		if err != nil {
			logger.Std.Debugf("doing releasing: %s", err.Error())
		}

		a.setupTimer()
	}()
}
