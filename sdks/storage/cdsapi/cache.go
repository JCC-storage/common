package cdsapi

import (
	"net/url"

	"gitlink.org.cn/cloudream/common/consts/errorcode"
	cdssdk "gitlink.org.cn/cloudream/common/sdks/storage"
	"gitlink.org.cn/cloudream/common/utils/http2"
)

const CacheMovePackagePath = "/cache/movePackage"

type CacheMovePackageReq struct {
	UserID    cdssdk.UserID    `json:"userID"`
	PackageID cdssdk.PackageID `json:"packageID"`
	StorageID cdssdk.StorageID `json:"storageID"`
}
type CacheMovePackageResp struct{}

func (c *Client) CacheMovePackage(req CacheMovePackageReq) (*CacheMovePackageResp, error) {
	url, err := url.JoinPath(c.baseURL, CacheMovePackagePath)
	if err != nil {
		return nil, err
	}

	resp, err := http2.PostJSON(url, http2.RequestParam{
		Body: req,
	})
	if err != nil {
		return nil, err
	}

	jsonResp, err := ParseJSONResponse[response[CacheMovePackageResp]](resp)
	if err != nil {
		return nil, err
	}

	if jsonResp.Code == errorcode.OK {
		return &jsonResp.Data, nil
	}

	return nil, jsonResp.ToError()
}
