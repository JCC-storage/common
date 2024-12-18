package uploadersdk

import cdssdk "gitlink.org.cn/cloudream/common/sdks/storage"

type ClusterID string

type BlockChain struct {
	DataID       DataID `gorm:"column:dataID" json:"dataID"`
	BlockChainID string `gorm:"column:blockChainID" json:"blockChainID"`
	FileName     string `gorm:"column:fileName" json:"fileName"`
	FileHash     string `gorm:"column:fileHash" json:"fileHash"`
	FileSize     int64  `gorm:"column:fileSize" json:"fileSize"`
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
