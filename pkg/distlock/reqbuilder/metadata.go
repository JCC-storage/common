package reqbuilder

import "gitlink.org.cn/cloudream/common/pkg/distlock"

type MetadataLockReqBuilder struct {
	*LockRequestBuilder
}

func (b *LockRequestBuilder) Metadata() *MetadataLockReqBuilder {
	return &MetadataLockReqBuilder{LockRequestBuilder: b}
}

func (b *MetadataLockReqBuilder) makePath(tableName string) []string {
	return []string{distlock.METADATA_LOCK_PATH_PREFIX, tableName}
}
