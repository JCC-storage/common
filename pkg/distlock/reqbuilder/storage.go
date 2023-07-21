package reqbuilder

import (
	"strconv"

	"gitlink.org.cn/cloudream/common/pkg/distlock"
	"gitlink.org.cn/cloudream/common/pkg/distlock/lockprovider"
)

type StorageLockReqBuilder struct {
	*LockRequestBuilder
}

func (b *LockRequestBuilder) Storage() *StorageLockReqBuilder {
	return &StorageLockReqBuilder{LockRequestBuilder: b}
}

func (b *StorageLockReqBuilder) ReadOneObject(storageID int64, userID int64, objectID int64) *StorageLockReqBuilder {
	b.locks = append(b.locks, distlock.Lock{
		Path:   b.makePath(storageID),
		Name:   lockprovider.STORAGE_ELEMENT_READ_LOCK,
		Target: *lockprovider.NewStringLockTarget().Add(userID, objectID),
	})
	return b
}

func (b *StorageLockReqBuilder) WriteOneObject(storageID int64, userID int64, objectID int64) *StorageLockReqBuilder {
	b.locks = append(b.locks, distlock.Lock{
		Path:   b.makePath(storageID),
		Name:   lockprovider.STORAGE_ELEMENT_WRITE_LOCK,
		Target: *lockprovider.NewStringLockTarget().Add(userID, objectID),
	})
	return b
}

func (b *StorageLockReqBuilder) CreateOneObject(storageID int64, userID int64, objectID int64) *StorageLockReqBuilder {
	b.locks = append(b.locks, distlock.Lock{
		Path:   b.makePath(storageID),
		Name:   lockprovider.STORAGE_ELEMENT_WRITE_LOCK,
		Target: *lockprovider.NewStringLockTarget().Add(userID, objectID),
	})
	return b
}

func (b *StorageLockReqBuilder) ReadAnyObject(storageID int64) *StorageLockReqBuilder {
	b.locks = append(b.locks, distlock.Lock{
		Path:   b.makePath(storageID),
		Name:   lockprovider.STORAGE_SET_READ_LOCK,
		Target: *lockprovider.NewStringLockTarget(),
	})
	return b
}

func (b *StorageLockReqBuilder) WriteAnyObject(storageID int64) *StorageLockReqBuilder {
	b.locks = append(b.locks, distlock.Lock{
		Path:   b.makePath(storageID),
		Name:   lockprovider.STORAGE_SET_WRITE_LOCK,
		Target: *lockprovider.NewStringLockTarget(),
	})
	return b
}

func (b *StorageLockReqBuilder) CreateAnyObject(storageID int64) *StorageLockReqBuilder {
	b.locks = append(b.locks, distlock.Lock{
		Path:   b.makePath(storageID),
		Name:   lockprovider.STORAGE_SET_CREATE_LOCK,
		Target: *lockprovider.NewStringLockTarget(),
	})
	return b
}

func (b *StorageLockReqBuilder) makePath(storageID int64) []string {
	return []string{distlock.STORAGE_LOCK_PATH_PREFIX, strconv.FormatInt(storageID, 10)}
}
