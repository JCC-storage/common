package cdsapi

import (
	"fmt"
	"io"
	"mime"
	"net/url"
	"strings"
	"time"

	"gitlink.org.cn/cloudream/common/consts/errorcode"
	"gitlink.org.cn/cloudream/common/pkgs/iterator"
	cdssdk "gitlink.org.cn/cloudream/common/sdks/storage"
	"gitlink.org.cn/cloudream/common/utils/http2"
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

const ObjectListPath = "/object/list"

type ObjectList struct {
	UserID    cdssdk.UserID    `form:"userID" binding:"required"`
	PackageID cdssdk.PackageID `form:"packageID" binding:"required"`
	Path      string           `form:"path"` // 允许为空字符串
	IsPrefix  bool             `form:"isPrefix"`
}
type ObjectListResp struct {
	Objects []cdssdk.Object `json:"objects"`
}

func (c *ObjectService) List(req ObjectList) (*ObjectListResp, error) {
	url, err := url.JoinPath(c.baseURL, ObjectListPath)
	if err != nil {
		return nil, err
	}

	resp, err := http2.GetForm(url, http2.RequestParam{
		Query: req,
	})
	if err != nil {
		return nil, err
	}

	jsonResp, err := ParseJSONResponse[response[ObjectListResp]](resp)
	if err != nil {
		return nil, err
	}

	if jsonResp.Code == errorcode.OK {
		return &jsonResp.Data, nil
	}

	return nil, jsonResp.ToError()
}

const ObjectUploadPath = "/object/upload"

type ObjectUpload struct {
	ObjectUploadInfo
	Files UploadObjectIterator `json:"-"`
}

type ObjectUploadInfo struct {
	UserID    cdssdk.UserID      `json:"userID" binding:"required"`
	PackageID cdssdk.PackageID   `json:"packageID" binding:"required"`
	Affinity  cdssdk.StorageID   `json:"affinity"`
	LoadTo    []cdssdk.StorageID `json:"loadTo"`
}

type UploadingObject struct {
	Path string
	File io.ReadCloser
}

type UploadObjectIterator = iterator.Iterator[*UploadingObject]

type ObjectUploadResp struct {
	Uploadeds []cdssdk.Object `json:"uploadeds"`
}

func (c *ObjectService) Upload(req ObjectUpload) (*ObjectUploadResp, error) {
	url, err := url.JoinPath(c.baseURL, ObjectUploadPath)
	if err != nil {
		return nil, err
	}

	infoJSON, err := serder.ObjectToJSON(req)
	if err != nil {
		return nil, fmt.Errorf("upload info to json: %w", err)
	}

	resp, err := http2.PostMultiPart(url, http2.MultiPartRequestParam{
		Form: map[string]string{"info": string(infoJSON)},
		Files: iterator.Map(req.Files, func(src *UploadingObject) (*http2.IterMultiPartFile, error) {
			return &http2.IterMultiPartFile{
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
	if strings.Contains(contType, http2.ContentTypeJSON) {
		var err error
		var codeResp response[ObjectUploadResp]
		if codeResp, err = serder.JSONToObjectStreamEx[response[ObjectUploadResp]](resp.Body); err != nil {
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

type ObjectDownload struct {
	UserID   cdssdk.UserID   `form:"userID" json:"userID" binding:"required"`
	ObjectID cdssdk.ObjectID `form:"objectID" json:"objectID" binding:"required"`
	Offset   int64           `form:"offset" json:"offset,omitempty"`
	Length   *int64          `form:"length" json:"length,omitempty"`
}
type DownloadingObject struct {
	Path string
	File io.ReadCloser
}

func (c *ObjectService) Download(req ObjectDownload) (*DownloadingObject, error) {
	url, err := url.JoinPath(c.baseURL, ObjectDownloadPath)
	if err != nil {
		return nil, err
	}

	resp, err := http2.GetJSON(url, http2.RequestParam{
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

	_, params, err := mime.ParseMediaType(resp.Header.Get("Content-Disposition"))
	if err != nil {
		return nil, fmt.Errorf("parsing content disposition: %w", err)
	}

	return &DownloadingObject{
		Path: params["filename"],
		File: resp.Body,
	}, nil
}

const ObjectDownloadByPathPath = "/object/downloadByPath"

type ObjectDownloadByPath struct {
	UserID    cdssdk.UserID    `form:"userID" json:"userID" binding:"required"`
	PackageID cdssdk.PackageID `form:"packageID" json:"packageID" binding:"required"`
	Path      string           `form:"path" json:"path" binding:"required"`
	Offset    int64            `form:"offset" json:"offset,omitempty"`
	Length    *int64           `form:"length" json:"length,omitempty"`
}

func (c *ObjectService) DownloadByPath(req ObjectDownloadByPath) (*DownloadingObject, error) {
	url, err := url.JoinPath(c.baseURL, ObjectDownloadByPathPath)
	if err != nil {
		return nil, err
	}

	resp, err := http2.GetJSON(url, http2.RequestParam{
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

	_, params, err := mime.ParseMediaType(resp.Header.Get("Content-Disposition"))
	if err != nil {
		return nil, fmt.Errorf("parsing content disposition: %w", err)
	}

	return &DownloadingObject{
		Path: params["filename"],
		File: resp.Body,
	}, nil
}

const ObjectUpdateInfoPath = "/object/updateInfo"

type UpdatingObject struct {
	ObjectID   cdssdk.ObjectID `json:"objectID" binding:"required"`
	UpdateTime time.Time       `json:"updateTime" binding:"required"`
}

func (u *UpdatingObject) ApplyTo(obj *cdssdk.Object) {
	obj.UpdateTime = u.UpdateTime
}

type ObjectUpdateInfo struct {
	UserID    cdssdk.UserID    `json:"userID" binding:"required"`
	Updatings []UpdatingObject `json:"updatings" binding:"required"`
}

type ObjectUpdateInfoResp struct {
	Successes []cdssdk.ObjectID `json:"successes"`
}

func (c *ObjectService) UpdateInfo(req ObjectUpdateInfo) (*ObjectUpdateInfoResp, error) {
	url, err := url.JoinPath(c.baseURL, ObjectUpdateInfoPath)
	if err != nil {
		return nil, err
	}

	resp, err := http2.PostJSON(url, http2.RequestParam{
		Body: req,
	})
	if err != nil {
		return nil, err
	}

	jsonResp, err := ParseJSONResponse[response[ObjectUpdateInfoResp]](resp)
	if err != nil {
		return nil, err
	}

	if jsonResp.Code == errorcode.OK {
		return &jsonResp.Data, nil
	}

	return nil, jsonResp.ToError()
}

const ObjectUpdateInfoByPathPath = "/object/updateInfoByPath"

type ObjectUpdateInfoByPath struct {
	UserID     cdssdk.UserID    `json:"userID" binding:"required"`
	PackageID  cdssdk.PackageID `json:"packageID" binding:"required"`
	Path       string           `json:"path" binding:"required"`
	UpdateTime time.Time        `json:"updateTime" binding:"required"`
}

type ObjectUpdateInfoByPathResp struct{}

func (c *ObjectService) UpdateInfoByPath(req ObjectUpdateInfoByPath) (*ObjectUpdateInfoByPathResp, error) {
	url, err := url.JoinPath(c.baseURL, ObjectUpdateInfoByPathPath)
	if err != nil {
		return nil, err
	}

	resp, err := http2.PostJSON(url, http2.RequestParam{
		Body: req,
	})
	if err != nil {
		return nil, err
	}

	jsonResp, err := ParseJSONResponse[response[ObjectUpdateInfoByPathResp]](resp)
	if err != nil {
		return nil, err
	}

	if jsonResp.Code == errorcode.OK {
		return &jsonResp.Data, nil
	}

	return nil, jsonResp.ToError()
}

const ObjectMovePath = "/object/move"

type MovingObject struct {
	ObjectID  cdssdk.ObjectID  `json:"objectID" binding:"required"`
	PackageID cdssdk.PackageID `json:"packageID" binding:"required"`
	Path      string           `json:"path" binding:"required"`
}

func (m *MovingObject) ApplyTo(obj *cdssdk.Object) {
	obj.PackageID = m.PackageID
	obj.Path = m.Path
}

type ObjectMove struct {
	UserID  cdssdk.UserID  `json:"userID" binding:"required"`
	Movings []MovingObject `json:"movings" binding:"required"`
}

type ObjectMoveResp struct {
	Successes []cdssdk.ObjectID `json:"successes"`
}

func (c *ObjectService) Move(req ObjectMove) (*ObjectMoveResp, error) {
	url, err := url.JoinPath(c.baseURL, ObjectMovePath)
	if err != nil {
		return nil, err
	}

	resp, err := http2.PostJSON(url, http2.RequestParam{
		Body: req,
	})
	if err != nil {
		return nil, err
	}

	jsonResp, err := ParseJSONResponse[response[ObjectMoveResp]](resp)
	if err != nil {
		return nil, err
	}

	if jsonResp.Code == errorcode.OK {
		return &jsonResp.Data, nil
	}

	return nil, jsonResp.ToError()
}

const ObjectDeletePath = "/object/delete"

type ObjectDelete struct {
	UserID    cdssdk.UserID     `json:"userID" binding:"required"`
	ObjectIDs []cdssdk.ObjectID `json:"objectIDs" binding:"required"`
}

type ObjectDeleteResp struct{}

func (c *ObjectService) Delete(req ObjectDelete) error {
	url, err := url.JoinPath(c.baseURL, ObjectDeletePath)
	if err != nil {
		return err
	}

	resp, err := http2.PostJSON(url, http2.RequestParam{
		Body: req,
	})
	if err != nil {
		return err
	}

	jsonResp, err := ParseJSONResponse[response[ObjectDeleteResp]](resp)
	if err != nil {
		return err
	}

	if jsonResp.Code == errorcode.OK {
		return nil
	}

	return jsonResp.ToError()
}

const ObjectDeleteByPathPath = "/object/deleteByPath"

type ObjectDeleteByPath struct {
	UserID    cdssdk.UserID    `json:"userID" binding:"required"`
	PackageID cdssdk.PackageID `json:"packageID" binding:"required"`
	Path      string           `json:"path" binding:"required"`
}
type ObjectDeleteByPathResp struct{}

func (c *ObjectService) DeleteByPath(req ObjectDeleteByPath) error {
	url, err := url.JoinPath(c.baseURL, ObjectDeleteByPathPath)
	if err != nil {
		return err
	}

	resp, err := http2.PostJSON(url, http2.RequestParam{
		Body: req,
	})
	if err != nil {
		return err
	}

	jsonResp, err := ParseJSONResponse[response[ObjectDeleteByPathResp]](resp)
	if err != nil {
		return err
	}

	if jsonResp.Code == errorcode.OK {
		return nil
	}

	return jsonResp.ToError()
}

const ObjectGetPackageObjectsPath = "/object/getPackageObjects"

type ObjectGetPackageObjects struct {
	UserID    cdssdk.UserID    `form:"userID" json:"userID" binding:"required"`
	PackageID cdssdk.PackageID `form:"packageID" json:"packageID" binding:"required"`
}
type ObjectGetPackageObjectsResp struct {
	Objects []cdssdk.Object `json:"objects"`
}

func (c *ObjectService) GetPackageObjects(req ObjectGetPackageObjects) (*ObjectGetPackageObjectsResp, error) {
	url, err := url.JoinPath(c.baseURL, ObjectGetPackageObjectsPath)
	if err != nil {
		return nil, err
	}

	resp, err := http2.GetForm(url, http2.RequestParam{
		Query: req,
	})
	if err != nil {
		return nil, err
	}

	jsonResp, err := ParseJSONResponse[response[ObjectGetPackageObjectsResp]](resp)
	if err != nil {
		return nil, err
	}

	if jsonResp.Code == errorcode.OK {
		return &jsonResp.Data, nil
	}

	return nil, jsonResp.ToError()
}
