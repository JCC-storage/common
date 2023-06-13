package lockprovider

import (
	"fmt"

	"github.com/samber/lo"
	"gitlink.org.cn/cloudream/common/pkg/distlock"
	mylo "gitlink.org.cn/cloudream/common/utils/lo"
)

const (
	IPFS_SET_READ_LOCK   = "SetRead"
	IPFS_SET_WRITE_LOCK  = "SetWrite"
	IPFS_SET_CREATE_LOCK = "SetCreate"

	IPFS_ELEMENT_READ_LOCK  = "ElementRead"
	IPFS_ELEMENT_WRITE_LOCK = "ElementWrite"
)

type ipfsElementLock struct {
	target     StringLockTarget
	requestIDs []string
}

type IPFSLock struct {
	setReadReqIDs   []string
	setWriteReqIDs  []string
	setCreateReqIDs []string

	elementReadLocks  []*ipfsElementLock
	elementWriteLocks []*ipfsElementLock

	lockCompatibilityTable LockCompatibilityTable
}

func NewIPFSLock() *IPFSLock {

	ipfsLock := IPFSLock{
		lockCompatibilityTable: LockCompatibilityTable{},
	}

	compTable := &ipfsLock.lockCompatibilityTable

	compTable.
		Column(IPFS_ELEMENT_READ_LOCK, func() bool { return len(ipfsLock.elementReadLocks) > 0 }).
		Column(IPFS_ELEMENT_WRITE_LOCK, func() bool { return len(ipfsLock.elementWriteLocks) > 0 }).
		Column(IPFS_SET_READ_LOCK, func() bool { return len(ipfsLock.setReadReqIDs) > 0 }).
		Column(IPFS_SET_WRITE_LOCK, func() bool { return len(ipfsLock.setWriteReqIDs) > 0 }).
		Column(IPFS_SET_CREATE_LOCK, func() bool { return len(ipfsLock.setCreateReqIDs) > 0 })

	comp := LockCompatible()
	uncp := LockUncompatible()
	trgt := LockSpecial(func(lock distlock.Lock, testLockName string) bool {
		strTar := lock.Target.(StringLockTarget)
		if testLockName == IPFS_ELEMENT_READ_LOCK {
			// 如果没有任何锁的锁对象与当前的锁对象冲突，那么这个锁可以加
			return lo.NoneBy(ipfsLock.elementReadLocks, func(other *ipfsElementLock) bool { return strTar.IsConflict(&other.target) })
		}

		return lo.NoneBy(ipfsLock.elementWriteLocks, func(other *ipfsElementLock) bool { return strTar.IsConflict(&other.target) })
	})

	compTable.MustRow(comp, trgt, comp, uncp, comp)
	compTable.MustRow(trgt, trgt, uncp, uncp, uncp)
	compTable.MustRow(comp, uncp, comp, uncp, uncp)
	compTable.MustRow(uncp, uncp, uncp, uncp, uncp)
	compTable.MustRow(comp, uncp, uncp, uncp, comp)

	return &ipfsLock
}

// CanLock 判断这个锁能否锁定成功
func (l *IPFSLock) CanLock(lock distlock.Lock) error {
	return l.lockCompatibilityTable.Test(lock)
}

// 锁定
func (l *IPFSLock) Lock(reqID string, lock distlock.Lock) error {
	switch lock.Name {
	case IPFS_SET_READ_LOCK:
		l.setReadReqIDs = append(l.setReadReqIDs, reqID)
	case IPFS_SET_WRITE_LOCK:
		l.setWriteReqIDs = append(l.setWriteReqIDs, reqID)
	case IPFS_SET_CREATE_LOCK:
		l.setCreateReqIDs = append(l.setCreateReqIDs, reqID)

	case IPFS_ELEMENT_READ_LOCK:
		l.elementReadLocks = l.addElementLock(lock, l.elementReadLocks, reqID)
	case IPFS_ELEMENT_WRITE_LOCK:
		l.elementWriteLocks = l.addElementLock(lock, l.elementWriteLocks, reqID)

	default:
		return fmt.Errorf("unknow lock name: %s", lock.Name)
	}

	return nil
}

func (l *IPFSLock) addElementLock(lock distlock.Lock, locks []*ipfsElementLock, reqID string) []*ipfsElementLock {
	strTarget := lock.Target.(StringLockTarget)
	lck, ok := lo.Find(locks, func(l *ipfsElementLock) bool { return strTarget.IsConflict(&l.target) })
	if !ok {
		lck = &ipfsElementLock{
			target: strTarget,
		}
		locks = append(locks, lck)
	}

	lck.requestIDs = append(lck.requestIDs, reqID)
	return locks
}

// 解锁
func (l *IPFSLock) Unlock(reqID string, lock distlock.Lock) error {
	switch lock.Name {
	case IPFS_SET_READ_LOCK:
		l.setReadReqIDs = mylo.Remove(l.setReadReqIDs, reqID)
	case IPFS_SET_WRITE_LOCK:
		l.setWriteReqIDs = mylo.Remove(l.setWriteReqIDs, reqID)
	case IPFS_SET_CREATE_LOCK:
		l.setCreateReqIDs = mylo.Remove(l.setCreateReqIDs, reqID)

	case IPFS_ELEMENT_READ_LOCK:
		l.elementReadLocks = l.removeElementLock(lock, l.elementReadLocks, reqID)
	case IPFS_ELEMENT_WRITE_LOCK:
		l.elementWriteLocks = l.removeElementLock(lock, l.elementWriteLocks, reqID)

	default:
		return fmt.Errorf("unknow lock name: %s", lock.Name)
	}

	return nil
}

func (l *IPFSLock) removeElementLock(lock distlock.Lock, locks []*ipfsElementLock, reqID string) []*ipfsElementLock {
	strTarget := lock.Target.(StringLockTarget)
	lck, index, ok := lo.FindIndexOf(locks, func(l *ipfsElementLock) bool { return strTarget.IsConflict(&l.target) })
	if !ok {
		return locks
	}

	lck.requestIDs = mylo.Remove(lck.requestIDs, reqID)

	if len(lck.requestIDs) == 0 {
		locks = mylo.RemoveAt(locks, index)
	}

	return locks
}

// GetTargetString 将锁对象序列化为字符串，方便存储到ETCD
func (l *IPFSLock) GetTargetString(target any) (string, error) {
	tar := target.(StringLockTarget)
	return StringLockTargetToString(&tar)
}

// ParseTargetString 解析字符串格式的锁对象数据
func (l *IPFSLock) ParseTargetString(targetStr string) (any, error) {
	return StringLockTargetFromString(targetStr)
}

// Clear 清除内部所有状态
func (l *IPFSLock) Clear() {
	l.setReadReqIDs = nil
	l.setWriteReqIDs = nil
	l.setCreateReqIDs = nil
	l.elementReadLocks = nil
	l.elementWriteLocks = nil
}
