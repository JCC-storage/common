package cdsapi

import (
	"net/url"

	"gitlink.org.cn/cloudream/common/consts/errorcode"
	cdssdk "gitlink.org.cn/cloudream/common/sdks/storage"
	"gitlink.org.cn/cloudream/common/utils/http2"
)

var NodeGetNodesPath = "/node/getNodes"

type NodeGetNodesReq struct {
	NodeIDs []cdssdk.NodeID `json:"nodeIDs"`
}

type NodeGetNodesResp struct {
	Nodes []cdssdk.Node `json:"nodes"`
}

func (c *Client) NodeGetNodes(req NodeGetNodesReq) (*NodeGetNodesResp, error) {
	url, err := url.JoinPath(c.baseURL, NodeGetNodesPath)
	if err != nil {
		return nil, err
	}

	resp, err := http2.GetForm(url, http2.RequestParam{
		Query: req,
	})
	if err != nil {
		return nil, err
	}

	jsonResp, err := ParseJSONResponse[response[NodeGetNodesResp]](resp)
	if err != nil {
		return nil, err
	}

	if jsonResp.Code == errorcode.OK {
		return &jsonResp.Data, nil
	}

	return nil, jsonResp.ToError()
}
