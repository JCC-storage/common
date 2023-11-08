package cdssdk

import (
	"fmt"
	"net/url"
	"strings"

	"gitlink.org.cn/cloudream/common/consts/errorcode"
	myhttp "gitlink.org.cn/cloudream/common/utils/http"
	"gitlink.org.cn/cloudream/common/utils/serder"
)

type StorageLoadPackageReq struct {
	UserID    int64 `json:"userID"`
	PackageID int64 `json:"packageID"`
	StorageID int64 `json:"storageID"`
}
type StorageLoadPackageResp struct {
	FullPath string `json:"fullPath"`
}

func (c *Client) StorageLoadPackage(req StorageLoadPackageReq) (*StorageLoadPackageResp, error) {
	url, err := url.JoinPath(c.baseURL, "/storage/loadPackage")
	if err != nil {
		return nil, err
	}

	resp, err := myhttp.PostJSON(url, myhttp.RequestParam{
		Body: req,
	})
	if err != nil {
		return nil, err
	}

	codeResp, err := myhttp.ParseJSONResponse[response[StorageLoadPackageResp]](resp)
	if err != nil {
		return nil, err
	}

	if codeResp.Code == errorcode.OK {
		return &codeResp.Data, nil
	}

	return nil, codeResp.ToError()
}

type StorageCreatePackageReq struct {
	UserID     int64               `json:"userID"`
	StorageID  int64               `json:"storageID"`
	Path       string              `json:"path"`
	BucketID   int64               `json:"bucketID"`
	Name       string              `json:"name"`
	Redundancy TypedRedundancyInfo `json:"redundancy"`
}

type StorageCreatePackageResp struct {
	PackageID int64 `json:"packageID"`
}

func (c *Client) StorageCreatePackage(req StorageCreatePackageReq) (*StorageCreatePackageResp, error) {
	url, err := url.JoinPath(c.baseURL, "/storage/createPackage")
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
		var codeResp response[StorageCreatePackageResp]
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

type StorageGetInfoReq struct {
	UserID    int64 `json:"userID"`
	StorageID int64 `json:"storageID"`
}
type StorageGetInfoResp struct {
	Name      string `json:"name"`
	NodeID    int64  `json:"nodeID"`
	Directory string `json:"directory"`
}

func (c *Client) StorageGetInfo(req StorageGetInfoReq) (*StorageGetInfoResp, error) {
	url, err := url.JoinPath(c.baseURL, "/storage/getInfo")
	if err != nil {
		return nil, err
	}

	resp, err := myhttp.GetForm(url, myhttp.RequestParam{
		Query: req,
	})
	if err != nil {
		return nil, err
	}

	codeResp, err := myhttp.ParseJSONResponse[response[StorageGetInfoResp]](resp)
	if err != nil {
		return nil, err
	}

	if codeResp.Code == errorcode.OK {
		return &codeResp.Data, nil
	}

	return nil, codeResp.ToError()
}
