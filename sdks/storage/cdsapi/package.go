package cdsapi

import (
	"fmt"
	"net/url"
	"strings"

	"gitlink.org.cn/cloudream/common/consts/errorcode"
	"gitlink.org.cn/cloudream/common/pkgs/iterator"
	cdssdk "gitlink.org.cn/cloudream/common/sdks/storage"
	"gitlink.org.cn/cloudream/common/utils/http2"
	"gitlink.org.cn/cloudream/common/utils/serder"
)

type PackageService struct {
	*Client
}

func (c *Client) Package() *PackageService {
	return &PackageService{c}
}

const PackageGetPath = "/package/get"

type PackageGetReq struct {
	UserID    cdssdk.UserID    `form:"userID" json:"userID" binding:"required"`
	PackageID cdssdk.PackageID `form:"packageID" json:"packageID" binding:"required"`
}
type PackageGetResp struct {
	cdssdk.Package
}

func (c *PackageService) Get(req PackageGetReq) (*PackageGetResp, error) {
	url, err := url.JoinPath(c.baseURL, PackageGetPath)
	if err != nil {
		return nil, err
	}

	resp, err := http2.GetForm(url, http2.RequestParam{
		Query: req,
	})
	if err != nil {
		return nil, err
	}

	codeResp, err := ParseJSONResponse[response[PackageGetResp]](resp)
	if err != nil {
		return nil, err
	}

	if codeResp.Code == errorcode.OK {
		return &codeResp.Data, nil
	}

	return nil, codeResp.ToError()
}

const PackageGetByNamePath = "/package/getByName"

type PackageGetByName struct {
	UserID      cdssdk.UserID `form:"userID" json:"userID" binding:"required"`
	BucketName  string        `form:"bucketName" json:"bucketName" binding:"required"`
	PackageName string        `form:"packageName" json:"packageName" binding:"required"`
}
type PackageGetByNameResp struct {
	Package cdssdk.Package `json:"package"`
}

func (c *PackageService) GetByName(req PackageGetByName) (*PackageGetByNameResp, error) {
	url, err := url.JoinPath(c.baseURL, PackageGetByNamePath)
	if err != nil {
		return nil, err
	}

	resp, err := http2.GetForm(url, http2.RequestParam{
		Query: req,
	})
	if err != nil {
		return nil, err
	}

	codeResp, err := ParseJSONResponse[response[PackageGetByNameResp]](resp)
	if err != nil {
		return nil, err
	}

	if codeResp.Code == errorcode.OK {
		return &codeResp.Data, nil
	}

	return nil, codeResp.ToError()
}

const PackageCreatePath = "/package/create"

type PackageCreate struct {
	UserID   cdssdk.UserID   `json:"userID"`
	BucketID cdssdk.BucketID `json:"bucketID"`
	Name     string          `json:"name"`
}

type PackageCreateResp struct {
	Package cdssdk.Package `json:"package"`
}

func (s *PackageService) Create(req PackageCreate) (*PackageCreateResp, error) {
	url, err := url.JoinPath(s.baseURL, PackageCreatePath)
	if err != nil {
		return nil, err
	}

	resp, err := http2.PostJSON(url, http2.RequestParam{
		Body: req,
	})
	if err != nil {
		return nil, err
	}

	codeResp, err := ParseJSONResponse[response[PackageCreateResp]](resp)
	if err != nil {
		return nil, err
	}

	if codeResp.Code == errorcode.OK {
		return &codeResp.Data, nil
	}

	return nil, codeResp.ToError()
}

const PackageCreateLoadPath = "/package/createLoad"

type PackageCreateLoad struct {
	PackageCreateLoadInfo
	Files UploadObjectIterator `json:"-"`
}
type PackageCreateLoadInfo struct {
	UserID   cdssdk.UserID      `json:"userID" binding:"required"`
	BucketID cdssdk.BucketID    `json:"bucketID" binding:"required"`
	Name     string             `json:"name" binding:"required"`
	LoadTo   []cdssdk.StorageID `json:"loadTo" binding:"required"`
}
type PackageCreateLoadResp struct {
	Package    cdssdk.Package  `json:"package"`
	Objects    []cdssdk.Object `json:"objects"`
	LoadedDirs []string        `json:"loadedDirs"`
}

