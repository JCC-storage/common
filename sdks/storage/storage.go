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
	GetType() string
	// 输出调试用的字符串，不要包含敏感信息
	String() string
}

var _ = serder.UseTypeUnionInternallyTagged(types.Ref(types.NewTypeUnion[StorageType](
	(*LocalStorageType)(nil),
)), "type")

type LocalStorageType struct {
	serder.Metadata `union:"Local"`
	Type            string `json:"type"`
}

func (a *LocalStorageType) GetType() string {
	return "Local"
}

func (a *LocalStorageType) String() string {
	return "Local"
}

type OSSType struct {
	serder.Metadata `union:"OSS"`
	Region          string `json:"region"`
	AK              string `json:"accessKeyId"`
	SK              string `json:"secretAccessKey"`
	Endpoint        string `json:"endpoint"`
	Bucket          string `json:"bucket"`
}

func (a *OSSType) GetType() string {
	return "OSSAddress"
}

func (a *OSSType) String() string {
	return "OSSAddress"
}

type OBSAddress struct {
	serder.Metadata `union:"Local"`
	Region          string `json:"region"`
	AK              string `json:"accessKeyId"`
	SK              string `json:"secretAccessKey"`
	Endpoint        string `json:"endpoint"`
	Bucket          string `json:"bucket"`
}

func (a *OBSAddress) GetType() string {
	return "OBSAddress"
}

func (a *OBSAddress) String() string {
	return "OBSAddress"
}

type COSAddress struct {
	serder.Metadata `union:"Local"`
	Region          string `json:"region"`
	AK              string `json:"accessKeyId"`
	SK              string `json:"secretAccessKey"`
	Endpoint        string `json:"endpoint"`
	Bucket          string `json:"bucket"`
}

func (a *COSAddress) GetType() string {
	return "COSAddress"
}

func (a *COSAddress) String() string {
	return "COSAddress"
}
