package internal

import (
	"context"
	"fmt"
	"strings"

	"gitlink.org.cn/cloudream/common/pkgs/actor"
	"gitlink.org.cn/cloudream/common/utils/serder"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type LockRequestEvent struct {
	IsLocking bool
	Data      LockRequestData
}

type ServiceEvent struct {
	IsNew bool
	Info  ServiceInfo
}

type OnLockRequestEventFn func(event LockRequestEvent)

type OnServiceEventFn func(event ServiceEvent)

type WatchEtcdActor struct {
	etcdCli *clientv3.Client

	watchChan            clientv3.WatchChan
	watchChanCancel      func()
	onLockRequestEventFn OnLockRequestEventFn
	onServiceEventFn     OnServiceEventFn
	commandChan          *actor.CommandChannel
}

func NewWatchEtcdActor(etcdCli *clientv3.Client) *WatchEtcdActor {
	return &WatchEtcdActor{
		etcdCli:     etcdCli,
		commandChan: actor.NewCommandChannel(),
	}
}

func (a *WatchEtcdActor) Init(onLockRequestEvent OnLockRequestEventFn, onServiceDown OnServiceEventFn) {
	a.onLockRequestEventFn = onLockRequestEvent
	a.onServiceEventFn = onServiceDown
}

func (a *WatchEtcdActor) Start(revision int64) {
	actor.Wait(context.Background(), a.commandChan, func() error {
		if a.watchChanCancel != nil {
			a.watchChanCancel()
			a.watchChanCancel = nil
		}

		ctx, cancel := context.WithCancel(context.Background())
		a.watchChan = a.etcdCli.Watch(ctx, EtcdWatchPrefix, clientv3.WithPrefix(), clientv3.WithPrevKV(), clientv3.WithRev(revision))
		a.watchChanCancel = cancel
		return nil
	})
}

func (a *WatchEtcdActor) Stop() {
	actor.Wait(context.Background(), a.commandChan, func() error {
		if a.watchChanCancel != nil {
			a.watchChanCancel()
			a.watchChanCancel = nil
		}
		a.watchChan = nil
		return nil
	})
}

func (a *WatchEtcdActor) Serve() error {
	cmdChan := a.commandChan.BeginChanReceive()
	defer a.commandChan.CloseChanReceive()

	for {
		if a.watchChan != nil {
			select {
			case cmd, ok := <-cmdChan:
				if !ok {
					return fmt.Errorf("command channel closed")
				}

				cmd()

			case msg := <-a.watchChan:
				if msg.Canceled {
					// TODO 更好的错误处理
					return fmt.Errorf("watch etcd channel closed")
				}

				err := a.dispatchEtcdEvent(msg)
				if err != nil {
					// TODO 更好的错误处理
					return err
				}
			}

		} else {
			select {
			case cmd, ok := <-cmdChan:
				if !ok {
					return fmt.Errorf("command channel closed")
				}

				cmd()
			}
		}
	}
}

func (a *WatchEtcdActor) dispatchEtcdEvent(watchResp clientv3.WatchResponse) error {
	for _, e := range watchResp.Events {
		key := string(e.Kv.Key)

		if strings.HasPrefix(key, EtcdLockRequestDataPrefix) {
			if err := a.applyLockRequestEvent(e); err != nil {
				return fmt.Errorf("parsing lock request event: %w", err)
			}

		} else if strings.HasPrefix(key, EtcdServiceInfoPrefix) {
			if err := a.applyServiceEvent(e); err != nil {
				return fmt.Errorf("parsing service event: %w", err)
			}
		}
	}

	return nil
}

func (a *WatchEtcdActor) applyLockRequestEvent(evt *clientv3.Event) error {
	isLocking := true
	var valueData []byte

	// 只监听新建和删除的事件，因为在设计上约定只有这两种事件才会影响Index
	if evt.Type == clientv3.EventTypeDelete {
		isLocking = false
		valueData = evt.PrevKv.Value
	} else if evt.IsCreate() {
		isLocking = true
		valueData = evt.Kv.Value
	} else {
		return nil
	}

	var reqData LockRequestData
	err := serder.JSONToObject(valueData, &reqData)
	if err != nil {
		return fmt.Errorf("parse lock request data failed, err: %w", err)
	}

	a.onLockRequestEventFn(LockRequestEvent{
		IsLocking: isLocking,
		Data:      reqData,
	})

	return nil
}

func (a *WatchEtcdActor) applyServiceEvent(evt *clientv3.Event) error {
	isNew := true
	var valueData []byte

	// 只监听新建和删除的事件，因为在设计上约定只有这两种事件才会影响Index
	if evt.Type == clientv3.EventTypeDelete {
		isNew = false
		valueData = evt.PrevKv.Value
	} else if evt.IsCreate() {
		isNew = true
		valueData = evt.Kv.Value
	} else {
		return nil
	}

	var svcInfo ServiceInfo
	err := serder.JSONToObject(valueData, &svcInfo)
	if err != nil {
		return fmt.Errorf("parsing service info: %w", err)
	}

	a.onServiceEventFn(ServiceEvent{
		IsNew: isNew,
		Info:  svcInfo,
	})

	return nil
}
