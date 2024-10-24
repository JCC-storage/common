package pcmsdk

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"gitlink.org.cn/cloudream/common/sdks"
	schsdk "gitlink.org.cn/cloudream/common/sdks/scheduler"
	"gitlink.org.cn/cloudream/common/utils/http2"
	"gitlink.org.cn/cloudream/common/utils/serder"
)

const CodeOK int = 200

type UploadImageReq struct {
	PartID    ParticipantID `json:"partID"`
	ImagePath string        `json:"imagePath"`
}

type UploadImageResp struct {
	Result  string  `json:"result"`
	ImageID ImageID `json:"imageID"`
	Name    string  `json:"name"`
}

// TODO
func (c *Client) UploadImage(req UploadImageReq) (*UploadImageResp, error) {
	url, err := url.JoinPath(c.baseURL, "/pcm/v1/storelink/uploadImage")
	if err != nil {
		return nil, err
	}
	resp, err := http2.PostJSON(url, http2.RequestParam{
		Body: req,
	})
	if err != nil {
		return nil, err
	}

	contType := resp.Header.Get("Content-Type")
	if strings.Contains(contType, http2.ContentTypeJSON) {

		var codeResp response[UploadImageResp]
		if err := serder.JSONToObjectStream(resp.Body, &codeResp); err != nil {
			return nil, fmt.Errorf("parsing response: %w", err)
		}

		if codeResp.Code == CodeOK {
			return &codeResp.Data, nil
		}

		return nil, codeResp.ToError()
	}

	return nil, fmt.Errorf("unknow response content type: %s", contType)
}

type GetParticipantsResp struct {
	Participants []Participant
}

func (c *Client) GetParticipants() (*GetParticipantsResp, error) {
	type Resp struct {
		Code         int           `json:"code"`
		Message      string        `json:"message"`
		Participants []Participant `json:"participants"`
	}

	url, err := url.JoinPath(c.baseURL, "/pcm/v1/storelink/uploadImage")
	if err != nil {
		return nil, err
	}
	rawResp, err := http2.GetJSON(url, http2.RequestParam{})
	if err != nil {
		return nil, err
	}

	resp, err := http2.ParseJSONResponse[Resp](rawResp)
	if err != nil {
		return nil, err
	}

	if resp.Code != CodeOK {
		return nil, &sdks.CodeMessageError{
			Code:    fmt.Sprintf("%d", resp.Code),
			Message: resp.Message,
		}
	}

	return &GetParticipantsResp{
		Participants: resp.Participants,
	}, nil
}

type GetImageListReq struct {
	PartID ParticipantID `json:"partId"`
}

type GetImageListResp struct {
	Images []Image
}

func (c *Client) GetImageList(req GetImageListReq) (*GetImageListResp, error) {
	type Resp struct {
		Success  bool    `json:"success"`
		Images   []Image `json:"images"`
		ErrorMsg string  `json:"errorMsg"`
	}

	url, err := url.JoinPath(c.baseURL, "/pcm/v1/storelink/getImageList")
	if err != nil {
		return nil, err
	}
	rawResp, err := http2.GetJSON(url, http2.RequestParam{
		Body: req,
	})
	if err != nil {
		return nil, err
	}

	resp, err := http2.ParseJSONResponse[Resp](rawResp)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf(resp.ErrorMsg)
	}

	return &GetImageListResp{
		Images: resp.Images,
	}, nil
}

type DeleteImageReq struct {
	PartID  ParticipantID `json:"partID"`
	ImageID ImageID       `json:"imageID"`
}

func (c *Client) DeleteImage(req DeleteImageReq) error {
	type Resp struct {
		Success  bool   `json:"success"`
		ErrorMsg string `json:"errorMsg"`
	}

	url, err := url.JoinPath(c.baseURL, "/pcm/v1/storelink/deleteImage")
	if err != nil {
		return err
	}
	rawResp, err := http2.PostJSON(url, http2.RequestParam{
		Body: req,
	})
	if err != nil {
		return err
	}

	resp, err := http2.ParseJSONResponse[Resp](rawResp)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf(resp.ErrorMsg)
	}

	return nil
}

