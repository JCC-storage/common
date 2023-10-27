package internal

import (
	"context"
	"fmt"

	"gitlink.org.cn/cloudream/common/pkgs/actor"
	mylo "gitlink.org.cn/cloudream/common/utils/lo"
	"gitlink.org.cn/cloudream/common/utils/serder"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type LockRequestEvent struct {
	IsLocking bool
	Data      LockRequestData
}

type LockRequestEventWatcher struct {
	OnEvent func(events []LockRequestEvent)
}

type WatchEtcdActor struct {
	etcdCli         *clientv3.Client
	watchChan       clientv3.WatchChan
	lockReqWatchers []*LockRequestEventWatcher

	commandChan *actor.CommandChannel
}

func NewWatchEtcdActor(etcdCli *clientv3.Client) *WatchEtcdActor {
	return &WatchEtcdActor{
		etcdCli:     etcdCli,
		commandChan: actor.NewCommandChannel(),
	}
}

func (a *WatchEtcdActor) Init() {
}

func (a *WatchEtcdActor) StartWatching(revision int64) error {
	return actor.Wait(context.TODO(), a.commandChan, func() error {
		a.watchChan = a.etcdCli.Watch(context.Background(), EtcdLockRequestData, clientv3.WithPrefix(), clientv3.WithPrevKV(), clientv3.WithRev(revision))
		return nil
	})
}

func (a *WatchEtcdActor) StopWatching() error {
	return actor.Wait(context.TODO(), a.commandChan, func() error {
		a.watchChan = nil
		return nil
	})
}

func (a *WatchEtcdActor) AddEventWatcher(watcher *LockRequestEventWatcher) error {
	return actor.Wait(context.TODO(), a.commandChan, func() error {
		a.lockReqWatchers = append(a.lockReqWatchers, watcher)
		return nil
	})
}

func (a *WatchEtcdActor) RemoveEventWatcher(watcher *LockRequestEventWatcher) error {
	return actor.Wait(context.TODO(), a.commandChan, func() error {
		a.lockReqWatchers = mylo.Remove(a.lockReqWatchers, watcher)
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

				events, err := a.parseEvents(msg)
				if err != nil {
					// TODO 更好的错误处理
					return fmt.Errorf("parse etcd lock request data failed, err: %w", err)
				}

				for _, w := range a.lockReqWatchers {
					w.OnEvent(events)
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

func (a *WatchEtcdActor) parseEvents(watchResp clientv3.WatchResponse) ([]LockRequestEvent, error) {
	var events []LockRequestEvent

	for _, e := range watchResp.Events {

		shouldParseData := false
		isLocking := true
		var valueData []byte

		// 只监听新建和删除的事件，因为在设计上约定只有这两种事件才会影响Index
		if e.Type == clientv3.EventTypeDelete {
			shouldParseData = true
			isLocking = false
			valueData = e.PrevKv.Value
		} else if e.IsCreate() {
			shouldParseData = true
			isLocking = true
			valueData = e.Kv.Value
		}

		if !shouldParseData {
			continue
		}

		var reqData LockRequestData
		err := serder.JSONToObject(valueData, &reqData)
		if err != nil {
			return nil, fmt.Errorf("parse lock request data failed, err: %w", err)
		}

		events = append(events, LockRequestEvent{
			IsLocking: isLocking,
			Data:      reqData,
		})
	}

	return events, nil
}
