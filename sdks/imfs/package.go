package imsdk

import (
	"net/url"

	"gitlink.org.cn/cloudream/common/consts/errorcode"
	cdssdk "gitlink.org.cn/cloudream/common/sdks/storage"
	myhttp "gitlink.org.cn/cloudream/common/utils/http"
)

const PackageGetWithObjectsPath = "/package/getWithObjects"

type PackageGetWithObjectsInfos struct {
	UserID    int64 `json:"userID"`
	PackageID int64 `json:"packageID"`
}
type PackageGetWithObjectsResp struct {
	Package cdssdk.Package  `json:"package"`
	Objects []cdssdk.Object `json:"objects"`
}

func (c *Client) PackageGetWithObjects(req PackageGetWithObjectsInfos) (*PackageGetWithObjectsResp, error) {
	url, err := url.JoinPath(c.baseURL, PackageGetWithObjectsPath)
	if err != nil {
		return nil, err
	}

	resp, err := myhttp.GetForm(url, myhttp.RequestParam{
		Query: req,
	})
	if err != nil {
		return nil, err
	}

	jsonResp, err := myhttp.ParseJSONResponse[response[PackageGetWithObjectsResp]](resp)
	if err != nil {
		return nil, err
	}

	if jsonResp.Code == errorcode.OK {
		return &jsonResp.Data, nil
	}

	return nil, jsonResp.ToError()
}
