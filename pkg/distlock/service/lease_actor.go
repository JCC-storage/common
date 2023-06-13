package service

import (
	"fmt"
	"time"

	"gitlink.org.cn/cloudream/common/pkg/actor"
)

type lockRequestLease struct {
	RequestID string
	Deadline  time.Time
}

type leaseActor struct {
	leases map[string]*lockRequestLease
	ticker *time.Ticker

	commandChan *actor.CommandChannel

	mainActor *mainActor
}

func newLeaseActor() *leaseActor {
	return &leaseActor{
		leases:      make(map[string]*lockRequestLease),
		commandChan: actor.NewCommandChannel(),
	}
}

func (a *leaseActor) Init(mainActor *mainActor) {
	a.mainActor = mainActor
}

func (a *leaseActor) StartChecking() error {
	return actor.Wait(a.commandChan, func() error {
		a.ticker = time.NewTicker(time.Second)
		return nil
	})
}

func (a *leaseActor) StopChecking() error {
	return actor.Wait(a.commandChan, func() error {
		if a.ticker != nil {
			a.ticker.Stop()
		}
		a.ticker = nil
		return nil
	})
}

func (a *leaseActor) Add(reqID string, leaseTime time.Duration) error {
	return actor.Wait(a.commandChan, func() error {
		lease, ok := a.leases[reqID]
		if !ok {
			lease = &lockRequestLease{
				RequestID: reqID,
				Deadline:  time.Now().Add(leaseTime),
			}
			a.leases[reqID] = lease
		} else {
			lease.Deadline = time.Now().Add(leaseTime)
		}

		return nil
	})
}

func (a *leaseActor) Renew(reqID string, leaseTime time.Duration) error {
	return actor.Wait(a.commandChan, func() error {
		lease, ok := a.leases[reqID]
		if !ok {
			return fmt.Errorf("lease not found for this lock request")

		} else {
			lease.Deadline = time.Now().Add(leaseTime)
		}

		return nil
	})
}

func (a *leaseActor) Remove(reqID string) error {
	return actor.Wait(a.commandChan, func() error {
		delete(a.leases, reqID)
		return nil
	})
}

func (a *leaseActor) Server() error {
	for {
		if a.ticker != nil {
			select {
			case cmd, ok := <-a.commandChan.ChanReceive():
				if !ok {
					a.ticker.Stop()
					return fmt.Errorf("command chan closed")
				}

				cmd()

			case now := <-a.ticker.C:
				for reqID, lease := range a.leases {
					if now.After(lease.Deadline) {
						delete(a.leases, reqID)

						// TODO 可以考虑打个日志

						a.mainActor.Release(reqID)
					}
				}

			}
		} else {
			select {
			case cmd, ok := <-a.commandChan.ChanReceive():
				if !ok {
					return fmt.Errorf("command chan closed")
				}

				cmd()
			}
		}
	}
}
