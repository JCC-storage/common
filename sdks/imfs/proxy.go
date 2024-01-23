package imsdk

import (
	"net/url"

	"gitlink.org.cn/cloudream/common/consts/errorcode"
	schsdk "gitlink.org.cn/cloudream/common/sdks/scheduler"
	myhttp "gitlink.org.cn/cloudream/common/utils/http"
)

const ProxyGetJobIDPath = "/proxy/getJobID"

type ProxyGetJobID struct {
	ServiceName string          `json:"serviceName"`
	JobSetID    schsdk.JobSetID `json:"jobSetID"`
}

type ProxyGetJobIDResp struct {
	LocalJobID string `json:"localJobID"`
}

func (c *Client) ProxyGetJobID(req ProxyGetJobID) (*ProxyGetJobIDResp, error) {
	url, err := url.JoinPath(c.baseURL, ProxyGetJobIDPath)
	if err != nil {
		return nil, err
	}

	resp, err := myhttp.GetForm(url, myhttp.RequestParam{
		Query: req,
	})
	if err != nil {
		return nil, err
	}

	jsonResp, err := myhttp.ParseJSONResponse[response[ProxyGetJobIDResp]](resp)
	if err != nil {
		return nil, err
	}

	if jsonResp.Code == errorcode.OK {
		return &jsonResp.Data, nil
	}

	return nil, jsonResp.ToError()
}
