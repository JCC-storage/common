package cdssdk

import (
	"net/url"

	"gitlink.org.cn/cloudream/common/consts/errorcode"
	myhttp "gitlink.org.cn/cloudream/common/utils/http"
)

type BucketService struct {
	*Client
}

func (c *Client) Bucket() *BucketService {
	return &BucketService{c}
}

const BucketCreatePath = "/bucket/create"

type BucketCreateReq struct {
	UserID     UserID `json:"userID" binding:"required"`
	BucketName string `json:"bucketName" binding:"required"`
}

type BucketCreateResp struct {
	BucketID BucketID `json:"bucketID"`
}

func (c *BucketService) Create(req BucketCreateReq) (*BucketCreateResp, error) {
	url, err := url.JoinPath(c.baseURL, BucketCreatePath)
	if err != nil {
		return nil, err
	}

	resp, err := myhttp.PostJSON(url, myhttp.RequestParam{
		Body: req,
	})
	if err != nil {
		return nil, err
	}

	codeResp, err := myhttp.ParseJSONResponse[response[BucketCreateResp]](resp)
	if err != nil {
		return nil, err
	}

	if codeResp.Code == errorcode.OK {
		return &codeResp.Data, nil
	}

	return nil, codeResp.ToError()
}

const BucketDeletePath = "/bucket/delete"

type BucketDeleteReq struct {
	UserID   UserID   `json:"userID" binding:"required"`
	BucketID BucketID `json:"bucketID" binding:"required"`
}

type BucketDeleteResp struct{}

func (c *BucketService) Delete(req BucketDeleteReq) (*BucketDeleteResp, error) {
	url, err := url.JoinPath(c.baseURL, BucketDeletePath)
	if err != nil {
		return nil, err
	}

	resp, err := myhttp.PostJSON(url, myhttp.RequestParam{
		Body: req,
	})
	if err != nil {
		return nil, err
	}

	codeResp, err := myhttp.ParseJSONResponse[response[BucketDeleteResp]](resp)
	if err != nil {
		return nil, err
	}

	if codeResp.Code == errorcode.OK {
		return &codeResp.Data, nil
	}

	return nil, codeResp.ToError()
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

	resp, err := myhttp.GetForm(url, myhttp.RequestParam{
		Query: req,
	})
	if err != nil {
		return nil, err
	}

	codeResp, err := myhttp.ParseJSONResponse[response[BucketListUserBucketsResp]](resp)
	if err != nil {
		return nil, err
	}

	if codeResp.Code == errorcode.OK {
		return &codeResp.Data, nil
	}

	return nil, codeResp.ToError()
}
