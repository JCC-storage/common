package reqbuilder

import (
	"gitlink.org.cn/cloudream/common/pkgs/distlock"
	"gitlink.org.cn/cloudream/common/pkgs/distlock/lockprovider"
)

type MetadataStorageObjectLockReqBuilder struct {
	*MetadataLockReqBuilder
}

func (b *MetadataLockReqBuilder) StorageObject() *MetadataStorageObjectLockReqBuilder {
	return &MetadataStorageObjectLockReqBuilder{MetadataLockReqBuilder: b}
}

func (b *MetadataStorageObjectLockReqBuilder) ReadOne(storageID int64, userID int64, objectID int64) *MetadataStorageObjectLockReqBuilder {
	b.locks = append(b.locks, distlock.Lock{
		Path:   b.makePath("StorageObject"),
		Name:   lockprovider.METADATA_ELEMENT_READ_LOCK,
		Target: *lockprovider.NewStringLockTarget().Add(storageID, userID, objectID),
	})
	return b
}
func (b *MetadataStorageObjectLockReqBuilder) WriteOne(storageID int64, userID int64, objectID int64) *MetadataStorageObjectLockReqBuilder {
	b.locks = append(b.locks, distlock.Lock{
		Path:   b.makePath("StorageObject"),
		Name:   lockprovider.METADATA_ELEMENT_WRITE_LOCK,
		Target: *lockprovider.NewStringLockTarget().Add(storageID, userID, objectID),
	})
	return b
}
func (b *MetadataStorageObjectLockReqBuilder) CreateOne(storageID int64, userID int64, objectID int64) *MetadataStorageObjectLockReqBuilder {
	b.locks = append(b.locks, distlock.Lock{
		Path:   b.makePath("StorageObject"),
		Name:   lockprovider.METADATA_ELEMENT_CREATE_LOCK,
		Target: *lockprovider.NewStringLockTarget().Add(storageID, userID, objectID),
	})
	return b
}
func (b *MetadataStorageObjectLockReqBuilder) ReadAny() *MetadataStorageObjectLockReqBuilder {
	b.locks = append(b.locks, distlock.Lock{
		Path:   b.makePath("StorageObject"),
		Name:   lockprovider.METADATA_SET_READ_LOCK,
		Target: *lockprovider.NewStringLockTarget(),
	})
	return b
}
func (b *MetadataStorageObjectLockReqBuilder) WriteAny() *MetadataStorageObjectLockReqBuilder {
	b.locks = append(b.locks, distlock.Lock{
		Path:   b.makePath("StorageObject"),
		Name:   lockprovider.METADATA_SET_WRITE_LOCK,
		Target: *lockprovider.NewStringLockTarget(),
	})
	return b
}
func (b *MetadataStorageObjectLockReqBuilder) CreateAny() *MetadataStorageObjectLockReqBuilder {
	b.locks = append(b.locks, distlock.Lock{
		Path:   b.makePath("StorageObject"),
		Name:   lockprovider.METADATA_SET_CREATE_LOCK,
		Target: *lockprovider.NewStringLockTarget(),
	})
	return b
}
