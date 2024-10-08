package cdssdk

import (
	"fmt"
	"net/url"
	"strings"

	"gitlink.org.cn/cloudream/common/consts/errorcode"
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
	UserID    UserID    `form:"userID" json:"userID" binding:"required"`
	PackageID PackageID `form:"packageID" json:"packageID" binding:"required"`
}
type PackageGetResp struct {
	Package
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
	UserID      UserID `form:"userID" json:"userID" binding:"required"`
	BucketName  string `form:"bucketName" json:"bucketName" binding:"required"`
	PackageName string `form:"packageName" json:"packageName" binding:"required"`
}
type PackageGetByNameResp struct {
	Package Package `json:"package"`
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
	UserID   UserID   `json:"userID"`
	BucketID BucketID `json:"bucketID"`
	Name     string   `json:"name"`
}

type PackageCreateResp struct {
	Package Package `json:"package"`
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

const PackageDeletePath = "/package/delete"

type PackageDelete struct {
	UserID    UserID    `json:"userID" binding:"required"`
	PackageID PackageID `json:"packageID" binding:"required"`
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
	UserID   UserID   `form:"userID" json:"userID" binding:"required"`
	BucketID BucketID `form:"bucketID" json:"bucketID" binding:"required"`
}

type PackageListBucketPackagesResp struct {
	Packages []Package `json:"packages"`
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

const PackageGetCachedNodesPath = "/package/getCachedNodes"

type PackageGetCachedNodesReq struct {
	PackageID PackageID `form:"packageID" json:"packageID" binding:"required"`
	UserID    UserID    `form:"userID" json:"userID" binding:"required"`
}

type PackageGetCachedNodesResp struct {
	PackageCachingInfo
}

func (c *PackageService) GetCachedNodes(req PackageGetCachedNodesReq) (*PackageGetCachedNodesResp, error) {
	url, err := url.JoinPath(c.baseURL, PackageGetCachedNodesPath)
	if err != nil {
		return nil, err
	}
	resp, err := http2.GetJSON(url, http2.RequestParam{
		Query: req,
	})
	if err != nil {
		return nil, err
	}

	codeResp, err := ParseJSONResponse[response[PackageGetCachedNodesResp]](resp)
	if err != nil {
		return nil, err
	}

	if codeResp.Code == errorcode.OK {
		return &codeResp.Data, nil
	}

	return nil, codeResp.ToError()
}

const PackageGetLoadedNodesPath = "/package/getLoadedNodes"

type PackageGetLoadedNodesReq struct {
	PackageID PackageID `form:"packageID" json:"packageID" binding:"required"`
	UserID    UserID    `form:"userID" json:"userID" binding:"required"`
}

type PackageGetLoadedNodesResp struct {
	NodeIDs []NodeID `json:"nodeIDs"`
}

func (c *PackageService) GetLoadedNodes(req PackageGetLoadedNodesReq) (*PackageGetLoadedNodesResp, error) {
	url, err := url.JoinPath(c.baseURL, PackageGetLoadedNodesPath)
	if err != nil {
		return nil, err
	}
	resp, err := http2.GetJSON(url, http2.RequestParam{
		Query: req,
	})
	if err != nil {
		return nil, err
	}

	codeResp, err := ParseJSONResponse[response[PackageGetLoadedNodesResp]](resp)
	if err != nil {
		return nil, err
	}

	if codeResp.Code == errorcode.OK {
		return &codeResp.Data, nil
	}

	return nil, codeResp.ToError()
}
