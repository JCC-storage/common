package storage

import (
	"fmt"
	"io"
	"net/url"
	"strings"

	"gitlink.org.cn/cloudream/common/consts/errorcode"
	myhttp "gitlink.org.cn/cloudream/common/utils/http"
	"gitlink.org.cn/cloudream/common/utils/serder"
)

type ObjectDownloadReq struct {
	UserID   int64 `json:"userID"`
	ObjectID int64 `json:"objectID"`
}

func (c *Client) ObjectDownload(req ObjectDownloadReq) (io.ReadCloser, error) {
	url, err := url.JoinPath(c.baseURL, "/object/download")
	if err != nil {
		return nil, err
	}

	resp, err := myhttp.GetJSON(url, myhttp.RequestParam{
		Query: req,
	})
	if err != nil {
		return nil, err
	}

	contType := resp.Header.Get("Content-Type")

	if strings.Contains(contType, myhttp.ContentTypeJSON) {
		var codeResp response[any]
		if err := serder.JSONToObjectStream(resp.Body, &codeResp); err != nil {
			return nil, fmt.Errorf("parsing response: %w", err)
		}

		return nil, codeResp.ToError()
	}

	if strings.Contains(contType, myhttp.ContentTypeOctetStream) {
		return resp.Body, nil
	}

	return nil, fmt.Errorf("unknow response content type: %s", contType)
}

type ObjectUploadReq struct {
	UserID     int64            `json:"userID"`
	BucketID   int64            `json:"bucketID"`
	FileSize   int64            `json:"fileSize"`
	ObjectName string           `json:"objectName"`
	Redundancy MyRedundancyInfo `json:"redundancy"`
	File       io.Reader        `json:"-"`
}

type MyRedundancyInfo struct {
	Type string `json:"type"`
	Info any    `json:"config"`
}

type ObjectUploadResp struct {
	ObjectID int64 `json:"objectID,string"`
}

func (c *Client) ObjectUpload(req ObjectUploadReq) (*ObjectUploadResp, error) {
	url, err := url.JoinPath(c.baseURL, "/object/upload")
	if err != nil {
		return nil, err
	}

	infoJSON, err := serder.ObjectToJSON(req)
	if err != nil {
		return nil, fmt.Errorf("object info to json: %w", err)
	}

	resp, err := myhttp.PostMultiPart(url, myhttp.MultiPartRequestParam{
		Form:     map[string]string{"info": string(infoJSON)},
		DataName: req.ObjectName,
		Data:     req.File,
	})
	if err != nil {
		return nil, err
	}

	contType := resp.Header.Get("Content-Type")
	if strings.Contains(contType, myhttp.ContentTypeJSON) {
		var codeResp response[ObjectUploadResp]
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

type ObjectDeleteReq struct {
	UserID   int64 `json:"userID"`
	ObjectID int64 `json:"objectID"`
}

func (c *Client) ObjectDelete(req ObjectDeleteReq) error {
	url, err := url.JoinPath(c.baseURL, "/object/delete")
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
