package distlock

import "gitlink.org.cn/cloudream/common/pkgs/distlock/internal"

type Mutex struct {
	svc       *Service
	lockReq   internal.LockRequest
	lockReqID string
}

func NewMutex(svc *Service, lockReq internal.LockRequest) *Mutex {
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

func (m *Mutex) Unlock() {
	m.svc.Release(m.lockReqID)
}
