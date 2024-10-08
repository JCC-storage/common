package cdssdk

import (
	"net/url"

	"gitlink.org.cn/cloudream/common/consts/errorcode"
	"gitlink.org.cn/cloudream/common/utils/http2"
)

type BucketService struct {
	*Client
}

func (c *Client) Bucket() *BucketService {
	return &BucketService{c}
}

const BucketGetByNamePath = "/bucket/getByName"

type BucketGetByName struct {
	UserID UserID `json:"userID" form:"userID" binding:"required"`
	Name   string `json:"name" form:"name" binding:"required"`
}
type BucketGetByNameResp struct {
	Bucket Bucket `json:"bucket"`
}

func (c *BucketService) GetByName(req BucketGetByName) (*BucketGetByNameResp, error) {
	url, err := url.JoinPath(c.baseURL, BucketGetByNamePath)
	if err != nil {
		return nil, err
	}

	resp, err := http2.GetForm(url, http2.RequestParam{
		Query: req,
	})
	if err != nil {
		return nil, err
	}

	codeResp, err := ParseJSONResponse[response[BucketGetByNameResp]](resp)
	if err != nil {
		return nil, err
	}

	if codeResp.Code == errorcode.OK {
		return &codeResp.Data, nil
	}

	return nil, codeResp.ToError()
}

const BucketCreatePath = "/bucket/create"

type BucketCreate struct {
	UserID UserID `json:"userID" binding:"required"`
	Name   string `json:"name" binding:"required"`
}

type BucketCreateResp struct {
	Bucket Bucket `json:"bucket"`
}

func (c *BucketService) Create(req BucketCreate) (*BucketCreateResp, error) {
	url, err := url.JoinPath(c.baseURL, BucketCreatePath)
	if err != nil {
		return nil, err
	}

	resp, err := http2.PostJSON(url, http2.RequestParam{
		Body: req,
	})
	if err != nil {
		return nil, err
	}

	codeResp, err := ParseJSONResponse[response[BucketCreateResp]](resp)
	if err != nil {
		return nil, err
	}

	if codeResp.Code == errorcode.OK {
		return &codeResp.Data, nil
	}

	return nil, codeResp.ToError()
}

const BucketDeletePath = "/bucket/delete"

type BucketDelete struct {
	UserID   UserID   `json:"userID" binding:"required"`
	BucketID BucketID `json:"bucketID" binding:"required"`
}

type BucketDeleteResp struct{}

func (c *BucketService) Delete(req BucketDelete) error {
	url, err := url.JoinPath(c.baseURL, BucketDeletePath)
	if err != nil {
		return err
	}

	resp, err := http2.PostJSON(url, http2.RequestParam{
		Body: req,
	})
	if err != nil {
		return err
	}

	codeResp, err := ParseJSONResponse[response[BucketDeleteResp]](resp)
	if err != nil {
		return err
	}

	if codeResp.Code == errorcode.OK {
		return nil
	}

	return codeResp.ToError()
}

const BucketListUserBucketsPath = "/bucket/listUserBuckets"

type BucketListUserBucketsReq struct {
	UserID UserID `form:"userID" json:"userID" binding:"required"`
}

type BucketListUserBucketsResp struct {
	Buckets []Bucket `json:"buckets"`
}

func (c *BucketService) ListUserBuckets(req BucketListUserBucketsReq) (*BucketListUserBucketsResp, error) {
	url, err := url.JoinPath(c.baseURL, BucketListUserBucketsPath)
	if err != nil {
		return nil, err
	}

	resp, err := http2.GetForm(url, http2.RequestParam{
		Query: req,
	})
	if err != nil {
		return nil, err
	}

	codeResp, err := ParseJSONResponse[response[BucketListUserBucketsResp]](resp)
	if err != nil {
		return nil, err
	}

	if codeResp.Code == errorcode.OK {
		return &codeResp.Data, nil
	}

	return nil, codeResp.ToError()
}
