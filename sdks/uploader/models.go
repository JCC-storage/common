package uploadersdk

import (
	cdssdk "gitlink.org.cn/cloudream/common/sdks/storage"
	"time"
)

type ClusterID string

type BlockChain struct {
	ObjectID     cdssdk.ObjectID `gorm:"column:object_id" json:"objectID"`
	BlockChainID string          `gorm:"column:blockChain_id" json:"blockChainID"`
	//FileName     string          `gorm:"column:file_name" json:"fileName"`
	//FileHash     string          `gorm:"column:file_hash" json:"fileHash"`
	//FileSize     int64           `gorm:"column:file_size" json:"fileSize"`
}

func (BlockChain) TableName() string {
	return "BlockChain" // 确保和数据库中的表名一致
}

type BindingData struct {
	ID          DataID        `gorm:"column:ID" json:"ID"`
	UserID      cdssdk.UserID `gorm:"column:userID" json:"userID"`
	BindingName string        `gorm:"column:bindingName" json:"bindingName"`
	BindingType string        `gorm:"column:bindingType" json:"bindingType"`
}

func (BindingData) TableName() string {
	return "BindingData" // 确保和数据库中的表名一致
}

type Folder struct {
	PackageID  cdssdk.PackageID `gorm:"column:package_id" json:"packageID"`
	Path       string           `gorm:"column:path_name" json:"path"`
	CreateTime time.Time        `gorm:"column:create_time" json:"createTime"`
}

func (Folder) TableName() string {
	return "folders"
}