func (c *PackageService) CreateLoad(req PackageCreateLoad) (*PackageCreateLoadResp, error) {
	url, err := url.JoinPath(c.baseURL, PackageCreateLoadPath)
	if err != nil {
		return nil, err
	}

	infoJSON, err := serder.ObjectToJSON(req)
	if err != nil {
		return nil, fmt.Errorf("upload info to json: %w", err)
	}

	resp, err := http2.PostMultiPart(url, http2.MultiPartRequestParam{
		Form: map[string]string{"info": string(infoJSON)},
		Files: iterator.Map(req.Files, func(src *UploadingObject) (*http2.IterMultiPartFile, error) {
			return &http2.IterMultiPartFile{
				FieldName: "files",
				FileName:  src.Path,
				File:      src.File,
			}, nil
		}),
	})
	if err != nil {
		return nil, err
	}

	codeResp, err := ParseJSONResponse[response[PackageCreateLoadResp]](resp)
	if err != nil {
		return nil, err
	}

	if codeResp.Code == errorcode.OK {
		return &codeResp.Data, nil
	}

	return nil, codeResp.ToError()
}

const PackageDeletePath = "/package/delete"

type PackageDelete struct {
	UserID    cdssdk.UserID    `json:"userID" binding:"required"`
	PackageID cdssdk.PackageID `json:"packageID" binding:"required"`
}

func (c *PackageService) Delete(req PackageDelete) error {
	url, err := url.JoinPath(c.baseURL, PackageDeletePath)
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

const PackageListBucketPackagesPath = "/package/listBucketPackages"

type PackageListBucketPackages struct {
	UserID   cdssdk.UserID   `form:"userID" json:"userID" binding:"required"`
	BucketID cdssdk.BucketID `form:"bucketID" json:"bucketID" binding:"required"`
}

type PackageListBucketPackagesResp struct {
	Packages []cdssdk.Package `json:"packages"`
}

func (c *PackageService) ListBucketPackages(req PackageListBucketPackages) (*PackageListBucketPackagesResp, error) {
	url, err := url.JoinPath(c.baseURL, PackageListBucketPackagesPath)
	if err != nil {
		return nil, err
	}

	resp, err := http2.GetForm(url, http2.RequestParam{
		Query: req,
	})
	if err != nil {
		return nil, err
	}

	codeResp, err := ParseJSONResponse[response[PackageListBucketPackagesResp]](resp)
	if err != nil {
		return nil, err
	}

	if codeResp.Code == errorcode.OK {
		return &codeResp.Data, nil
	}

	return nil, codeResp.ToError()
}

const PackageGetCachedStoragesPath = "/package/getCachedStorages"

type PackageGetCachedStoragesReq struct {
	PackageID cdssdk.PackageID `form:"packageID" json:"packageID" binding:"required"`
	UserID    cdssdk.UserID    `form:"userID" json:"userID" binding:"required"`
}

type PackageGetCachedStoragesResp struct {
	cdssdk.PackageCachingInfo
}

func (c *PackageService) GetCachedStorages(req PackageGetCachedStoragesReq) (*PackageGetCachedStoragesResp, error) {
	url, err := url.JoinPath(c.baseURL, PackageGetCachedStoragesPath)
	if err != nil {
		return nil, err
	}
	resp, err := http2.GetJSON(url, http2.RequestParam{
		Query: req,
	})
	if err != nil {
		return nil, err
	}

	codeResp, err := ParseJSONResponse[response[PackageGetCachedStoragesResp]](resp)
	if err != nil {
		return nil, err
	}

	if codeResp.Code == errorcode.OK {
		return &codeResp.Data, nil
	}

	return nil, codeResp.ToError()
}

const PackageGetLoadedStoragesPath = "/package/getLoadedStorages"

type PackageGetLoadedStoragesReq struct {
	PackageID cdssdk.PackageID `form:"packageID" json:"packageID" binding:"required"`
	UserID    cdssdk.UserID    `form:"userID" json:"userID" binding:"required"`
}

type PackageGetLoadedStoragesResp struct {
	StorageIDs []cdssdk.StorageID `json:"storageIDs"`
}

func (c *PackageService) GetLoadedStorages(req PackageGetLoadedStoragesReq) (*PackageGetLoadedStoragesResp, error) {
	url, err := url.JoinPath(c.baseURL, PackageGetLoadedStoragesPath)
	if err != nil {
		return nil, err
	}
	resp, err := http2.GetJSON(url, http2.RequestParam{
		Query: req,
	})
	if err != nil {
		return nil, err
	}

	codeResp, err := ParseJSONResponse[response[PackageGetLoadedStoragesResp]](resp)
	if err != nil {
		return nil, err
	}

	if codeResp.Code == errorcode.OK {
		return &codeResp.Data, nil
	}

	return nil, codeResp.ToError()
}