type SubmitTaskReq struct {
	PartID     ParticipantID   `json:"partId"`
	ImageID    ImageID         `json:"imageId"`
	ResourceID ResourceID      `json:"resourceId"`
	CMD        string          `json:"cmd"`
	Params     []schsdk.KVPair `json:"params"`
	Envs       []schsdk.KVPair `json:"envs"`
}

type SubmitTaskResp struct {
	TaskID TaskID
}

func (c *Client) SubmitTask(req SubmitTaskReq) (*SubmitTaskResp, error) {
	type Resp struct {
		Success  bool   `json:"success"`
		TaskID   TaskID `json:"taskId"`
		ErrorMsg string `json:"errorMsg"`
	}

	url, err := url.JoinPath(c.baseURL, "/pcm/v1/storelink/submitTask")
	if err != nil {
		return nil, err
	}
	rawResp, err := http2.PostJSON(url, http2.RequestParam{
		Body: req,
	})
	if err != nil {
		return nil, err
	}

	resp, err := http2.ParseJSONResponse[Resp](rawResp)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf(resp.ErrorMsg)
	}

	return &SubmitTaskResp{
		TaskID: resp.TaskID,
	}, nil
}

type GetTaskReq struct {
	PartID ParticipantID `json:"partId"`
	TaskID TaskID        `json:"taskId"`
}

type GetTaskResp struct {
	TaskStatus  TaskStatus
	TaskName    string
	StartedAt   time.Time
	CompletedAt time.Time
}

func (c *Client) GetTask(req GetTaskReq) (*GetTaskResp, error) {
	type Resp struct {
		Success bool `json:"success"`
		Task    struct {
			TaskID      TaskID                 `json:"taskId"`
			TaskStatus  TaskStatus             `json:"taskStatus"`
			TaskName    string                 `json:"taskName"`
			StartedAt   serder.TimestampSecond `json:"startedAt"`
			CompletedAt serder.TimestampSecond `json:"completedAt"`
		} `json:"task"`
		ErrorMsg string `json:"errorMsg"`
	}

	url, err := url.JoinPath(c.baseURL, "/pcm/v1/storelink/getTask")
	if err != nil {
		return nil, err
	}
	rawResp, err := http2.GetJSON(url, http2.RequestParam{
		Body: req,
	})
	if err != nil {
		return nil, err
	}

	resp, err := http2.ParseJSONResponse[Resp](rawResp)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf(resp.ErrorMsg)
	}

	return &GetTaskResp{
		TaskStatus:  resp.Task.TaskStatus,
		TaskName:    resp.Task.TaskName,
		StartedAt:   time.Time(resp.Task.StartedAt),
		CompletedAt: time.Time(resp.Task.CompletedAt),
	}, nil
}

type DeleteTaskReq struct {
	PartID ParticipantID `json:"partId"`
	TaskID TaskID        `json:"taskId"`
}

func (c *Client) DeleteTask(req DeleteTaskReq) error {
	type Resp struct {
		Success  bool   `json:"success"`
		ErrorMsg string `json:"errorMsg"`
	}

	url, err := url.JoinPath(c.baseURL, "/pcm/v1/storelink/deleteTask")
	if err != nil {
		return err
	}
	rawResp, err := http2.PostJSON(url, http2.RequestParam{
		Body: req,
	})
	if err != nil {
		return err
	}

	resp, err := http2.ParseJSONResponse[Resp](rawResp)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf(resp.ErrorMsg)
	}

	return nil
}

type GetResourceSpecs struct {
	PartID ParticipantID `json:"partId"`
}

type GetResourceSpecsResp struct {
	Resources []Resource
}

func (c *Client) GetResourceSpecs(req GetImageListReq) (*GetResourceSpecsResp, error) {
	type Resp struct {
		Success       bool       `json:"success"`
		ResourceSpecs []Resource `json:"resourceSpecs"`
		ErrorMsg      string     `json:"errorMsg"`
	}

	url, err := url.JoinPath(c.baseURL, "/pcm/v1/storelink/getResourceSpecs")
	if err != nil {
		return nil, err
	}
	rawResp, err := http2.GetJSON(url, http2.RequestParam{
		Body: req,
	})
	if err != nil {
		return nil, err
	}

	resp, err := http2.ParseJSONResponse[Resp](rawResp)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf(resp.ErrorMsg)
	}

	return &GetResourceSpecsResp{
		Resources: resp.ResourceSpecs,
	}, nil
}
