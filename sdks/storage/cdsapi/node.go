package cdsapi

import (
	"net/url"

	"gitlink.org.cn/cloudream/common/consts/errorcode"
	cdssdk "gitlink.org.cn/cloudream/common/sdks/storage"
	"gitlink.org.cn/cloudream/common/utils/http2"
)

var HubGetHubsPath = "/hub/getHubs"

type HubGetHubsReq struct {
	HubIDs []cdssdk.HubID `json:"hubIDs"`
}

type HubGetHubsResp struct {
	Hubs []cdssdk.Hub `json:"hubs"`
}

func (c *Client) HubGetHubs(req HubGetHubsReq) (*HubGetHubsResp, error) {
	url, err := url.JoinPath(c.baseURL, HubGetHubsPath)
	if err != nil {
		return nil, err
	}

	resp, err := http2.GetForm(url, http2.RequestParam{
		Query: req,
	})
	if err != nil {
		return nil, err
	}

	jsonResp, err := ParseJSONResponse[response[HubGetHubsResp]](resp)
	if err != nil {
		return nil, err
	}

	if jsonResp.Code == errorcode.OK {
		return &jsonResp.Data, nil
	}

	return nil, jsonResp.ToError()
}
