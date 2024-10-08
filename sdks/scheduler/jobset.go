package schsdk

import (
	"fmt"
	"net/url"
	"strings"

	"gitlink.org.cn/cloudream/common/consts/errorcode"
	cdssdk "gitlink.org.cn/cloudream/common/sdks/storage"
	"gitlink.org.cn/cloudream/common/utils/http2"
	"gitlink.org.cn/cloudream/common/utils/serder"
)

type JobSetSumbitReq struct {
	JobSetInfo
}

type JobSetSumbitResp struct {
	JobSetID          JobSetID                `json:"jobSetID"`
	FilesUploadScheme JobSetFilesUploadScheme `json:"filesUploadScheme"`
}

func (c *Client) JobSetSumbit(req JobSetSumbitReq) (*JobSetSumbitResp, error) {
	url, err := url.JoinPath(c.baseURL, "/jobSet/submit")
	if err != nil {
		return nil, err
	}

	resp, err := http2.PostJSON(url, http2.RequestParam{
		Body: req,
	})
	if err != nil {
		return nil, err
	}

	contType := resp.Header.Get("Content-Type")
	if strings.Contains(contType, http2.ContentTypeJSON) {
		var codeResp response[JobSetSumbitResp]
		if err := serder.JSONToObjectStream(resp.Body, &codeResp); err != nil {
			return nil, fmt.Errorf("parsing response: %w", err)
		}

		if codeResp.Code == errorcode.OK {
			return &codeResp.Data, nil
		}

		return nil, codeResp.ToError()
	}

	return nil, fmt.Errorf("unknow response content type: %s", contType)
}

type JobSetLocalFileUploadedReq struct {
	JobSetID  JobSetID         `json:"jobSetID"`
	LocalPath string           `json:"localPath"`
	Error     string           `json:"error"`
	PackageID cdssdk.PackageID `json:"packageID"`
}

func (c *Client) JobSetLocalFileUploaded(req JobSetLocalFileUploadedReq) error {
	url, err := url.JoinPath(c.baseURL, "/jobSet/localFileUploaded")
	if err != nil {
		return err
	}

	resp, err := http2.PostJSON(url, http2.RequestParam{
		Body: req,
	})
	if err != nil {
		return err
	}

	contType := resp.Header.Get("Content-Type")
	if strings.Contains(contType, http2.ContentTypeJSON) {
		var codeResp response[any]
		if err := serder.JSONToObjectStream(resp.Body, &codeResp); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		if codeResp.Code == errorcode.OK {
			return nil
		}

		return codeResp.ToError()
	}

	return fmt.Errorf("unknow response content type: %s", contType)
}

type JobSetGetServiceListReq struct {
	JobSetID JobSetID `json:"jobSetID"`
}

type JobSetGetServiceListResp struct {
	ServiceList []JobSetServiceInfo `json:"serviceList"`
}

func (c *Client) JobSetGetServiceList(req JobSetGetServiceListReq) (*JobSetGetServiceListResp, error) {
	url, err := url.JoinPath(c.baseURL, "/jobSet/getServiceList")
	if err != nil {
		return nil, err
	}

	resp, err := http2.GetJSON(url, http2.RequestParam{
		Body: req,
	})
	if err != nil {
		return nil, err
	}

	contType := resp.Header.Get("Content-Type")
	if strings.Contains(contType, http2.ContentTypeJSON) {
		var codeResp response[JobSetGetServiceListResp]
		if err := serder.JSONToObjectStream(resp.Body, &codeResp); err != nil {
			return nil, fmt.Errorf("parsing response: %w", err)
		}

		if codeResp.Code == errorcode.OK {
			return &codeResp.Data, nil
		}

		return nil, codeResp.ToError()
	}

	return nil, fmt.Errorf("unknow response content type: %s", contType)
}
