package schsdk

import (
	"fmt"
	"gitlink.org.cn/cloudream/common/consts/errorcode"
	"gitlink.org.cn/cloudream/common/pkgs/mq"
	myhttp "gitlink.org.cn/cloudream/common/utils/http"
	"gitlink.org.cn/cloudream/common/utils/serder"
	"net/url"
	"strings"
)

// 这个结构体无任何字段，但实现了Noop，每种MessageBody都要内嵌这个结构体
type MessageBodyBase struct{}

// 此处的receiver是指针
func (b *MessageBodyBase) Noop() {}

type RunningModelResp struct {
	MessageBodyBase
	RunningModels map[string]RunningModelInfo `json:"allNode"`
}

type AllModelResp struct {
	MessageBodyBase
	AllModels []Models `json:"allModels"`
}

type Models struct {
	ModelID   ModelID   `json:"modelID"`
	ModelName ModelName `json:"modelName"`
}

type NodeInfo struct {
	MessageBodyBase
	InstanceID JobID `json:"instanceID"`
	//NodeID     NodeID  `json:"nodeID"`
	Address Address `json:"address"`
	Status  string  `json:"status"`
}

type RunningModelInfo struct {
	MessageBodyBase
	JobSetID        JobSetID   `json:"jobSetID"`
	ModelID         ModelID    `json:"modelID"`
	ModelName       ModelName  `json:"modelName"`
	CustomModelName ModelName  `json:"customModelName"`
	Nodes           []NodeInfo `json:"nodes"`
}

type ECSNodeRunningInfoReq struct {
	mq.MessageBodyBase
	CustomModelName ModelName `form:"customModelName" json:"customModelName" binding:"required"`
	ModelID         ModelID   `form:"modelID" json:"modelID" binding:"required"`
}

type ECSNodeRunningInfoResp struct {
	MessageBodyBase
	NodeUsageRateInfos []NodeUsageRateInfo `json:"nodeUsageRateInfos"`
}

func NewECSNodeRunningInfoResp(nodeUsageRateInfos []NodeUsageRateInfo) *ECSNodeRunningInfoResp {
	return &ECSNodeRunningInfoResp{
		NodeUsageRateInfos: nodeUsageRateInfos,
	}
}

type NodeUsageRateInfo struct {
	MessageBodyBase
	InstanceID        JobID       `json:"instanceID"`
	Address           Address     `json:"address"`
	MemoryUtilization []UsageRate `json:"memoryUtilization"`
	GPUUtilization    []UsageRate `json:"GPUUtilization"`
	CPUUtilization    []UsageRate `json:"CPUUtilization"`
}

type UsageRate struct {
	Timestamp string `json:"timestamp"`
	Number    string `json:"number"`
}

const (
	FineTuning = "finetuning"

	CreateECS     = "create"
	RunECS        = "run"
	PauseECS      = "pause"
	DestroyECS    = "destroy"
	OperateServer = "operate"

	GPUMonitor = "GPUMonitor"

	RcloneMount = "rclone"
	Mounted     = "mounted"

	Deploying = "Deploying"
	Waiting   = "Waiting"
	Failed    = "Failed"
)

type QueryRunningModelsReq struct {
	UserID int64 `form:"userID" json:"userID"`
}

func (c *Client) QueryRunningModels(req QueryRunningModelsReq) (*RunningModelResp, error) {
	url, err := url.JoinPath(c.baseURL, "/job/queryRunningModels")
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
		var codeResp response[RunningModelResp]
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

func (c *Client) QueryAllModels(req QueryRunningModelsReq) (*AllModelResp, error) {
	url, err := url.JoinPath(c.baseURL, "/job/getAllModels")
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
		var codeResp response[AllModelResp]
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

func (c *Client) ECSNodeRunningInfo(req ECSNodeRunningInfoReq) (*ECSNodeRunningInfoResp, error) {
	url, err := url.JoinPath(c.baseURL, "/job/getECSNodeRunningInfo")
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
		var codeResp response[ECSNodeRunningInfoResp]
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
