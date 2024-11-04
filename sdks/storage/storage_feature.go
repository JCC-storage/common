package cdssdk

import (
	"gitlink.org.cn/cloudream/common/pkgs/types"
	"gitlink.org.cn/cloudream/common/utils/serder"
)

// 存储服务特性
type StorageFeature interface {
	GetType() string
	// 输出调试用的字符串，不要包含敏感信息
	String() string
}

var _ = serder.UseTypeUnionInternallyTagged(types.Ref(types.NewTypeUnion[StorageFeature](
	(*BypassUploadFeature)(nil),
	(*MultipartUploadFeature)(nil),
)), "type")

// 存储服务支持被非MasterHub直接上传文件
type BypassUploadFeature struct {
	serder.Metadata `union:"BypassUpload"`
	Type            string `json:"type"`
	// 存放上传文件的临时目录
	TempRoot string `json:"tempRoot"`
}

func (f *BypassUploadFeature) GetType() string {
	return "BypassUpload"
}

func (f *BypassUploadFeature) String() string {
	return "BypassUpload"
}

// 存储服务支持分段上传
type MultipartUploadFeature struct {
	serder.Metadata `union:"MultipartUpload"`
	Type            string `json:"type"`
}

func (f *MultipartUploadFeature) GetType() string {
	return "MultipartUpload"
}

func (f *MultipartUploadFeature) String() string {
	return "MultipartUpload"
}
