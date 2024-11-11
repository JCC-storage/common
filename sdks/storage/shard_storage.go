package cdssdk

import (
	"fmt"

	"gitlink.org.cn/cloudream/common/pkgs/types"
	"gitlink.org.cn/cloudream/common/utils/serder"
)

// 分片存储服务的配置数据
type ShardStoreConfig interface {
	GetType() string
	// 输出调试用的字符串，不要包含敏感信息
	String() string
}

var _ = serder.UseTypeUnionInternallyTagged(types.Ref(types.NewTypeUnion[ShardStoreConfig](
	(*LocalShardStorage)(nil),
)), "type")

type LocalShardStorage struct {
	serder.Metadata `union:"Local"`
	Type            string `json:"type"`
	Root            string `json:"root"`
	MaxSize         int64  `json:"maxSize"`
}

func (s *LocalShardStorage) GetType() string {
	return "Local"
}

func (s *LocalShardStorage) String() string {
	return fmt.Sprintf("Local[root=%s, maxSize=%d]", s.Root, s.MaxSize)
}
