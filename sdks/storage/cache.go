package stgsdk

import (
	"net/url"

	"gitlink.org.cn/cloudream/common/consts/errorcode"
	myhttp "gitlink.org.cn/cloudream/common/utils/http"
)

type CacheMovePackageReq struct {
	UserID    int64 `json:"userID"`
	PackageID int64 `json:"packageID"`
	NodeID    int64 `json:"nodeID"`
}
type CacheMovePackageResp struct {
	CacheInfos []ObjectCacheInfo `json:"cacheInfos"`
}

func (c *Client) CacheMovePackage(req CacheMovePackageReq) (*CacheMovePackageResp, error) {
	url, err := url.JoinPath(c.baseURL, "/cache/movePackage")
	if err != nil {
		return nil, err
	}

	resp, err := myhttp.PostJSON(url, myhttp.RequestParam{
		Body: req,
	})
	if err != nil {
		return nil, err
	}

	jsonResp, err := myhttp.ParseJSONResponse[response[CacheMovePackageResp]](resp)
	if err != nil {
		return nil, err
	}

	if jsonResp.Code == errorcode.OK {
		return &jsonResp.Data, nil
	}

	return nil, jsonResp.ToError()
}

type GetPackageObjectCacheInfosReq struct {
	UserID    int64 `json:"userID"`
	PackageID int64 `json:"packageID"`
}
type GetPackageObjectCacheInfosResp struct {
	Infos []ObjectCacheInfo `json:"cacheInfos"`
}

func (c *Client) GetPackageObjectCacheInfos(req GetPackageObjectCacheInfosReq) (*GetPackageObjectCacheInfosResp, error) {
	url, err := url.JoinPath(c.baseURL, "/cache/getPackageObjectCacheInfos")
	if err != nil {
		return nil, err
	}

	resp, err := myhttp.GetForm(url, myhttp.RequestParam{
		Query: req,
	})
	if err != nil {
		return nil, err
	}

	jsonResp, err := myhttp.ParseJSONResponse[response[GetPackageObjectCacheInfosResp]](resp)
	if err != nil {
		return nil, err
	}

	if jsonResp.Code == errorcode.OK {
		return &jsonResp.Data, nil
	}

	return nil, jsonResp.ToError()
}
