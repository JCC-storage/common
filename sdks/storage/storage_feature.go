package cdssdk

import (
	"gitlink.org.cn/cloudream/common/pkgs/types"
	"gitlink.org.cn/cloudream/common/utils/serder"
)

// 存储服务特性
type StorageFeature interface {
	GetFeatureType() string
	// 输出调试用的字符串，不要包含敏感信息
	String() string
}

var _ = serder.UseTypeUnionInternallyTagged(types.Ref(types.NewTypeUnion[StorageFeature](
	(*TempStore)(nil),
	(*BypassWriteFeature)(nil),
	(*MultipartUploadFeature)(nil),
	(*InternalServerlessCallFeature)(nil),
)), "type")

type TempStore struct {
	serder.Metadata `union:"TempStore"`
	Type            string `json:"type"`
	TempRoot        string `json:"tempRoot"` // 临时文件存放目录
}

func (f *TempStore) GetFeatureType() string {
	return "TempStore"
}

func (f *TempStore) String() string {
	return "TempStore"
}

// 存储服务支持被非MasterHub直接上传文件
type BypassWriteFeature struct {
	serder.Metadata `union:"BypassWrite"`
	Type            string `json:"type"`
}

func (f *BypassWriteFeature) GetFeatureType() string {
	return "BypassWrite"
}

func (f *BypassWriteFeature) String() string {
	return "BypassWrite"
}

// 存储服务支持分段上传
type MultipartUploadFeature struct {
	serder.Metadata `union:"MultipartUpload"`
	Type            string `json:"type"`
	TempDir         string `json:"tempDir"`     // 临时文件存放目录
	MinPartSize     int64  `json:"minPartSize"` // 最小分段大小
	MaxPartSize     int64  `json:"maxPartSize"` // 最大分段大小
}

func (f *MultipartUploadFeature) GetFeatureType() string {
	return "MultipartUpload"
}

func (f *MultipartUploadFeature) String() string {
	return "MultipartUpload"
}

// 在存储服务所在的环境中部署有内部的Serverless服务
type InternalServerlessCallFeature struct {
	serder.Metadata `union:"InternalServerlessCall"`
	Type            string `json:"type"`
	CommandDir      string `json:"commandDir"` // 存放命令文件的目录
}

func (f *InternalServerlessCallFeature) GetFeatureType() string {
	return "InternalServerlessCall"
}

func (f *InternalServerlessCallFeature) String() string {
	return "InternalServerlessCall"
}
