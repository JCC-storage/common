package blockchain

import (
	"gitlink.org.cn/cloudream/common/utils/http2"
	"net/url"
)

type InvokeReq struct {
	ContractAddress string   `json:"contractAddress"`
	FunctionName    string   `json:"functionName"`
	MemberName      string   `json:"memberName"`
	Type            string   `json:"type"`
	Args            []string `json:"args"`
}

func (c *Client) BlockChainInvoke(req InvokeReq) error {
	targetUrl, err := url.JoinPath(c.baseURL, "/contract/invoke")
	if err != nil {
		return err
	}

	header := make(map[string]string)
	header["Content-Type"] = http2.ContentTypeJSON

	_, err = http2.PostJSON(targetUrl, http2.RequestParam{
		Body:   req,
		Header: header,
	})
	if err != nil {
		return err
	}

	return nil
}
