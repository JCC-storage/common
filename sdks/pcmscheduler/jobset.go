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
	IDs []schsdk.ClusterID `json:"clusterIDs"`
}

type GetClusterInfoResp struct {
	Data    []ClusterDetail `json:"data"`
	TraceId string          `json:"traceId"`
}

func (c *Client) GetClusterInfo(req GetClusterInfoReq) ([]ClusterDetail, error) {
	targetUrl, err := url.JoinPath(c.baseURL, "/queryResources")
	if err != nil {
		return nil, err
	}

	resp, err := http2.PostJSON(targetUrl, http2.RequestParam{Body: req})
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
	JobResources   schsdk.JobResources `json:"jobResources"`
	DataDistribute DataDistribute      `json:"dataDistribute"`
}

type DataDistribute struct {
	Dataset []DatasetDistribute `json:"dataset"`
	Code    []CodeDistribute    `json:"code"`
	Image   []ImageDistribute   `json:"image"`
	Model   []ModelDistribute   `json:"model"`
}

type DataDetail struct {
	ClusterID schsdk.ClusterID `json:"clusterID"`
	JsonData  string           `json:"jsonData"`
}

type DatasetDistribute struct {
	DataName  string           `json:"dataName"`
	PackageID cdssdk.PackageID `json:"packageID"`
	Clusters  []DataDetail     `json:"clusters"`
}

type CodeDistribute struct {
	DataName  string           `json:"dataName"`
	PackageID cdssdk.PackageID `json:"packageID"`
	Clusters  []DataDetail     `json:"clusters"`
}

type ImageDistribute struct {
	DataName  string           `json:"dataName"`
	PackageID cdssdk.PackageID `json:"packageID"`
	Clusters  []DataDetail     `json:"clusters"`
}

type ModelDistribute struct {
	DataName  string           `json:"dataName"`
	PackageID cdssdk.PackageID `json:"packageID"`
	Clusters  []DataDetail     `json:"clusters"`
}

type CreateJobResp struct {
	TaskID        TaskID         `json:"taskID"`
	ScheduleDatas []ScheduleData `json:"scheduleDatas"`
}

type ScheduleData struct {
	DataType    string             `json:"dataType"`
	PackageID   cdssdk.PackageID   `json:"packageID"`
	StorageType string             `json:"storageType"`
	ClusterIDs  []schsdk.ClusterID `json:"clusterIDs"`
}

func (c *Client) CreateJob(req CreateJobReq) (*CreateJobResp, error) {
	targetUrl, err := url.JoinPath(c.baseURL, "/jobSet/submit")
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
		var codeResp response[CreateJobResp]
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

type RunJobReq struct {
	TaskID         TaskID                `json:"taskID"`
	ScheduledDatas []DataScheduleResults `json:"scheduledDatas"`
}

type DataScheduleResult struct {
	Clusters        DataDetail       `json:"clusters"`
	PackageID       cdssdk.PackageID `json:"packageID"`
	PackageFullPath string           `json:"packageFullPath"`
	Status          bool             `json:"status"`
	Msg             string           `json:"msg"`
}

type DataScheduleResults struct {
	DataType string               `json:"dataType"`
	Results  []DataScheduleResult `json:"results"`
}

func (c *Client) RunJob(req RunJobReq) error {
	targetUrl, err := url.JoinPath(c.baseURL, "/jobSet/submit")
	if err != nil {
		return err
	}

	resp, err := http2.PostJSON(targetUrl, http2.RequestParam{
		Body: req,
	})
	if err != nil {
		return err
	}

	contType := resp.Header.Get("Content-Type")
	if strings.Contains(contType, http2.ContentTypeJSON) {
		var codeResp response[string]
		if err := serder.JSONToObjectStream(resp.Body, &codeResp); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		if codeResp.Code == ResponseCodeOK {
			return nil
		}

		return codeResp.ToError()
	}

	return fmt.Errorf("unknow response content type: %s", contType)

}

type CancelJobReq struct {
	TaskID TaskID `json:"taskID"`
	Msg    string `json:"msg"`
}

func (c *Client) CancelJob(req CancelJobReq) error {
	targetUrl, err := url.JoinPath(c.baseURL, "/queryResources")
	if err != nil {
		return err
	}
	resp, err := http2.GetJSON(targetUrl, http2.RequestParam{Body: req})
	if err != nil {
		return err
	}
	contType := resp.Header.Get("Content-Type")
	if strings.Contains(contType, http2.ContentTypeJSON) {

		var codeResp response[string]
		if err := serder.JSONToObjectStream(resp.Body, &codeResp); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		if codeResp.Code == ResponseCodeOK {
			return nil
		}

		return codeResp.ToError()
	}

	return fmt.Errorf("unknow response content type: %s", contType)
}
