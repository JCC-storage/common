package storage

import (
	"fmt"
	"io"
	"net/url"
	"strings"

	"gitlink.org.cn/cloudream/common/consts/errorcode"
	"gitlink.org.cn/cloudream/common/models"
	myhttp "gitlink.org.cn/cloudream/common/utils/http"
	"gitlink.org.cn/cloudream/common/utils/serder"
)

type PackageUploadReq struct {
	UserID     int64                      `json:"userID"`
	BucketID   int64                      `json:"bucketID"`
	Name       string                     `json:"name"`
	Redundancy models.TypedRedundancyInfo `json:"redundancy"`
	Files      []PackageUploadFile        `json:"-"`
}
type PackageUploadFile struct {
	Path string
	File io.Reader
}
type PackageUploadResp struct {
	PackageID int64 `json:"packageID,string"`
}

func (c *Client) PackageUpload(req PackageUploadReq) (*PackageUploadResp, error) {
	url, err := url.JoinPath(c.baseURL, "/package/upload")
	if err != nil {
		return nil, err
	}

	infoJSON, err := serder.ObjectToJSON(req)
	if err != nil {
		return nil, fmt.Errorf("package info to json: %w", err)
	}

	var uploadFiles []myhttp.MultiPartRequestFile
	for _, file := range req.Files {
		uploadFiles = append(uploadFiles, myhttp.MultiPartRequestFile{
			FieldName: "files",
			FileName:  file.Path,
			File:      file.File,
		})
	}

	resp, err := myhttp.PostMultiPart(url, myhttp.MultiPartRequestParam{
		Form:  map[string]string{"info": string(infoJSON)},
		Files: uploadFiles,
	})
	if err != nil {
		return nil, err
	}

	contType := resp.Header.Get("Content-Type")
	if strings.Contains(contType, myhttp.ContentTypeJSON) {
		var codeResp response[PackageUploadResp]
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

type PackageDeleteReq struct {
	UserID    int64 `json:"userID"`
	PackageID int64 `json:"packageID"`
}

func (c *Client) PackageDelete(req PackageDeleteReq) error {
	url, err := url.JoinPath(c.baseURL, "/package/delete")
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
