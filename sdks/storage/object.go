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

type ObjectUpload struct {
	ObjectUploadInfo
	Files UploadObjectIterator `json:"-"`
}

type ObjectUploadInfo struct {
	UserID       UserID    `json:"userID" binding:"required"`
	PackageID    PackageID `json:"packageID" binding:"required"`
	NodeAffinity *NodeID   `json:"nodeAffinity"`
}

type UploadingObject struct {
	Path string
	File io.ReadCloser
}

type UploadObjectIterator = iterator.Iterator[*UploadingObject]

type ObjectUploadResp struct {
	Uploadeds []UploadedObject `json:"uploadeds"`
}
type UploadedObject struct {
	Object *Object `json:"object"`
	Error  string  `json:"error"`
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

	resp, err := myhttp.PostMultiPart(url, myhttp.MultiPartRequestParam{
		Form: map[string]string{"info": string(infoJSON)},
		Files: iterator.Map(req.Files, func(src *UploadingObject) (*myhttp.IterMultiPartFile, error) {
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
	UserID   UserID   `form:"userID" json:"userID" binding:"required"`
	ObjectID ObjectID `form:"objectID" json:"objectID" binding:"required"`
	Offset   int64    `form:"offset" json:"offset,omitempty"`
	Length   *int64   `form:"length" json:"length,omitempty"`
	PartSize int64    `form:"partSize" json:"partSize,omitempty"`
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

	_, files, err := myhttp.ParseMultiPartResponse(resp)
	if err != nil {
		return nil, err
	}

	startTime := time.Now()
	file, err := files.MoveNext()
	endTime := time.Now()
	fmt.Printf("files.MoveNext(), spend time: %.0f s", endTime.Sub(startTime).Seconds())
	if err == iterator.ErrNoMoreItem {
		return nil, fmt.Errorf("no file found in response")
	}
	if err != nil {
		return nil, err
	}

	return &DownloadingObject{
		Path: file.FileName,
		File: file.File,
	}, nil
}

const ObjectUpdateInfoPath = "/object/updateInfo"

type UpdatingObject struct {
	ObjectID   ObjectID  `json:"objectID" binding:"required"`
	UpdateTime time.Time `json:"updateTime" binding:"required"`
}

func (u *UpdatingObject) ApplyTo(obj *Object) {
	obj.UpdateTime = u.UpdateTime
}

type ObjectUpdateInfo struct {
	UserID    UserID           `json:"userID" binding:"required"`
	Updatings []UpdatingObject `json:"updatings" binding:"required"`
}

type ObjectUpdateInfoResp struct {
	Successes []ObjectID `json:"successes"`
}

func (c *ObjectService) UpdateInfo(req ObjectUpdateInfo) (*ObjectUpdateInfoResp, error) {
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

	jsonResp, err := ParseJSONResponse[response[ObjectUpdateInfoResp]](resp)
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
	ObjectID  ObjectID  `json:"objectID" binding:"required"`
	PackageID PackageID `json:"packageID" binding:"required"`
	Path      string    `json:"path" binding:"required"`
}

func (m *MovingObject) ApplyTo(obj *Object) {
	obj.PackageID = m.PackageID
	obj.Path = m.Path
}

type ObjectMove struct {
	UserID  UserID         `json:"userID" binding:"required"`
	Movings []MovingObject `json:"movings" binding:"required"`
}

type ObjectMoveResp struct {
	Successes []ObjectID `json:"successes"`
}

func (c *ObjectService) Move(req ObjectMove) (*ObjectMoveResp, error) {
	url, err := url.JoinPath(c.baseURL, ObjectMovePath)
	if err != nil {
		return nil, err
	}

	resp, err := myhttp.PostJSON(url, myhttp.RequestParam{
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
	UserID    UserID     `json:"userID" binding:"required"`
	ObjectIDs []ObjectID `json:"objectIDs" binding:"required"`
}

type ObjectDeleteResp struct{}

func (c *ObjectService) Delete(req ObjectDelete) error {
	url, err := url.JoinPath(c.baseURL, ObjectDeletePath)
	if err != nil {
		return err
	}

	resp, err := myhttp.PostJSON(url, myhttp.RequestParam{
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

const ObjectGetPackageObjectsPath = "/object/getPackageObjects"

type ObjectGetPackageObjects struct {
	UserID    UserID    `form:"userID" json:"userID" binding:"required"`
	PackageID PackageID `form:"packageID" json:"packageID" binding:"required"`
}
type ObjectGetPackageObjectsResp struct {
	Objects []Object `json:"objects"`
}

func (c *ObjectService) GetPackageObjects(req ObjectGetPackageObjects) (*ObjectGetPackageObjectsResp, error) {
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

	jsonResp, err := ParseJSONResponse[response[ObjectGetPackageObjectsResp]](resp)
	if err != nil {
		return nil, err
	}

	if jsonResp.Code == errorcode.OK {
		return &jsonResp.Data, nil
	}

	return nil, jsonResp.ToError()
}
