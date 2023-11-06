package imsdk

import (
	"net/url"

	"gitlink.org.cn/cloudream/common/consts/errorcode"
	cdssdk "gitlink.org.cn/cloudream/common/sdks/storage"
	myhttp "gitlink.org.cn/cloudream/common/utils/http"
)

const PackageGetWithObjectCacheInfosPath = "/package/getWithObjectCacheInfos"

type PackageGetWithObjectCacheInfos struct {
	UserID    int64 `json:"userID"`
	PackageID int64 `json:"packageID"`
}
type PackageGetWithObjectCacheInfosResp struct {
	Package          cdssdk.Package           `json:"package"`
	ObjectCacheInfos []cdssdk.ObjectCacheInfo `json:"objectCacheInfos"`
}

func (c *Client) PackageGetWithObjectCacheInfos(req PackageGetWithObjectCacheInfos) (*PackageGetWithObjectCacheInfosResp, error) {
	url, err := url.JoinPath(c.baseURL, PackageGetWithObjectCacheInfosPath)
	if err != nil {
		return nil, err
	}

	resp, err := myhttp.GetForm(url, myhttp.RequestParam{
		Query: req,
	})
	if err != nil {
		return nil, err
	}

	jsonResp, err := myhttp.ParseJSONResponse[response[PackageGetWithObjectCacheInfosResp]](resp)
	if err != nil {
		return nil, err
	}

	if jsonResp.Code == errorcode.OK {
		return &jsonResp.Data, nil
	}

	return nil, jsonResp.ToError()
}
