package imsdk

import (
	"net/url"

	"gitlink.org.cn/cloudream/common/consts/errorcode"
	schsdk "gitlink.org.cn/cloudream/common/sdks/scheduler"
	myhttp "gitlink.org.cn/cloudream/common/utils/http"
)

const ProxyGetServiceInfoPath = "/proxy/getServiceInfo"

type ProxyGetServiceInfo struct {
	ServiceName string          `json:"serviceName"`
	JobSetID    schsdk.JobSetID `json:"jobSetID"`
}

type ProxyGetServiceInfoResp struct {
	LocalJobID string `json:"localJobID"`
}

func (c *Client) ProxyGetServiceInfo(req ProxyGetServiceInfo) (*ProxyGetServiceInfoResp, error) {
	url, err := url.JoinPath(c.baseURL, ProxyGetServiceInfoPath)
	if err != nil {
		return nil, err
	}

	resp, err := myhttp.GetForm(url, myhttp.RequestParam{
		Query: req,
	})
	if err != nil {
		return nil, err
	}

	jsonResp, err := myhttp.ParseJSONResponse[response[ProxyGetServiceInfoResp]](resp)
	if err != nil {
		return nil, err
	}

	if jsonResp.Code == errorcode.OK {
		return &jsonResp.Data, nil
	}

	return nil, jsonResp.ToError()
}
