package internal

import (
	"context"
	"fmt"
	"time"

	"gitlink.org.cn/cloudream/common/pkgs/actor"
	"gitlink.org.cn/cloudream/common/pkgs/logger"
)

type lockRequestLease struct {
	RequestID string
	LeaseTime time.Duration
	Deadline  time.Time
}

type LeaseActor struct {
	leases map[string]*lockRequestLease
	ticker *time.Ticker

	commandChan *actor.CommandChannel

	releaseActor *ReleaseActor
}

func NewLeaseActor() *LeaseActor {
	return &LeaseActor{
		leases:      make(map[string]*lockRequestLease),
		commandChan: actor.NewCommandChannel(),
	}
}

func (a *LeaseActor) Init(releaseActor *ReleaseActor) {
	a.releaseActor = releaseActor
}

func (a *LeaseActor) Start() error {
	return actor.Wait(context.TODO(), a.commandChan, func() error {
		a.ticker = time.NewTicker(time.Second)
		return nil
	})
}

func (a *LeaseActor) Stop() error {
	return actor.Wait(context.TODO(), a.commandChan, func() error {
		if a.ticker != nil {
			a.ticker.Stop()
		}
		a.ticker = nil
		return nil
	})
}

func (a *LeaseActor) Add(reqID string, leaseTime time.Duration) error {
	return actor.Wait(context.TODO(), a.commandChan, func() error {
		lease, ok := a.leases[reqID]
		if !ok {
			lease = &lockRequestLease{
				RequestID: reqID,
				LeaseTime: leaseTime,
				Deadline:  time.Now().Add(leaseTime),
			}
			a.leases[reqID] = lease
		} else {
			lease.Deadline = time.Now().Add(leaseTime)
		}

		return nil
	})
}

func (a *LeaseActor) Renew(reqID string) error {
	return actor.Wait(context.TODO(), a.commandChan, func() error {
		lease, ok := a.leases[reqID]
		if !ok {
			return fmt.Errorf("lease not found for this lock request")

		} else {
			lease.Deadline = time.Now().Add(lease.LeaseTime)
		}

		return nil
	})
}

func (a *LeaseActor) Remove(reqID string) error {
	return actor.Wait(context.TODO(), a.commandChan, func() error {
		delete(a.leases, reqID)
		return nil
	})
}

func (a *LeaseActor) Serve() {
	cmdChan := a.commandChan.BeginChanReceive()
	defer a.commandChan.CloseChanReceive()

	for {
		if a.ticker != nil {
			select {
			case cmd := <-cmdChan:
				cmd()

			case now := <-a.ticker.C:
				for reqID, lease := range a.leases {
					if now.After(lease.Deadline) {

						// TODO 可以考虑打个日志
						logger.Std.Infof("lock request %s is timeout, will release it", reqID)

						a.releaseActor.DelayRelease([]string{reqID})
					}
				}

			}
		} else {
			select {
			case cmd := <-cmdChan:
				cmd()
			}
		}
	}
}
