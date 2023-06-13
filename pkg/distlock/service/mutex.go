package service

import "gitlink.org.cn/cloudream/common/pkg/distlock"

type Mutex struct {
	svc       *Service
	lockReq   distlock.LockRequest
	lockReqID string
}

func NewMutex(svc *Service, lockReq distlock.LockRequest) *Mutex {
	return &Mutex{
		svc:     svc,
		lockReq: lockReq,
	}
}

func (m *Mutex) Lock() error {
	reqID, err := m.svc.Acquire(m.lockReq)
	if err != nil {
		return err
	}

	m.lockReqID = reqID
	return nil
}

func (m *Mutex) Unlock() error {
	return m.svc.Release(m.lockReqID)
}
