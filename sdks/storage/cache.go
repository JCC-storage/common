package stgsdk

import (
	"fmt"
	"net/url"
	"strings"

	"gitlink.org.cn/cloudream/common/consts/errorcode"
	myhttp "gitlink.org.cn/cloudream/common/utils/http"
	"gitlink.org.cn/cloudream/common/utils/serder"
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

	contType := resp.Header.Get("Content-Type")
	if strings.Contains(contType, myhttp.ContentTypeJSON) {
		var codeResp response[CacheMovePackageResp]
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
