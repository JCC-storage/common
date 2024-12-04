package cdssdk

import (
	"fmt"

	"gitlink.org.cn/cloudream/common/pkgs/types"
	"gitlink.org.cn/cloudream/common/utils/serder"
)

type SharedStoreConfig interface {
	GetSharedStoreType() string
	// 输出调试用的字符串，不要包含敏感信息
	String() string
}

var _ = serder.UseTypeUnionInternallyTagged(types.Ref(types.NewTypeUnion[SharedStoreConfig](
	(*LocalSharedStorage)(nil),
)), "type")

type LocalSharedStorage struct {
	serder.Metadata `union:"Local"`
	Type            string `json:"type"`
	// 调度Package时的Package的根路径
	LoadBase string `json:"loadBase"`
}

func (s *LocalSharedStorage) GetSharedStoreType() string {
	return "Local"
}

func (s *LocalSharedStorage) String() string {
	return fmt.Sprintf("Local[LoadBase=%v]", s.LoadBase)
}
