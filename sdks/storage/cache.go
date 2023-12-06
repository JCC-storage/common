package cdssdk

import (
	"net/url"

	"gitlink.org.cn/cloudream/common/consts/errorcode"
	myhttp "gitlink.org.cn/cloudream/common/utils/http"
)

var CacheMovePackagePath = "/cache/movePackage"

type CacheMovePackageReq struct {
	UserID    UserID    `json:"userID"`
	PackageID PackageID `json:"packageID"`
	NodeID    NodeID    `json:"nodeID"`
}
type CacheMovePackageResp struct{}

func (c *Client) CacheMovePackage(req CacheMovePackageReq) (*CacheMovePackageResp, error) {
	url, err := url.JoinPath(c.baseURL, CacheMovePackagePath)
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
