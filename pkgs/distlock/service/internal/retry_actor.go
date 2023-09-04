package internal

import (
	"context"
	"fmt"

	"github.com/samber/lo"
	"gitlink.org.cn/cloudream/common/pkgs/actor"
	"gitlink.org.cn/cloudream/common/pkgs/distlock"
	"gitlink.org.cn/cloudream/common/pkgs/future"
	"gitlink.org.cn/cloudream/common/pkgs/logger"
	mylo "gitlink.org.cn/cloudream/common/utils/lo"
)

type retryInfo struct {
	Callback *future.SetValueFuture[string]
	LastErr  error
}

type RetryActor struct {
	retrys     []distlock.LockRequest
	retryInfos []*retryInfo

	commandChan *actor.CommandChannel

	mainActor *MainActor
}

func NewRetryActor() *RetryActor {
	return &RetryActor{
		commandChan: actor.NewCommandChannel(),
	}
}

func (a *RetryActor) Init(mainActor *MainActor) {
	a.mainActor = mainActor
}

func (a *RetryActor) Retry(ctx context.Context, req distlock.LockRequest, lastErr error) (future.ValueFuture[string], error) {
	fut := future.NewSetValue[string]()

	var info *retryInfo
	err := actor.Wait(ctx, a.commandChan, func() error {
		a.retrys = append(a.retrys, req)
		info = &retryInfo{
			Callback: fut,
			LastErr:  lastErr,
		}
		a.retryInfos = append(a.retryInfos, info)
		return nil
	})
	if err != nil {
		return nil, err
	}

	go func() {
		<-ctx.Done()
		a.commandChan.Send(func() {
			// 由于只可能在cmd中修改future状态，所以此处的IsComplete判断可以作为后续操作的依据
			if fut.IsComplete() {
				return
			}

			index := lo.IndexOf(a.retryInfos, info)
			if index == -1 {
				return
			}

			a.retryInfos[index].Callback.SetError(a.retryInfos[index].LastErr)

			mylo.RemoveAt(a.retrys, index)
			mylo.RemoveAt(a.retryInfos, index)
		})
	}()

	return fut, nil
}

func (a *RetryActor) OnLocalStateUpdated() {
	a.commandChan.Send(func() {
		if len(a.retrys) == 0 {
			return
		}

		rets, err := a.mainActor.AcquireMany(context.Background(), a.retrys)
		if err != nil {
			// TODO 处理错误
			logger.Std.Warnf("acquire many lock requests failed, err: %s", err.Error())
			return
		}

		// 根据尝试的结果更新状态
		delCnt := 0
		for i, ret := range rets {
			a.retrys[i-delCnt] = a.retrys[i]
			a.retryInfos[i-delCnt] = a.retryInfos[i]

			if !ret.IsTried {
				continue
			}

			if ret.Err != nil {
				a.retryInfos[i].LastErr = ret.Err
			} else {
				a.retryInfos[i].Callback.SetValue(ret.RequestID)
				delCnt++
			}
		}
		a.retrys = a.retrys[:len(a.retrys)-delCnt]
		a.retryInfos = a.retryInfos[:len(a.retryInfos)-delCnt]
	})
}

func (a *RetryActor) Serve() error {
	cmdChan := a.commandChan.BeginChanReceive()
	defer a.commandChan.CloseChanReceive()

	for {
		select {
		case cmd, ok := <-cmdChan:
			if !ok {
				return fmt.Errorf("command channel closed")
			}

			cmd()
		}
	}
}
