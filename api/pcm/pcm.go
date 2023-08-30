package pcm

import (
	"fmt"
	"net/url"
	"strings"

	myhttp "gitlink.org.cn/cloudream/common/utils/http"
	"gitlink.org.cn/cloudream/common/utils/serder"
)

const CORRECT_CODE int = 200

type UploadImgReq struct {
	NodeID  int64  `json:"nodeID"`
	ImgPath string `json:"imgPath"`
}

type UploadImgResp struct {
	Result string `json:"result"`
	ImgID  int64  `json:"imgID"`
}

func (c *Client) UploadImg(req UploadImgReq) (*UploadImgResp, error) {
	url, err := url.JoinPath(c.baseURL, "/pcm/v1/core/uploadImg")
	if err != nil {
		return nil, err
	}
	resp, err := myhttp.PostJSON(url, myhttp.RequestParam{
		Body: req,
	})
	if err != nil {
		return nil, err
	}

	contType := resp.Header.Get("Content-Type")
	if strings.Contains(contType, myhttp.ContentTypeJSON) {

		var codeResp response[UploadImgResp]
		if err := serder.JSONToObjectStream(resp.Body, &codeResp); err != nil {
			return nil, fmt.Errorf("parsing response: %w", err)
		}

		if codeResp.Code == CORRECT_CODE {
			return &codeResp.Data, nil
		}

		return nil, codeResp.ToError()
	}

	return nil, fmt.Errorf("unknow response content type: %s", contType)
}

type GetImgListReq struct {
	NodeID int64 `json:"nodeID"`
}

type GetImgListResp struct {
	ImgIDs []int64 `json:"imgIDs"`
}

func (c *Client) GetImgList(req GetImgListReq) (*GetImgListResp, error) {
	url, err := url.JoinPath(c.baseURL, "/pcm/v1/core/getImgList")
	if err != nil {
		return nil, err
	}
	resp, err := myhttp.PostJSON(url, myhttp.RequestParam{
		Body: req,
	})
	if err != nil {
		return nil, err
	}

	contType := resp.Header.Get("Content-Type")
	if strings.Contains(contType, myhttp.ContentTypeJSON) {

		var codeResp response[GetImgListResp]
		if err := serder.JSONToObjectStream(resp.Body, &codeResp); err != nil {
			return nil, fmt.Errorf("parsing response: %w", err)
		}

		if codeResp.Code == CORRECT_CODE {
			return &codeResp.Data, nil
		}

		return nil, codeResp.ToError()
	}

	return nil, fmt.Errorf("unknow response content type: %s", contType)
}

type DeleteImgReq struct {
	NodeID   int64 `json:"nodeID"`
	PCMJobID int64 `json:"pcmJobID"`
}

type DeleteImgResp struct {
	Result string `json:"result"`
}

func (c *Client) DeleteImg(req DeleteImgReq) (*DeleteImgResp, error) {
	url, err := url.JoinPath(c.baseURL, "/pcm/v1/core/deleteImg")
	if err != nil {
		return nil, err
	}
	resp, err := myhttp.PostJSON(url, myhttp.RequestParam{
		Body: req,
	})
	if err != nil {
		return nil, err
	}

	contType := resp.Header.Get("Content-Type")
	if strings.Contains(contType, myhttp.ContentTypeJSON) {

		var codeResp response[DeleteImgResp]
		if err := serder.JSONToObjectStream(resp.Body, &codeResp); err != nil {
			return nil, fmt.Errorf("parsing response: %w", err)
		}

		if codeResp.Code == CORRECT_CODE {
			return &codeResp.Data, nil
		}

		return nil, codeResp.ToError()
	}

	return nil, fmt.Errorf("unknow response content type: %s", contType)
}

type SchedulerTaskReq struct {
	NodeID  int64               `json:"nodeID"`
	Envs    []map[string]string `json:"envs"`
	ImgID   int64               `json:"imgID"`
	CMDLine string              `json:"cmdLine"`
}

type SchedulerTaskResp struct {
	Result   string `json:"result"`
	PCMJobID int64  `json:"pcmJobID"`
}

func (c *Client) SchedulerTask(req SchedulerTaskReq) (*SchedulerTaskResp, error) {
	url, err := url.JoinPath(c.baseURL, "/pcm/v1/core/schedulerTask")
	if err != nil {
		return nil, err
	}
	resp, err := myhttp.PostJSON(url, myhttp.RequestParam{
		Body: req,
	})
	if err != nil {
		return nil, err
	}

	contType := resp.Header.Get("Content-Type")
	if strings.Contains(contType, myhttp.ContentTypeJSON) {

		var codeResp response[SchedulerTaskResp]
		if err := serder.JSONToObjectStream(resp.Body, &codeResp); err != nil {
			return nil, fmt.Errorf("parsing response: %w", err)
		}

		if codeResp.Code == CORRECT_CODE {
			return &codeResp.Data, nil
		}

		return nil, codeResp.ToError()
	}

	return nil, fmt.Errorf("unknow response content type: %s", contType)
}

type GetTaskStatusReq struct {
	NodeID   int64 `json:"nodeID"`
	PCMJobID int64 `json:"pcmJobID"`
}

type GetTaskStatusResp struct {
	Result string `json:"result"`
	Status string `json:"status"`
}

func (c *Client) GetTaskStatus(req GetTaskStatusReq) (*GetTaskStatusResp, error) {
	url, err := url.JoinPath(c.baseURL, "/pcm/v1/core/getTaskStatus")
	if err != nil {
		return nil, err
	}
	resp, err := myhttp.PostJSON(url, myhttp.RequestParam{
		Body: req,
	})
	if err != nil {
		return nil, err
	}

	contType := resp.Header.Get("Content-Type")
	if strings.Contains(contType, myhttp.ContentTypeJSON) {

		var codeResp response[GetTaskStatusResp]
		if err := serder.JSONToObjectStream(resp.Body, &codeResp); err != nil {
			return nil, fmt.Errorf("parsing response: %w", err)
		}

		if codeResp.Code == CORRECT_CODE {
			return &codeResp.Data, nil
		}

		return nil, codeResp.ToError()
	}

	return nil, fmt.Errorf("unknow response content type: %s", contType)
}

type DeleteTaskReq struct {
	NodeID   int64 `json:"nodeID"`
	PCMJobID int64 `json:"pcmJobID"`
}

type DeleteTaskResp struct {
	Result string `json:"result"`
}

func (c *Client) DeleteTask(req DeleteTaskReq) (*DeleteTaskResp, error) {
	url, err := url.JoinPath(c.baseURL, "/pcm/v1/core/deleteTask")
	if err != nil {
		return nil, err
	}
	resp, err := myhttp.PostJSON(url, myhttp.RequestParam{
		Body: req,
	})
	if err != nil {
		return nil, err
	}

	contType := resp.Header.Get("Content-Type")
	if strings.Contains(contType, myhttp.ContentTypeJSON) {

		var codeResp response[DeleteTaskResp]
		if err := serder.JSONToObjectStream(resp.Body, &codeResp); err != nil {
			return nil, fmt.Errorf("parsing response: %w", err)
		}

		if codeResp.Code == CORRECT_CODE {
			return &codeResp.Data, nil
		}

		return nil, codeResp.ToError()
	}

	return nil, fmt.Errorf("unknow response content type: %s", contType)
}
