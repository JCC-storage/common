package imsdk

import (
	"fmt"
	"io"
	"net/url"
	"strings"

	"gitlink.org.cn/cloudream/common/utils/http2"
	"gitlink.org.cn/cloudream/common/utils/serder"
)

const IPFSReadPath = "/ipfs/read"

type IPFSRead struct {
	FileHash string `json:"fileHash"`
	Offset   int64  `json:"offset"`
	Length   int64  `json:"length"`
}

func (c *Client) IPFSRead(req IPFSRead) (io.ReadCloser, error) {
	url, err := url.JoinPath(c.baseURL, IPFSReadPath)
	if err != nil {
		return nil, err
	}

	resp, err := http2.GetForm(url, http2.RequestParam{
		Query: req,
	})
	if err != nil {
		return nil, err
	}

	contType := resp.Header.Get("Content-Type")

	if strings.Contains(contType, http2.ContentTypeJSON) {
		var codeResp response[any]
		if err := serder.JSONToObjectStream(resp.Body, &codeResp); err != nil {
			return nil, fmt.Errorf("parsing response: %w", err)
		}

		return nil, codeResp.ToError()
	}

	if strings.Contains(contType, http2.ContentTypeOctetStream) {
		return resp.Body, nil
	}

	return nil, fmt.Errorf("unknow response content type: %s", contType)
}
