package cdssdk

import (
	"fmt"
	"io"
	"net/url"
	"strings"

	"gitlink.org.cn/cloudream/common/consts/errorcode"
	"gitlink.org.cn/cloudream/common/pkgs/iterator"
	myhttp "gitlink.org.cn/cloudream/common/utils/http"
	"gitlink.org.cn/cloudream/common/utils/serder"
)

type PackageGetReq struct {
	UserID    UserID    `json:"userID"`
	PackageID PackageID `json:"packageID"`
}
type PackageGetResp struct {
	Package
}

func (c *Client) PackageGet(req PackageGetReq) (*PackageGetResp, error) {
	url, err := url.JoinPath(c.baseURL, "/package/get")
	if err != nil {
		return nil, err
	}

	resp, err := myhttp.GetForm(url, myhttp.RequestParam{
		Query: req,
	})
	if err != nil {
		return nil, err
	}

	codeResp, err := myhttp.ParseJSONResponse[response[PackageGetResp]](resp)
	if err != nil {
		return nil, err
	}

	if codeResp.Code == errorcode.OK {
		return &codeResp.Data, nil
	}

	return nil, codeResp.ToError()
}

type PackageUploadReq struct {
	UserID       UserID                    `json:"userID"`
	BucketID     BucketID                  `json:"bucketID"`
	Name         string                    `json:"name"`
	NodeAffinity *NodeID                   `json:"nodeAffinity"`
	Files        PackageUploadFileIterator `json:"-"`
}

type IterPackageUploadFile struct {
	Path string
	File io.ReadCloser
}

type PackageUploadFileIterator = iterator.Iterator[*IterPackageUploadFile]

type PackageUploadResp struct {
	PackageID PackageID `json:"packageID,string"`
}

func (c *Client) PackageUpload(req PackageUploadReq) (*PackageUploadResp, error) {
	url, err := url.JoinPath(c.baseURL, "/package/upload")
	if err != nil {
		return nil, err
	}

	infoJSON, err := serder.ObjectToJSON(req)
	if err != nil {
		return nil, fmt.Errorf("package info to json: %w", err)
	}

	resp, err := myhttp.PostMultiPart(url, myhttp.MultiPartRequestParam{
		Form: map[string]string{"info": string(infoJSON)},
		Files: iterator.Map(req.Files, func(src *IterPackageUploadFile) (*myhttp.IterMultiPartFile, error) {
			return &myhttp.IterMultiPartFile{
				FieldName: "files",
				FileName:  src.Path,
				File:      src.File,
			}, nil
		}),
	})
	if err != nil {
		return nil, err
	}

	contType := resp.Header.Get("Content-Type")
	if strings.Contains(contType, myhttp.ContentTypeJSON) {
		var codeResp response[PackageUploadResp]
		if err := serder.JSONToObjectStream(resp.Body, &codeResp); err != nil {
			return nil, fmt.Errorf("parsing response: %w", err)
		}

		if codeResp.Code == errorcode.OK {
			return &codeResp.Data, nil
		}

		return nil, codeResp.ToError()
	}

	return nil, fmt.Errorf("unknow response content type: %s", contType)

}

type PackageDeleteReq struct {
	UserID    UserID    `json:"userID"`
	PackageID PackageID `json:"packageID"`
}

func (c *Client) PackageDelete(req PackageDeleteReq) error {
	url, err := url.JoinPath(c.baseURL, "/package/delete")
	if err != nil {
		return err
	}

	resp, err := myhttp.PostJSON(url, myhttp.RequestParam{
		Body: req,
	})
	if err != nil {
		return err
	}

	contType := resp.Header.Get("Content-Type")

	if strings.Contains(contType, myhttp.ContentTypeJSON) {
		var codeResp response[any]
		if err := serder.JSONToObjectStream(resp.Body, &codeResp); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		if codeResp.Code == errorcode.OK {
			return nil
		}

		return codeResp.ToError()
	}

	return fmt.Errorf("unknow response content type: %s", contType)
}

type PackageGetCachedNodesReq struct {
	PackageID PackageID `json:"packageID"`
	UserID    UserID    `json:"userID"`
}

type PackageGetCachedNodesResp struct {
	PackageCachingInfo
}

func (c *Client) PackageGetCachedNodes(req PackageGetCachedNodesReq) (*PackageGetCachedNodesResp, error) {
	url, err := url.JoinPath(c.baseURL, "/package/getCachedNodes")
	if err != nil {
		return nil, err
	}
	resp, err := myhttp.GetJSON(url, myhttp.RequestParam{
		Body: req,
	})
	if err != nil {
		return nil, err
	}
	contType := resp.Header.Get("Content-Type")
	if strings.Contains(contType, myhttp.ContentTypeJSON) {

		var codeResp response[PackageGetCachedNodesResp]
		if err := serder.JSONToObjectStream(resp.Body, &codeResp); err != nil {
			return nil, fmt.Errorf("parsing response: %w", err)
		}

		if codeResp.Code == errorcode.OK {
			return &codeResp.Data, nil
		}

		return nil, codeResp.ToError()
	}

	return nil, fmt.Errorf("unknow response content type: %s", contType)
}

type PackageGetLoadedNodesReq struct {
	PackageID PackageID `json:"packageID"`
	UserID    UserID    `json:"userID"`
}

type PackageGetLoadedNodesResp struct {
	NodeIDs []NodeID `json:"nodeIDs"`
}

func (c *Client) PackageGetLoadedNodes(req PackageGetLoadedNodesReq) (*PackageGetLoadedNodesResp, error) {
	url, err := url.JoinPath(c.baseURL, "/package/getLoadedNodes")
	if err != nil {
		return nil, err
	}
	resp, err := myhttp.GetJSON(url, myhttp.RequestParam{
		Body: req,
	})
	if err != nil {
		return nil, err
	}

	contType := resp.Header.Get("Content-Type")
	if strings.Contains(contType, myhttp.ContentTypeJSON) {

		var codeResp response[PackageGetLoadedNodesResp]
		if err := serder.JSONToObjectStream(resp.Body, &codeResp); err != nil {
			return nil, fmt.Errorf("parsing response: %w", err)
		}

		if codeResp.Code == errorcode.OK {
			return &codeResp.Data, nil
		}

		return nil, codeResp.ToError()
	}

	return nil, fmt.Errorf("unknow response content type: %s", contType)
}
