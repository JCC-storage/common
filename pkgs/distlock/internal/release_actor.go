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

	lock                    sync.Mutex
	isMaintenance           bool
	releasingLockRequestIDs map[string]bool
	timer                   *time.Timer
	timerSetup              bool
	doReleasingChan         chan any
}

func NewReleaseActor(cfg *Config, etcdCli *clientv3.Client) *ReleaseActor {
	return &ReleaseActor{
		cfg:                     cfg,
		etcdCli:                 etcdCli,
		isMaintenance:           true,
		releasingLockRequestIDs: make(map[string]bool),
		doReleasingChan:         make(chan any, 1),
	}
}

// 立刻尝试释放这些锁。一般用于在用户主动释放了一个锁之后
func (a *ReleaseActor) Release(reqIDs []string) {
	a.lock.Lock()
	defer a.lock.Unlock()

	for _, id := range reqIDs {
		a.releasingLockRequestIDs[id] = true
	}

	if a.isMaintenance {
		return
	}

	select {
	case a.doReleasingChan <- nil:
	default:
	}
}

// 延迟释放锁。一般用于清理崩溃的锁服务遗留下来的锁
func (a *ReleaseActor) DelayRelease(reqIDs []string) {
	a.lock.Lock()
	defer a.lock.Unlock()

	for _, id := range reqIDs {
		a.releasingLockRequestIDs[id] = true
	}

	if a.isMaintenance {
		return
	}

	a.setupTimer()
}

// 重试一下内部的解锁请求。不会阻塞调用者
func (a *ReleaseActor) TryReleaseNow() {
	a.lock.Lock()
	defer a.lock.Unlock()

	// 如果处于维护模式，那么即使主动进行释放操作，也不予理会
	if a.isMaintenance {
		return
	}

	select {
	case a.doReleasingChan <- nil:
	default:
	}
}

// 进入维护模式。在维护模式期间只接受请求，不处理请求，包括延迟释放请求。
func (a *ReleaseActor) EnterMaintenance() {
	a.lock.Lock()
	defer a.lock.Unlock()

	a.isMaintenance = true
}

// 退出维护模式。退出之后建议调用一下TryReleaseNow。
func (a *ReleaseActor) LeaveMaintenance() {
	a.lock.Lock()
	defer a.lock.Unlock()

	a.isMaintenance = false
}

func (a *ReleaseActor) OnLockRequestEvent(event LockRequestEvent) {
	if event.IsLocking {
		return
	}

	a.lock.Lock()
	defer a.lock.Unlock()

	delete(a.releasingLockRequestIDs, event.Data.ID)
}

func (a *ReleaseActor) Serve() {
	for {
		// 与Acquire不同，解锁操作不需要进行互斥判断，而且能一次性解锁多个，
		// 所以此处也能保证新提交的解锁请求都会被尝试后再进入等待。
		select {
		case <-a.doReleasingChan:
		}

		// 先看一眼，如果没有需要释放的锁，就重新进入等待状态
		a.lock.Lock()
		if len(a.releasingLockRequestIDs) == 0 {
			a.lock.Unlock()
			continue
		}
		a.lock.Unlock()

		err := a.doReleasing()
		if err != nil {
			logger.Std.Debugf("doing releasing: %s", err.Error())
		}
	}
}

func (a *ReleaseActor) doReleasing() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	// 在获取全局锁的时候不用锁Actor，只有获取成功了，才加锁
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

	a.lock.Lock()
	defer a.lock.Unlock()
	defer a.setupTimer()

	// TODO 可以考虑优化成一次性删除多个锁
	for id := range a.releasingLockRequestIDs {
		lockReqKey := MakeEtcdLockRequestKey(id)

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

	if a.timerSetup {
		return
	}
	a.timerSetup = true

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

		a.lock.Lock()
		defer a.lock.Unlock()

		a.timerSetup = false

		// 如果处于维护模式，那么即使是定时器要求的释放操作，也不予理会
		if a.isMaintenance {
			return
		}

		select {
		case a.doReleasingChan <- nil:
		default:
		}
	}()
}
