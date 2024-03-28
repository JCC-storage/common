package cdssdk

import (
	"fmt"
	"net/url"
	"strings"

	"gitlink.org.cn/cloudream/common/consts/errorcode"
	myhttp "gitlink.org.cn/cloudream/common/utils/http"
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

	resp, err := myhttp.GetForm(url, myhttp.RequestParam{
		Query: req,
	})
	if err != nil {
		return nil, err
	}

	codeResp, err := myhttp.ParseJSONResponse[response[PackageGetResp]](resp)
	if err != nil {
		return nil, err
	}

	if codeResp.Code == errorcode.OK {
		return &codeResp.Data, nil
	}

	return nil, codeResp.ToError()
}

const PackageCreatePath = "/package/create"

type PackageCreateReq struct {
	UserID   UserID   `json:"userID"`
	BucketID BucketID `json:"bucketID"`
	Name     string   `json:"name"`
}

type PackageCreateResp struct {
	PackageID PackageID `json:"packageID,string"`
}

func (s *PackageService) Create(req PackageCreateReq) (*PackageCreateResp, error) {
	url, err := url.JoinPath(s.baseURL, PackageCreatePath)
	if err != nil {
		return nil, err
	}

	resp, err := myhttp.PostJSON(url, myhttp.RequestParam{
		Body: req,
	})
	if err != nil {
		return nil, err
	}

	codeResp, err := myhttp.ParseJSONResponse[response[PackageCreateResp]](resp)
	if err != nil {
		return nil, err
	}

	if codeResp.Code == errorcode.OK {
		return &codeResp.Data, nil
	}

	return nil, codeResp.ToError()
}

const PackageDeletePath = "/package/delete"

type PackageDeleteReq struct {
	UserID    UserID    `json:"userID" binding:"required"`
	PackageID PackageID `json:"packageID" binding:"required"`
}

func (c *PackageService) Delete(req PackageDeleteReq) error {
	url, err := url.JoinPath(c.baseURL, PackageDeletePath)
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

const PackageListBucketPackagesPath = "/package/listBucketPackages"

type PackageListBucketPackagesReq struct {
	UserID   UserID   `form:"userID" json:"userID" binding:"required"`
	BucketID BucketID `form:"bucketID" json:"bucketID" binding:"required"`
}

type PackageListBucketPackagesResp struct {
	Packages []Package `json:"packages"`
}

func (c *PackageService) ListBucketPackages(req PackageListBucketPackagesReq) (*PackageListBucketPackagesResp, error) {
	url, err := url.JoinPath(c.baseURL, PackageListBucketPackagesPath)
	if err != nil {
		return nil, err
	}

	resp, err := myhttp.GetForm(url, myhttp.RequestParam{
		Query: req,
	})
	if err != nil {
		return nil, err
	}

	codeResp, err := myhttp.ParseJSONResponse[response[PackageListBucketPackagesResp]](resp)
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
	resp, err := myhttp.GetJSON(url, myhttp.RequestParam{
		Body: req,
	})
	if err != nil {
		return nil, err
	}

	codeResp, err := myhttp.ParseJSONResponse[response[PackageGetCachedNodesResp]](resp)
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
	resp, err := myhttp.GetJSON(url, myhttp.RequestParam{
		Body: req,
	})
	if err != nil {
		return nil, err
	}

	codeResp, err := myhttp.ParseJSONResponse[response[PackageGetLoadedNodesResp]](resp)
	if err != nil {
		return nil, err
	}

	if codeResp.Code == errorcode.OK {
		return &codeResp.Data, nil
	}

	return nil, codeResp.ToError()
}
