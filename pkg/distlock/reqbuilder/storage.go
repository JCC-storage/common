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

func (b *StorageLockReqBuilder) ReadOneObject(storageID int, fileHash string) *StorageLockReqBuilder {
	b.locks = append(b.locks, distlock.Lock{
		Path:   b.makePath(storageID),
		Name:   lockprovider.STORAGE_ELEMENT_READ_LOCK,
		Target: *lockprovider.NewStringLockTarget().Add(fileHash),
	})
	return b
}

func (b *StorageLockReqBuilder) WriteOneObject(storageID int, fileHash string) *StorageLockReqBuilder {
	b.locks = append(b.locks, distlock.Lock{
		Path:   b.makePath(storageID),
		Name:   lockprovider.STORAGE_ELEMENT_WRITE_LOCK,
		Target: *lockprovider.NewStringLockTarget().Add(fileHash),
	})
	return b
}

func (b *StorageLockReqBuilder) CreateOneObject(storageID int, fileHash string) *StorageLockReqBuilder {
	b.locks = append(b.locks, distlock.Lock{
		Path:   b.makePath(storageID),
		Name:   lockprovider.STORAGE_ELEMENT_WRITE_LOCK,
		Target: *lockprovider.NewStringLockTarget().Add(fileHash),
	})
	return b
}

func (b *StorageLockReqBuilder) ReadAnyObject(storageID int) *StorageLockReqBuilder {
	b.locks = append(b.locks, distlock.Lock{
		Path:   b.makePath(storageID),
		Name:   lockprovider.STORAGE_SET_READ_LOCK,
		Target: *lockprovider.NewStringLockTarget(),
	})
	return b
}

func (b *StorageLockReqBuilder) WriteAnyObject(storageID int) *StorageLockReqBuilder {
	b.locks = append(b.locks, distlock.Lock{
		Path:   b.makePath(storageID),
		Name:   lockprovider.STORAGE_SET_WRITE_LOCK,
		Target: *lockprovider.NewStringLockTarget(),
	})
	return b
}

func (b *StorageLockReqBuilder) CreateAnyObject(storageID int) *StorageLockReqBuilder {
	b.locks = append(b.locks, distlock.Lock{
		Path:   b.makePath(storageID),
		Name:   lockprovider.STORAGE_SET_CREATE_LOCK,
		Target: *lockprovider.NewStringLockTarget(),
	})
	return b
}

func (b *StorageLockReqBuilder) makePath(storageID int) []string {
	return []string{distlock.STORAGE_LOCK_PATH_PREFIX, strconv.Itoa(storageID)}
}
