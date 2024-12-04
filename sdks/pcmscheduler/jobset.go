package sch

import (
	"fmt"
	schsdk "gitlink.org.cn/cloudream/common/sdks/scheduler"
	cdssdk "gitlink.org.cn/cloudream/common/sdks/storage"
	"gitlink.org.cn/cloudream/common/utils/http2"
	"gitlink.org.cn/cloudream/common/utils/serder"
	"net/url"
	"strings"
)

type GetClusterInfoReq struct {
	IDs []ClusterID `json:"ids"`
}

func (c *Client) GetClusterInfo(req GetClusterInfoReq) ([]ClusterDetail, error) {
	targetUrl, err := url.JoinPath(c.baseURL, "/queryResources")
	if err != nil {
		return nil, err
	}
	resp, err := http2.GetJSON(targetUrl, http2.RequestParam{Body: req})
	if err != nil {
		return nil, err
	}
	contType := resp.Header.Get("Content-Type")
	if strings.Contains(contType, http2.ContentTypeJSON) {

		var codeResp response[[]ClusterDetail]
		if err := serder.JSONToObjectStream(resp.Body, &codeResp); err != nil {
			return nil, fmt.Errorf("parsing response: %w", err)
		}

		if codeResp.Code == ResponseCodeOK {
			return codeResp.Data, nil
		}

		return nil, codeResp.ToError()
	}

	return nil, fmt.Errorf("unknow response content type: %s", contType)
}

type CreateJobReq struct {
	DataDistribute DataDistribute          `json:"dataDistribute"`
	Resources      schsdk.JobResourcesInfo `json:"resources"`
}

type DataDistribute struct {
	Dataset []DatasetDistribute `json:"dataset"`
	Code    []CodeDistribute    `json:"code"`
	Image   []ImageDistribute   `json:"image"`
	Model   []ModelDistribute   `json:"model"`
}

type DatasetDistribute struct {
	DataName  string           `json:"dataName"`
	PackageID cdssdk.PackageID `json:"packageID"`
	Clusters  []ClusterID      `json:"clusters"`
}

type CodeDistribute struct {
	DataName  string           `json:"dataName"`
	PackageID cdssdk.PackageID `json:"packageID"`
	Clusters  []ClusterID      `json:"clusters"`
}

type ImageDistribute struct {
	DataName  string           `json:"dataName"`
	PackageID cdssdk.PackageID `json:"packageID"`
	Clusters  []ClusterID      `json:"clusters"`
}

type ModelDistribute struct {
	DataName  string           `json:"dataName"`
	PackageID cdssdk.PackageID `json:"packageID"`
	Clusters  []ClusterID      `json:"clusters"`
}

type Cluster struct {
	ClusterID ClusterID        `json:"clusterID"`
	StorageID cdssdk.StorageID `json:"storageID"`
}

type CreateJobResp struct {
	TaskID        TaskID         `json:"taskID"`
	ScheduleDatas []ScheduleData `json:"scheduleDatas"`
}

type ScheduleData struct {
	DataType    string           `json:"dataType"`
	PackageID   cdssdk.PackageID `json:"packageID"`
	StorageType string           `json:"storageType"`
	ClusterIDs  []ClusterID      `json:"clusterIDs"`
}

func (c *Client) CreateJob(req CreateJobReq) (CreateJobResp, error) {

}

type DataScheduleReq struct {
	PackageID   cdssdk.PackageID `json:"packageID"`
	StorageType string           `json:"storageType"`
	Clusters    []Cluster        `json:"clusters"`
}

type DataScheduleResp struct {
	Results []DataScheduleResult `json:"results"`
}

type DataScheduleResult struct {
	ClusterID       ClusterID        `json:"clusterID"`
	PackageID       cdssdk.PackageID `json:"packageID"`
	PackageFullPath string           `json:"packageFullPath"`
	Status          bool             `json:"status"`
	Msg             string           `json:"msg"`
}

func (c *Client) DataSchedule(req DataScheduleReq) (DataScheduleResp, error) {

}

type RunJobReq struct {
	TaskID         TaskID                `json:"taskID"`
	ScheduledDatas []DataScheduleResults `json:"scheduledDatas"`
}

type DataScheduleResults struct {
	DataType string               `json:"dataType"`
	Results  []DataScheduleResult `json:"results"`
}

func (c *Client) RunJob(req RunJobReq) error {

}

type CancelJobReq struct {
	TaskID TaskID `json:"taskID"`
	Msg    string `json:"msg"`
}

func (c *Client) CancelJob(req CancelJobReq) error {

}
