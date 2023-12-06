package cdssdk

import (
	"fmt"
	"io"
	"net/url"
	"strings"

	"gitlink.org.cn/cloudream/common/consts/errorcode"
	myhttp "gitlink.org.cn/cloudream/common/utils/http"
	"gitlink.org.cn/cloudream/common/utils/serder"
)

type ObjectDownloadReq struct {
	UserID   int64 `json:"userID"`
	ObjectID int64 `json:"objectID"`
}

func (c *Client) ObjectDownload(req ObjectDownloadReq) (io.ReadCloser, error) {
	url, err := url.JoinPath(c.baseURL, "/object/download")
	if err != nil {
		return nil, err
	}

	resp, err := myhttp.GetJSON(url, myhttp.RequestParam{
		Query: req,
	})
	if err != nil {
		return nil, err
	}

	contType := resp.Header.Get("Content-Type")

	if strings.Contains(contType, myhttp.ContentTypeJSON) {
		var codeResp response[any]
		if err := serder.JSONToObjectStream(resp.Body, &codeResp); err != nil {
			return nil, fmt.Errorf("parsing response: %w", err)
		}

		return nil, codeResp.ToError()
	}

	if strings.Contains(contType, myhttp.ContentTypeOctetStream) {
		return resp.Body, nil
	}

	return nil, fmt.Errorf("unknow response content type: %s", contType)
}

var ObjectGetPackageObjectsPath = "/object/getPackageObjects"

type ObjectGetPackageObjectsReq struct {
	UserID    UserID    `json:"userID"`
	PackageID PackageID `json:"packageID"`
}
type ObjectGetPackageObjectsResp struct {
	Objects []Object `json:"objects"`
}

func (c *Client) ObjectGetPackageObjects(req ObjectGetPackageObjectsReq) (*ObjectGetPackageObjectsResp, error) {
	url, err := url.JoinPath(c.baseURL, ObjectGetPackageObjectsPath)
	if err != nil {
		return nil, err
	}

	resp, err := myhttp.GetForm(url, myhttp.RequestParam{
		Query: req,
	})
	if err != nil {
		return nil, err
	}

	jsonResp, err := myhttp.ParseJSONResponse[response[ObjectGetPackageObjectsResp]](resp)
	if err != nil {
		return nil, err
	}

	if jsonResp.Code == errorcode.OK {
		return &jsonResp.Data, nil
	}

	return nil, jsonResp.ToError()
}
