package reqbuilder

import (
	"gitlink.org.cn/cloudream/common/pkg/distlock"
	"gitlink.org.cn/cloudream/common/pkg/distlock/lockprovider"
)

type MetadataUserBucketLockReqBuilder struct {
	*MetadataLockReqBuilder
}

func (b *MetadataLockReqBuilder) UserBucket() *MetadataUserBucketLockReqBuilder {
	return &MetadataUserBucketLockReqBuilder{MetadataLockReqBuilder: b}
}

func (b *MetadataUserBucketLockReqBuilder) ReadOne(userID int, bucketID int) *MetadataUserBucketLockReqBuilder {
	b.locks = append(b.locks, distlock.Lock{
		Path:   b.makePath("UserBucket"),
		Name:   lockprovider.METADATA_ELEMENT_READ_LOCK,
		Target: *lockprovider.NewStringLockTarget().Add(userID, bucketID),
	})
	return b
}
func (b *MetadataUserBucketLockReqBuilder) WriteOne(userID int, bucketID int) *MetadataUserBucketLockReqBuilder {
	b.locks = append(b.locks, distlock.Lock{
		Path:   b.makePath("UserBucket"),
		Name:   lockprovider.METADATA_ELEMENT_WRITE_LOCK,
		Target: *lockprovider.NewStringLockTarget().Add(userID, bucketID),
	})
	return b
}
func (b *MetadataUserBucketLockReqBuilder) CreateOne(userID int, bucketID int) *MetadataUserBucketLockReqBuilder {
	b.locks = append(b.locks, distlock.Lock{
		Path:   b.makePath("UserBucket"),
		Name:   lockprovider.METADATA_ELEMENT_CREATE_LOCK,
		Target: *lockprovider.NewStringLockTarget().Add(userID, bucketID),
	})
	return b
}
func (b *MetadataUserBucketLockReqBuilder) ReadAny() *MetadataUserBucketLockReqBuilder {
	b.locks = append(b.locks, distlock.Lock{
		Path:   b.makePath("UserBucket"),
		Name:   lockprovider.METADATA_SET_READ_LOCK,
		Target: *lockprovider.NewStringLockTarget(),
	})
	return b
}
func (b *MetadataUserBucketLockReqBuilder) WriteAny() *MetadataUserBucketLockReqBuilder {
	b.locks = append(b.locks, distlock.Lock{
		Path:   b.makePath("UserBucket"),
		Name:   lockprovider.METADATA_SET_WRITE_LOCK,
		Target: *lockprovider.NewStringLockTarget(),
	})
	return b
}
func (b *MetadataUserBucketLockReqBuilder) CreateAny() *MetadataUserBucketLockReqBuilder {
	b.locks = append(b.locks, distlock.Lock{
		Path:   b.makePath("UserBucket"),
		Name:   lockprovider.METADATA_SET_CREATE_LOCK,
		Target: *lockprovider.NewStringLockTarget(),
	})
	return b
}