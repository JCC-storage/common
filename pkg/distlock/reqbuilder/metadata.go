package reqbuilder

import (
	"gitlink.org.cn/cloudream/common/pkg/distlock"
	"gitlink.org.cn/cloudream/common/pkg/distlock/lockprovider"
)

type MetadataLockReqBuilder struct {
	*LockRequestBuilder
}

func (b *LockRequestBuilder) Metadata() *MetadataLockReqBuilder {
	return &MetadataLockReqBuilder{LockRequestBuilder: b}
}

func (b *IPFSLockReqBuilder) Metadata() *MetadataLockReqBuilder {
	return &MetadataLockReqBuilder{LockRequestBuilder: b.LockRequestBuilder}
}

func (b *StorageLockReqBuilder) Metadata() *MetadataLockReqBuilder {
	return &MetadataLockReqBuilder{LockRequestBuilder: b.LockRequestBuilder}
}

func (b *MetadataLockReqBuilder) ReadOneNode(nodeID int) *MetadataLockReqBuilder {
	b.locks = append(b.locks, distlock.Lock{
		Path:   b.makePath("Node"),
		Name:   lockprovider.METADATA_ELEMENT_READ_LOCK,
		Target: *lockprovider.NewStringLockTarget().AddComponent(nodeID),
	})
	return b
}
func (b *MetadataLockReqBuilder) WriteOneNode(nodeID int) *MetadataLockReqBuilder {
	b.locks = append(b.locks, distlock.Lock{
		Path:   b.makePath("Node"),
		Name:   lockprovider.METADATA_ELEMENT_WRITE_LOCK,
		Target: *lockprovider.NewStringLockTarget().AddComponent(nodeID),
	})
	return b
}
func (b *MetadataLockReqBuilder) CreateOneNode() *MetadataLockReqBuilder {
	b.locks = append(b.locks, distlock.Lock{
		Path:   b.makePath("Node"),
		Name:   lockprovider.METADATA_ELEMENT_CREATE_LOCK,
		Target: *lockprovider.NewStringLockTarget(),
	})
	return b
}
func (b *MetadataLockReqBuilder) ReadAnyNode() *MetadataLockReqBuilder {
	b.locks = append(b.locks, distlock.Lock{
		Path:   b.makePath("Node"),
		Name:   lockprovider.METADATA_SET_READ_LOCK,
		Target: *lockprovider.NewStringLockTarget(),
	})
	return b
}
func (b *MetadataLockReqBuilder) WriteAnyNode(nodeID int) *MetadataLockReqBuilder {
	b.locks = append(b.locks, distlock.Lock{
		Path:   b.makePath("Node"),
		Name:   lockprovider.METADATA_SET_WRITE_LOCK,
		Target: *lockprovider.NewStringLockTarget(),
	})
	return b
}
func (b *MetadataLockReqBuilder) CreateAnyNode() *MetadataLockReqBuilder {
	b.locks = append(b.locks, distlock.Lock{
		Path:   b.makePath("Node"),
		Name:   lockprovider.METADATA_SET_CREATE_LOCK,
		Target: *lockprovider.NewStringLockTarget(),
	})
	return b
}

func (b *MetadataLockReqBuilder) makePath(tableName string) []string {
	return []string{distlock.METADATA_LOCK_PATH_PREFIX, tableName}
}
