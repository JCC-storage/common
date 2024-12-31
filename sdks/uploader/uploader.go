package uploadersdk

import (
	"fmt"
	"gitlink.org.cn/cloudream/common/pkgs/types"
	sch "gitlink.org.cn/cloudream/common/sdks/pcmscheduler"
	schsdk "gitlink.org.cn/cloudream/common/sdks/scheduler"
	cdssdk "gitlink.org.cn/cloudream/common/sdks/storage"
	"gitlink.org.cn/cloudream/common/utils/http2"
	"gitlink.org.cn/cloudream/common/utils/serder"
	"net/url"
	"strings"
)

type DataID int64
type FolderID int64

type Cluster struct {
	PackageID cdssdk.PackageID `gorm:"column:package_id" json:"dataID"`
	ClusterID schsdk.ClusterID `gorm:"column:cluster_id" json:"clusterID"`
	StorageID cdssdk.StorageID `gorm:"column:storage_id" json:"storageID"`
}

func (Cluster) TableName() string {
	return "uploadedCluster" // 确保和数据库中的表名一致
}

type Package struct {
	UserID          cdssdk.UserID    `gorm:"column:user_id" json:"userID"`
	PackageID       cdssdk.PackageID `gorm:"column:package_id" json:"packageID"`
	PackageName     string           `gorm:"column:package_name" json:"packageName"`
	DataType        string           `gorm:"column:data_type" json:"dataType"`
	JsonData        string           `gorm:"column:json_data" json:"jsonData"` // JSON 数据字段
	BindingID       DataID           `gorm:"column:binding_id" json:"bindingID"`
	Objects         []cdssdk.Object  `gorm:"column:objects" json:"objects"`
	UploadedCluster []Cluster        `gorm:"column:uploadedCluster" json:"uploadedCluster"`
	//UploadedCluster []Cluster `gorm:"foreignKey:package_id;references:package_id" json:"clusters"` // 关联 Cluster 数据
	//BlockChain      []BlockChain     `gorm:"foreignKey:package_id;references:package_id" json:"blockChains"` // 关联 BlockChain 数据
}

type PackageDAO struct {
	UserID          cdssdk.UserID    `gorm:"column:user_id" json:"userID"`
	PackageID       cdssdk.PackageID `gorm:"column:package_id" json:"packageID"`
	PackageName     string           `gorm:"column:package_name" json:"packageName"`
	DataType        string           `gorm:"column:data_type" json:"dataType"`
	JsonData        string           `gorm:"column:json_data" json:"jsonData"` // JSON 数据字段
	BindingID       DataID           `gorm:"column:binding_id" json:"bindingID"`
	UploadedCluster []Cluster        `gorm:"foreignKey:package_id;references:package_id" json:"clusters"` // 关联 Cluster 数据
}

type DataScheduleReq struct {
	PackageID cdssdk.PackageID `json:"packageID"`
	DataType  string           `json:"dataType"`
	Clusters  []Cluster        `json:"clusters"`
}

type codeRepository struct {
	RepositoryName string
	ClusterID      ClusterID
}

type DataScheduleResp struct {
	Results []sch.DataScheduleResult `json:"results"`
}

func (c *Client) DataSchedule(req DataScheduleReq) (*DataScheduleResp, error) {
	targetUrl, err := url.JoinPath(c.baseURL, "/jobSet/schedule")
	if err != nil {
		return nil, err
	}

	resp, err := http2.PostJSON(targetUrl, http2.RequestParam{
		Body: req,
	})
	if err != nil {
		return nil, err
	}

	contType := resp.Header.Get("Content-Type")
	if strings.Contains(contType, http2.ContentTypeJSON) {
		var codeResp response[DataScheduleResp]
		if err := serder.JSONToObjectStream(resp.Body, &codeResp); err != nil {
			return nil, fmt.Errorf("parsing response: %w", err)
		}

		if codeResp.Code == ResponseCodeOK {
			return &codeResp.Data, nil
		}

		return nil, codeResp.ToError()
	}

	return nil, fmt.Errorf("unknow response content type: %s", contType)
}

type UploadReq struct {
	Type       string             `json:"type"`
	Source     UploadSource       `json:"source"`
	Target     UploadTarget       `json:"target"`
	StorageIDs []cdssdk.StorageID `json:"storageIDs"`
}

type UploadSource interface {
	Noop()
}

var UploadSourceTypeUnion = types.NewTypeUnion[UploadSource](
	(*PackageSource)(nil),
	(*UrlSource)(nil),
)

var _ = serder.UseTypeUnionInternallyTagged(&UploadSourceTypeUnion, "type")

type PackageSource struct {
	serder.Metadata `union:"packageSource"`
	UploadSourceBase
	Type      string           `json:"type"`
	PackageID cdssdk.PackageID `json:"packageID"`
}

type UrlSource struct {
	serder.Metadata `union:"urlSource"`
	UploadSourceBase
	Type string `json:"type"`
	Url  string `json:"url"`
}

type UploadSourceBase struct{}

func (d *UploadSourceBase) Noop() {}

type UploadTarget interface {
	Noop()
}

var UploadTargetTypeUnion = types.NewTypeUnion[UploadTarget](
	(*UrlTarget)(nil),
	(*ApiTarget)(nil),
)

var _ = serder.UseTypeUnionInternallyTagged(&UploadTargetTypeUnion, "type")

type UrlTarget struct {
	serder.Metadata `union:"url"`
	UploadTargetBase
	Clusters []ClusterID `json:"clusters"`
}

type ApiTarget struct {
	serder.Metadata `union:"api"`
	UploadTargetBase
	Clusters []ClusterID `json:"clusters"`
}

type UploadTargetBase struct{}

func (d *UploadTargetBase) Noop() {}

type UploadResp struct {
	PackageID cdssdk.PackageID  `json:"packageID"`
	ObjectIDs []cdssdk.ObjectID `json:"objectIDs"`
	JsonData  string            `json:"jsonData"`
}

func (c *Client) Upload(req UploadReq) (*UploadResp, error) {
	targetUrl, err := url.JoinPath(c.baseURL, "/data/upload")
	if err != nil {
		return nil, err
	}

	resp, err := http2.PostJSON(targetUrl, http2.RequestParam{
		Body: req,
	})
	if err != nil {
		return nil, err
	}

	contType := resp.Header.Get("Content-Type")
	if strings.Contains(contType, http2.ContentTypeJSON) {
		var codeResp response[UploadResp]
		if err := serder.JSONToObjectStream(resp.Body, &codeResp); err != nil {
			return nil, fmt.Errorf("parsing response: %w", err)
		}

		if codeResp.Code == ResponseCodeOK {
			return &codeResp.Data, nil
		}

		return nil, codeResp.ToError()
	}

	return nil, fmt.Errorf("unknow response content type: %s", contType)
}
