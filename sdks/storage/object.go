package cdssdk

import (
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"gitlink.org.cn/cloudream/common/consts/errorcode"
	"gitlink.org.cn/cloudream/common/pkgs/iterator"
	myhttp "gitlink.org.cn/cloudream/common/utils/http"
	"gitlink.org.cn/cloudream/common/utils/serder"
)

type ObjectService struct {
	*Client
}

func (c *Client) Object() *ObjectService {
	return &ObjectService{
		Client: c,
	}
}

const ObjectUploadPath = "/object/upload"

type ObjectUploadReq struct {
	ObjectUploadInfo
	Files UploadObjectIterator `json:"-"`
}

type ObjectUploadInfo struct {
	UserID       UserID    `json:"userID" binding:"required"`
	PackageID    PackageID `json:"packageID" binding:"required"`
	NodeAffinity *NodeID   `json:"nodeAffinity"`
}

type IterObjectUpload struct {
	Path string
	File io.ReadCloser
}

type UploadObjectIterator = iterator.Iterator[*IterObjectUpload]

type ObjectUploadResp struct{}

func (c *ObjectService) Upload(req ObjectUploadReq) (*ObjectUploadResp, error) {
	url, err := url.JoinPath(c.baseURL, ObjectUploadPath)
	if err != nil {
		return nil, err
	}

	infoJSON, err := serder.ObjectToJSON(req)
	if err != nil {
		return nil, fmt.Errorf("upload info to json: %w", err)
	}

	resp, err := myhttp.PostMultiPart(url, myhttp.MultiPartRequestParam{
		Form: map[string]string{"info": string(infoJSON)},
		Files: iterator.Map(req.Files, func(src *IterObjectUpload) (*myhttp.IterMultiPartFile, error) {
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
		var codeResp response[ObjectUploadResp]
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

const ObjectDownloadPath = "/object/download"

type ObjectDownloadReq struct {
	UserID   UserID   `form:"userID" json:"userID" binding:"required"`
	ObjectID ObjectID `form:"objectID" json:"objectID" binding:"required"`
}

func (c *ObjectService) Download(req ObjectDownloadReq) (io.ReadCloser, error) {
	url, err := url.JoinPath(c.baseURL, ObjectDownloadPath)
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

const ObjectUpdateInfoPath = "/object/updateInfo"

type UpdatingObject struct {
	ObjectID   ObjectID  `json:"objectID" binding:"required"`
	UpdateTime time.Time `json:"updateTime" binding:"required"`
}

func (u *UpdatingObject) ApplyTo(obj *Object) {
	obj.UpdateTime = u.UpdateTime
}

type ObjectUpdateInfoReq struct {
	UserID    UserID           `json:"userID" binding:"required"`
	Updatings []UpdatingObject `json:"updatings" binding:"required"`
}

type ObjectUpdateInfoResp struct{}

func (c *ObjectService) Update(req ObjectUpdateInfoReq) (*ObjectUpdateInfoResp, error) {
	url, err := url.JoinPath(c.baseURL, ObjectUpdateInfoPath)
	if err != nil {
		return nil, err
	}

	resp, err := myhttp.PostJSON(url, myhttp.RequestParam{
		Body: req,
	})
	if err != nil {
		return nil, err
	}

	jsonResp, err := myhttp.ParseJSONResponse[response[ObjectUpdateInfoResp]](resp)
	if err != nil {
		return nil, err
	}

	if jsonResp.Code == errorcode.OK {
		return &jsonResp.Data, nil
	}

	return nil, jsonResp.ToError()
}

const ObjectDeletePath = "/object/delete"

type ObjectDeleteReq struct {
	UserID    UserID     `json:"userID" binding:"required"`
	ObjectIDs []ObjectID `json:"objectIDs" binding:"required"`
}

type ObjectDeleteResp struct{}

func (c *ObjectService) Delete(req ObjectDeleteReq) (*ObjectDeleteResp, error) {
	url, err := url.JoinPath(c.baseURL, ObjectDeletePath)
	if err != nil {
		return nil, err
	}

	resp, err := myhttp.PostJSON(url, myhttp.RequestParam{
		Body: req,
	})
	if err != nil {
		return nil, err
	}

	jsonResp, err := myhttp.ParseJSONResponse[response[ObjectDeleteResp]](resp)
	if err != nil {
		return nil, err
	}

	if jsonResp.Code == errorcode.OK {
		return &jsonResp.Data, nil
	}

	return nil, jsonResp.ToError()
}

const ObjectGetPackageObjectsPath = "/object/getPackageObjects"

type ObjectGetPackageObjectsReq struct {
	UserID    UserID    `form:"userID" json:"userID" binding:"required"`
	PackageID PackageID `form:"packageID" json:"packageID" binding:"required"`
}
type ObjectGetPackageObjectsResp struct {
	Objects []Object `json:"objects"`
}

func (c *ObjectService) GetPackageObjects(req ObjectGetPackageObjectsReq) (*ObjectGetPackageObjectsResp, error) {
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
