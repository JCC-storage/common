package schsdk

import (
	"fmt"
	"net/url"
	"strings"

	"gitlink.org.cn/cloudream/common/consts/errorcode"
	cdssdk "gitlink.org.cn/cloudream/common/sdks/storage"
	myhttp "gitlink.org.cn/cloudream/common/utils/http"
	"gitlink.org.cn/cloudream/common/utils/serder"
)

type JobSetSumbitReq struct {
	JobSetInfo
}

type JobSetSumbitResp struct {
	JobSetID          string                  `json:"jobSetID"`
	FilesUploadScheme JobSetFilesUploadScheme `json:"filesUploadScheme"`
}

func (c *Client) JobSetSumbit(req JobSetSumbitReq) (*JobSetSumbitResp, error) {
	url, err := url.JoinPath(c.baseURL, "/jobSet/submit")
	if err != nil {
		return nil, err
	}

	resp, err := myhttp.PostJSON(url, myhttp.RequestParam{
		Body: req,
	})
	if err != nil {
		return nil, err
	}

	contType := resp.Header.Get("Content-Type")
	if strings.Contains(contType, myhttp.ContentTypeJSON) {
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
	JobSetID  string           `json:"jobSetID"`
	LocalPath string           `json:"localPath"`
	Error     string           `json:"error"`
	PackageID cdssdk.PackageID `json:"packageID"`
}

func (c *Client) JobSetLocalFileUploaded(req JobSetLocalFileUploadedReq) error {
	url, err := url.JoinPath(c.baseURL, "/jobSet/localFileUploaded")
	if err != nil {
		return err
	}

	resp, err := myhttp.PostJSON(url, myhttp.RequestParam{
		Body: req,
	})
	if err != nil {
		return err
	}

	contType := resp.Header.Get("Content-Type")
	if strings.Contains(contType, myhttp.ContentTypeJSON) {
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
