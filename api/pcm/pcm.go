package pcm

import (
	"fmt"
	"net/url"
	"strings"

	myhttp "gitlink.org.cn/cloudream/common/utils/http"
	"gitlink.org.cn/cloudream/common/utils/serder"
)

const CORRECT_CODE int = 200

type UploadImageReq struct {
	SlwNodeID int64  `json:"slwNodeID"`
	ImagePath string `json:"imagePath"`
}

type UploadImageResp struct {
	Result  string `json:"result"`
	ImageID int64  `json:"imageID"`
}

func (c *Client) UploadImage(req UploadImageReq) (*UploadImageResp, error) {
	url, err := url.JoinPath(c.baseURL, "/pcm/v1/core/uploadImage")
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

		var codeResp response[UploadImageResp]
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

type GetImageListReq struct {
	SlwNodeID int64 `json:"slwNodeID"`
}

type GetImageListResp struct {
	ImageIDs []int64 `json:"imageIDs"`
}

func (c *Client) GetImageList(req GetImageListReq) (*GetImageListResp, error) {
	url, err := url.JoinPath(c.baseURL, "/pcm/v1/core/getImageList")
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

		var codeResp response[GetImageListResp]
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

type DeleteImageReq struct {
	SlwNodeID int64 `json:"slwNodeID"`
	PCMJobID  int64 `json:"pcmJobID"`
}

type DeleteImageResp struct {
	Result string `json:"result"`
}

func (c *Client) DeleteImage(req DeleteImageReq) (*DeleteImageResp, error) {
	url, err := url.JoinPath(c.baseURL, "/pcm/v1/core/deleteImage")
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

		var codeResp response[DeleteImageResp]
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

type ScheduleTaskReq struct {
	SlwNodeID int64               `json:"slwNodeID"`
	Envs      []map[string]string `json:"envs"`
	ImageID   int64               `json:"imageID"`
	CMDLine   string              `json:"cmdLine"`
}

type ScheduleTaskResp struct {
	Result   string `json:"result"`
	PCMJobID int64  `json:"pcmJobID"`
}

func (c *Client) ScheduleTask(req ScheduleTaskReq) (*ScheduleTaskResp, error) {
	url, err := url.JoinPath(c.baseURL, "/pcm/v1/core/scheduleTask")
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

		var codeResp response[ScheduleTaskResp]
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
	SlwNodeID int64 `json:"slwNodeID"`
	PCMJobID  int64 `json:"pcmJobID"`
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
	SlwNodeID int64 `json:"slwNodeID"`
	PCMJobID  int64 `json:"pcmJobID"`
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
