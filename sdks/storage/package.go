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
	UserID    UserID    `json:"userID"`
	PackageID PackageID `json:"packageID"`
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

type PackageDeleteReq struct {
	UserID    UserID    `json:"userID"`
	PackageID PackageID `json:"packageID"`
}

func (c *PackageService) Delete(req PackageDeleteReq) error {
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

type PackageGetCachedNodesReq struct {
	PackageID PackageID `json:"packageID"`
	UserID    UserID    `json:"userID"`
}

type PackageGetCachedNodesResp struct {
	PackageCachingInfo
}

func (c *PackageService) GetCachedNodes(req PackageGetCachedNodesReq) (*PackageGetCachedNodesResp, error) {
	url, err := url.JoinPath(c.baseURL, "/package/getCachedNodes")
	if err != nil {
		return nil, err
	}
	resp, err := myhttp.GetJSON(url, myhttp.RequestParam{
		Body: req,
	})
	if err != nil {
		return nil, err
	}
	contType := resp.Header.Get("Content-Type")
	if strings.Contains(contType, myhttp.ContentTypeJSON) {

		var codeResp response[PackageGetCachedNodesResp]
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

type PackageGetLoadedNodesReq struct {
	PackageID PackageID `json:"packageID"`
	UserID    UserID    `json:"userID"`
}

type PackageGetLoadedNodesResp struct {
	NodeIDs []NodeID `json:"nodeIDs"`
}

func (c *PackageService) GetLoadedNodes(req PackageGetLoadedNodesReq) (*PackageGetLoadedNodesResp, error) {
	url, err := url.JoinPath(c.baseURL, "/package/getLoadedNodes")
	if err != nil {
		return nil, err
	}
	resp, err := myhttp.GetJSON(url, myhttp.RequestParam{
		Body: req,
	})
	if err != nil {
		return nil, err
	}

	contType := resp.Header.Get("Content-Type")
	if strings.Contains(contType, myhttp.ContentTypeJSON) {

		var codeResp response[PackageGetLoadedNodesResp]
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
