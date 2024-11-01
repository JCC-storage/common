package cdssdk

import (
	"fmt"

	"gitlink.org.cn/cloudream/common/pkgs/types"
	"gitlink.org.cn/cloudream/common/utils/serder"
)

// 存储服务地址
type StorageAddress interface {
	GetType() string
	// 输出调试用的字符串，不要包含敏感信息
	String() string
}

var _ = serder.UseTypeUnionInternallyTagged(types.Ref(types.NewTypeUnion[StorageAddress](
	(*LocalStorageAddress)(nil),
)), "type")

type LocalStorageAddress struct {
	serder.Metadata `union:"Local"`
}

func (a *LocalStorageAddress) GetType() string {
	return "Local"
}

func (a *LocalStorageAddress) String() string {
	return "Local"
}

type Storage struct {
	StorageID StorageID `json:"storageID" gorm:"column:StorageID; primaryKey; type:bigint; autoIncrement;"`
	Name      string    `json:"name" gorm:"column:Name; type:varchar(256); not null"`
	// 完全管理此存储服务的Hub的ID
	MasterHub NodeID `json:"masterHub" gorm:"column:MasterHub; type:bigint; not null"`
	// 存储服务的地址，包含鉴权所需数据
	Address StorageAddress `json:"address" gorm:"column:Address; type:json; not null; serializer:union"`
	// 存储服务拥有的特别功能
	Features []StorageFeature `json:"features" gorm:"column:Features; type:json; serializer:union"`
}

func (Storage) TableName() string {
	return "Storage"
}

func (s *Storage) String() string {
	return fmt.Sprintf("%v(%v)", s.Name, s.StorageID)
}

// 共享存储服务的配置数据
type SharedStorage struct {
	StorageID StorageID `json:"storageID" gorm:"column:StorageID; primaryKey; type:bigint"`
	// 调度文件时保存文件的根路径
	LoadBase string `json:"loadBase" gorm:"column:LoadBase; type:varchar(1024); not null"`
	// 回源数据时数据存放位置的根路径
	DataReturnBase string `json:"dataReturnBase" gorm:"column:DataReturnBase; type:varchar(1024); not null"`
}

func (SharedStorage) TableName() string {
	return "SharedStorage"
}
