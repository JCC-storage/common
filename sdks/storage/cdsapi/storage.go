package cdsapi

import (
	"fmt"
	"net/url"
	"strings"

	"gitlink.org.cn/cloudream/common/consts/errorcode"
	cdssdk "gitlink.org.cn/cloudream/common/sdks/storage"
	"gitlink.org.cn/cloudream/common/utils/http2"
	"gitlink.org.cn/cloudream/common/utils/serder"
)

const StorageLoadPackagePath = "/storage/loadPackage"

type StorageLoadPackageReq struct {
	UserID    cdssdk.UserID    `json:"userID" binding:"required"`
	PackageID cdssdk.PackageID `json:"packageID" binding:"required"`
	StorageID cdssdk.StorageID `json:"storageID" binding:"required"`
}
type StorageLoadPackageResp struct {
	FullPath    string `json:"fullPath"` // TODO 临时保留给中期测试的前端使用，后续会删除
	PackagePath string `json:"packagePath"`
	LocalBase   string `json:"localBase"`
	RemoteBase  string `json:"remoteBase"`
}

func (c *Client) StorageLoadPackage(req StorageLoadPackageReq) (*StorageLoadPackageResp, error) {
	url, err := url.JoinPath(c.baseURL, StorageLoadPackagePath)
	if err != nil {
		return nil, err
	}

	resp, err := http2.PostJSON(url, http2.RequestParam{
		Body: req,
	})
	if err != nil {
		return nil, err
	}

	codeResp, err := ParseJSONResponse[response[StorageLoadPackageResp]](resp)
	if err != nil {
		return nil, err
	}

	if codeResp.Code == errorcode.OK {
		return &codeResp.Data, nil
	}

	return nil, codeResp.ToError()
}

const StorageCreatePackagePath = "/storage/createPackage"

type StorageCreatePackageReq struct {
	UserID       cdssdk.UserID    `json:"userID" binding:"required"`
	StorageID    cdssdk.StorageID `json:"storageID" binding:"required"`
	Path         string           `json:"path" binding:"required"`
	BucketID     cdssdk.BucketID  `json:"bucketID" binding:"required"`
	Name         string           `json:"name" binding:"required"`
	NodeAffinity *cdssdk.NodeID   `json:"nodeAffinity"`
}

type StorageCreatePackageResp struct {
	PackageID cdssdk.PackageID `json:"packageID"`
}

func (c *Client) StorageCreatePackage(req StorageCreatePackageReq) (*StorageCreatePackageResp, error) {
	url, err := url.JoinPath(c.baseURL, StorageCreatePackagePath)
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

const StorageGetPath = "/storage/get"

type StorageGet struct {
	UserID    cdssdk.UserID    `form:"userID" json:"userID" binding:"required"`
	StorageID cdssdk.StorageID `form:"storageID" json:"storageID" binding:"required"`
}
type StorageGetResp struct {
	cdssdk.Storage
}

func (c *Client) StorageGet(req StorageGet) (*StorageGetResp, error) {
	url, err := url.JoinPath(c.baseURL, StorageGetPath)
	if err != nil {
		return nil, err
	}

	resp, err := http2.GetForm(url, http2.RequestParam{
		Query: req,
	})
	if err != nil {
		return nil, err
	}

	codeResp, err := ParseJSONResponse[response[StorageGetResp]](resp)
	if err != nil {
		return nil, err
	}

	if codeResp.Code == errorcode.OK {
		return &codeResp.Data, nil
	}

	return nil, codeResp.ToError()
}
