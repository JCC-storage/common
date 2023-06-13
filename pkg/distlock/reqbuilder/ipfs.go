package reqbuilder

import (
	"strconv"

	"gitlink.org.cn/cloudream/common/pkg/distlock"
	"gitlink.org.cn/cloudream/common/pkg/distlock/lockprovider"
)

type IPFSLockReqBuilder struct {
	*LockRequestBuilder
}

func (b *LockRequestBuilder) IPFS() *IPFSLockReqBuilder {
	return &IPFSLockReqBuilder{LockRequestBuilder: b}
}

func (b *MetadataLockReqBuilder) IPFS() *IPFSLockReqBuilder {
	return &IPFSLockReqBuilder{LockRequestBuilder: b.LockRequestBuilder}
}

func (b *StorageLockReqBuilder) IPFS() *IPFSLockReqBuilder {
	return &IPFSLockReqBuilder{LockRequestBuilder: b.LockRequestBuilder}
}

func (b *IPFSLockReqBuilder) ReadOneRep(nodeID int, fileHash string) *IPFSLockReqBuilder {
	b.locks = append(b.locks, distlock.Lock{
		Path:   b.makePath(nodeID),
		Name:   lockprovider.IPFS_ELEMENT_READ_LOCK,
		Target: *lockprovider.NewStringLockTarget().AddComponent(fileHash),
	})
	return b
}

func (b *IPFSLockReqBuilder) WriteOneRep(nodeID int, fileHash string) *IPFSLockReqBuilder {
	b.locks = append(b.locks, distlock.Lock{
		Path:   b.makePath(nodeID),
		Name:   lockprovider.IPFS_ELEMENT_WRITE_LOCK,
		Target: *lockprovider.NewStringLockTarget().AddComponent(fileHash),
	})
	return b
}

func (b *IPFSLockReqBuilder) ReadAnyRep(nodeID int) *IPFSLockReqBuilder {
	b.locks = append(b.locks, distlock.Lock{
		Path:   b.makePath(nodeID),
		Name:   lockprovider.IPFS_SET_READ_LOCK,
		Target: *lockprovider.NewStringLockTarget(),
	})
	return b
}

func (b *IPFSLockReqBuilder) WriteAnyRep(nodeID int) *IPFSLockReqBuilder {
	b.locks = append(b.locks, distlock.Lock{
		Path:   b.makePath(nodeID),
		Name:   lockprovider.IPFS_SET_WRITE_LOCK,
		Target: *lockprovider.NewStringLockTarget(),
	})
	return b
}

func (b *IPFSLockReqBuilder) CreateAnyRep(nodeID int) *IPFSLockReqBuilder {
	b.locks = append(b.locks, distlock.Lock{
		Path:   b.makePath(nodeID),
		Name:   lockprovider.IPFS_SET_CREATE_LOCK,
		Target: *lockprovider.NewStringLockTarget(),
	})
	return b
}

func (b *IPFSLockReqBuilder) makePath(nodeID int) []string {
	return []string{distlock.IPFS_LOCK_PATH_PREFIX, strconv.Itoa(nodeID)}
}
