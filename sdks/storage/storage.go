package cdssdk

import (
	"fmt"

	"gitlink.org.cn/cloudream/common/pkgs/types"
	"gitlink.org.cn/cloudream/common/utils/serder"
)

type Storage struct {
	StorageID StorageID `json:"storageID" gorm:"column:StorageID; primaryKey; type:bigint; autoIncrement;"`
	Name      string    `json:"name" gorm:"column:Name; type:varchar(256); not null"`
	// 完全管理此存储服务的Hub的ID
	MasterHub HubID `json:"masterHub" gorm:"column:MasterHub; type:bigint; not null"`
	// 存储服务的类型，包含地址信息以及鉴权所需数据
	Type StorageType `json:"type" gorm:"column:Type; type:json; not null; serializer:union"`
	// 分片存储服务的配置数据
	ShardStore ShardStoreConfig `json:"shardStore" gorm:"column:ShardStore; type:json; serializer:union"`
	// 共享存储服务的配置数据
	SharedStore SharedStoreConfig `json:"sharedStore" gorm:"column:SharedStore; type:json; serializer:union"`
	// SharedStore
	// 存储服务拥有的特别功能
	Features []StorageFeature `json:"features" gorm:"column:Features; type:json; serializer:union"`
}

func (Storage) TableName() string {
	return "Storage"
}

func (s *Storage) String() string {
	return fmt.Sprintf("%v(%v)", s.Name, s.StorageID)
}

// 存储服务地址
type StorageType interface {
	GetStorageType() string
	// 输出调试用的字符串，不要包含敏感信息
	String() string
}

var _ = serder.UseTypeUnionInternallyTagged(types.Ref(types.NewTypeUnion[StorageType](
	(*LocalStorageType)(nil),
	(*OBSType)(nil),
	(*OSSType)(nil),
	(*COSType)(nil),
)), "type")

type LocalStorageType struct {
	serder.Metadata `union:"Local"`
	Type            string `json:"type"`
}

func (a *LocalStorageType) GetStorageType() string {
	return "Local"
}

func (a *LocalStorageType) String() string {
	return "Local"
}

type OSSType struct {
	serder.Metadata `union:"OSS"`
	Type            string `json:"type"`
	Region          string `json:"region"`
	AK              string `json:"accessKeyId"`
	SK              string `json:"secretAccessKey"`
	Endpoint        string `json:"endpoint"`
	Bucket          string `json:"bucket"`
}

func (a *OSSType) GetStorageType() string {
	return "OSS"
}

func (a *OSSType) String() string {
	return "OSS"
}

type OBSType struct {
	serder.Metadata `union:"OBS"`
	Type            string `json:"type"`
	Region          string `json:"region"`
	AK              string `json:"accessKeyId"`
	SK              string `json:"secretAccessKey"`
	Endpoint        string `json:"endpoint"`
	Bucket          string `json:"bucket"`
}

func (a *OBSType) GetStorageType() string {
	return "OBS"
}

func (a *OBSType) String() string {
	return "OBS"
}

type COSType struct {
	serder.Metadata `union:"COS"`
	Type            string `json:"type"`
	Region          string `json:"region"`
	AK              string `json:"accessKeyId"`
	SK              string `json:"secretAccessKey"`
	Endpoint        string `json:"endpoint"`
	Bucket          string `json:"bucket"`
}

func (a *COSType) GetStorageType() string {
	return "COS"
}

func (a *COSType) String() string {
	return "COS"
}
