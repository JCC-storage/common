package imsdk

import (
	"net/url"

	"gitlink.org.cn/cloudream/common/consts/errorcode"
	schsdk "gitlink.org.cn/cloudream/common/sdks/scheduler"
	"gitlink.org.cn/cloudream/common/utils/http2"
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

	resp, err := http2.GetForm(url, http2.RequestParam{
		Query: req,
	})
	if err != nil {
		return nil, err
	}

	jsonResp, err := http2.ParseJSONResponse[response[ProxyGetServiceInfoResp]](resp)
	if err != nil {
		return nil, err
	}

	if jsonResp.Code == errorcode.OK {
		return &jsonResp.Data, nil
	}

	return nil, jsonResp.ToError()
}
