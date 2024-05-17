package cdssdk

import (
	"net/url"

	"gitlink.org.cn/cloudream/common/consts/errorcode"
	myhttp "gitlink.org.cn/cloudream/common/utils/http"
)

var NodeGetNodesPath = "/node/getNodes"

type NodeGetNodesReq struct {
	NodeIDs []NodeID `json:"nodeIDs"`
}

type NodeGetNodesResp struct {
	Nodes []Node `json:"nodes"`
}

func (c *Client) NodeGetNodes(req NodeGetNodesReq) (*NodeGetNodesResp, error) {
	url, err := url.JoinPath(c.baseURL, NodeGetNodesPath)
	if err != nil {
		return nil, err
	}

	resp, err := myhttp.GetForm(url, myhttp.RequestParam{
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
