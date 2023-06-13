package internal

import (
	"context"
	"fmt"

	"gitlink.org.cn/cloudream/common/pkg/actor"
	"gitlink.org.cn/cloudream/common/utils/serder"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type WatchEtcdActor struct {
	etcdCli   *clientv3.Client
	watchChan clientv3.WatchChan

	commandChan *actor.CommandChannel

	providers *ProvidersActor
}

func NewWatchEtcdActor() *WatchEtcdActor {
	return &WatchEtcdActor{
		commandChan: actor.NewCommandChannel(),
	}
}

func (a *WatchEtcdActor) Init(providers *ProvidersActor) {
	a.providers = providers
}

func (a *WatchEtcdActor) StartWatching() error {
	return actor.Wait(a.commandChan, func() error {
		a.watchChan = a.etcdCli.Watch(context.Background(), LOCK_REQUEST_DATA_PREFIX, clientv3.WithPrefix())
		return nil
	})
}

func (a *WatchEtcdActor) StopWatching() error {
	return actor.Wait(a.commandChan, func() error {
		a.watchChan = nil
		return nil
	})
}

func (a *WatchEtcdActor) Serve() error {
	for {
		if a.watchChan != nil {
			select {
			case cmd, ok := <-a.commandChan.ChanReceive():
				if !ok {
					return fmt.Errorf("command channel closed")
				}

				cmd()

			case msg := <-a.watchChan:
				if msg.Canceled {
					// TODO 更好的错误处理
					return fmt.Errorf("watch etcd channel closed")
				}

				ops, err := a.parseEvents(msg)
				if err != nil {
					// TODO 更好的错误处理
					return fmt.Errorf("parse etcd lock request data failed, err: %w", err)
				}

				err = a.providers.BatchUpdateByLockRequestData(ops)
				if err != nil {
					// TODO 更好的错误处理
					return fmt.Errorf("update local lock request data failed, err: %w", err)
				}
			}

		} else {
			select {
			case cmd, ok := <-a.commandChan.ChanReceive():
				if !ok {
					return fmt.Errorf("command channel closed")
				}

				cmd()
			}
		}
	}
}

func (a *WatchEtcdActor) parseEvents(watchResp clientv3.WatchResponse) ([]lockRequestDataUpdateOp, error) {
	var ops []lockRequestDataUpdateOp

	for _, e := range watchResp.Events {

		shouldParseData := false
		isLock := true

		// 只监听新建和删除的事件，因为在设计上约定只有这两种事件才会影响Index
		if e.Type == clientv3.EventTypeDelete {
			shouldParseData = true
			isLock = false
		} else if e.IsCreate() {
			shouldParseData = true
			isLock = true
		}

		if !shouldParseData {
			continue
		}

		var reqData lockRequestData
		err := serder.JSONToObject(e.Kv.Value, &reqData)
		if err != nil {
			return nil, fmt.Errorf("parse lock request data failed, err: %w", err)
		}

		ops = append(ops, lockRequestDataUpdateOp{
			IsLock: isLock,
			Data:   reqData,
		})
	}

	return ops, nil
}
