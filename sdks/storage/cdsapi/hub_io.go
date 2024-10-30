package cdsapi

import (
	"fmt"
	"io"
	"net/url"
	"strings"

	"gitlink.org.cn/cloudream/common/consts/errorcode"
	"gitlink.org.cn/cloudream/common/pkgs/ioswitch/exec"
	"gitlink.org.cn/cloudream/common/utils/http2"
	"gitlink.org.cn/cloudream/common/utils/serder"
)

// TODO2 重新梳理代码

const GetStreamPath = "/hubIO/getStream"

type GetStreamReq struct {
	PlanID   exec.PlanID   `json:"planID"`
	VarID    exec.VarID    `json:"varID"`
	SignalID exec.VarID    `json:"signalID"`
	Signal   exec.VarValue `json:"signal"`
}

func (c *Client) GetStream(req GetStreamReq) (io.ReadCloser, error) {
	targetUrl, err := url.JoinPath(c.baseURL, GetStreamPath)
	if err != nil {
		return nil, err
	}

	body, err := serder.ObjectToJSONEx(req)
	if err != nil {
		return nil, fmt.Errorf("request to json: %w", err)
	}

	resp, err := http2.GetJSON(targetUrl, http2.RequestParam{
		Body: body,
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

	return resp.Body, nil
}

const SendStreamPath = "/hubIO/sendStream"

type SendStreamReq struct {
	SendStreamInfo
	Stream io.ReadCloser `json:"-"`
}
type SendStreamInfo struct {
	PlanID exec.PlanID `json:"planID"`
	VarID  exec.VarID  `json:"varID"`
}

func (c *Client) SendStream(req SendStreamReq) error {
	// targetUrl, err := url.JoinPath(c.baseURL, SendStreamPath)
	// if err != nil {
	// 	return err
	// }

	// infoJSON, err := serder.ObjectToJSON(req)
	// if err != nil {
	// 	return fmt.Errorf("info to json: %w", err)
	// }

	// resp, err := http2.PostMultiPart(targetUrl, http2.MultiPartRequestParam{
	// 	Form: map[string]string{"info": string(infoJSON)},
	// 	Files: iterator.Array(&http2.IterMultiPartFile{
	// 		FieldName: "stream",
	// 		FileName:  "stream",
	// 		File:      req.Stream,
	// 	}),
	// })
	// if err != nil {
	// 	return err
	// }

	// contType := resp.Header.Get("Content-Type")
	// if strings.Contains(contType, http2.ContentTypeJSON) {
	// 	var err error
	// 	var codeResp response[ObjectUploadResp]
	// 	if codeResp, err = serder.JSONToObjectStreamEx[response[ObjectUploadResp]](resp.Body); err != nil {
	// 		return fmt.Errorf("parsing response: %w", err)
	// 	}

	// 	if codeResp.Code == errorcode.OK {
	// 		return nil
	// 	}

	// 	return codeResp.ToError()
	// }

	// return fmt.Errorf("unknow response content type: %s", contType)
	return fmt.Errorf("not implemented")
}

const ExecuteIOPlanPath = "/hubIO/executeIOPlan"

type ExecuteIOPlanReq struct {
	Plan exec.Plan `json:"plan"`
}

func (c *Client) ExecuteIOPlan(req ExecuteIOPlanReq) error {
	targetUrl, err := url.JoinPath(c.baseURL, ExecuteIOPlanPath)
	if err != nil {
		return err
	}

	body, err := serder.ObjectToJSONEx(req)
	if err != nil {
		return fmt.Errorf("request to json: %w", err)
	}

	resp, err := http2.PostJSON(targetUrl, http2.RequestParam{
		Body: body,
	})
	if err != nil {
		return err
	}

	codeResp, err := ParseJSONResponse[response[any]](resp)
	if err != nil {
		return err
	}

	if codeResp.Code == errorcode.OK {
		return nil
	}

	return codeResp.ToError()
}

const SendVarPath = "/hubIO/sendVar"

type SendVarReq struct {
	PlanID   exec.PlanID   `json:"planID"`
	VarID    exec.VarID    `json:"varID"`
	VarValue exec.VarValue `json:"varValue"`
}

func (c *Client) SendVar(req SendVarReq) error {
	targetUrl, err := url.JoinPath(c.baseURL, SendVarPath)
	if err != nil {
		return err
	}

	body, err := serder.ObjectToJSONEx(req)
	if err != nil {
		return fmt.Errorf("request to json: %w", err)
	}

	resp, err := http2.PostJSON(targetUrl, http2.RequestParam{
		Body: body,
	})
	if err != nil {
		return err
	}

	jsonResp, err := ParseJSONResponse[response[any]](resp)
	if err != nil {
		return err
	}

	if jsonResp.Code == errorcode.OK {
		return nil
	}

	return jsonResp.ToError()
}

const GetVarPath = "/hubIO/getVar"

type GetVarReq struct {
	PlanID   exec.PlanID   `json:"planID"`
	VarID    exec.VarID    `json:"varID"`
	SignalID exec.VarID    `json:"signalID"`
	Signal   exec.VarValue `json:"signal"`
}

type GetVarResp struct {
	Value exec.VarValue `json:"value"`
}

func (c *Client) GetVar(req GetVarReq) (*GetVarResp, error) {
	targetUrl, err := url.JoinPath(c.baseURL, GetVarPath)
	if err != nil {
		return nil, err
	}

	body, err := serder.ObjectToJSONEx(req)
	if err != nil {
		return nil, fmt.Errorf("request to json: %w", err)
	}

	resp, err := http2.GetJSON(targetUrl, http2.RequestParam{
		Body: body,
	})
	if err != nil {
		return nil, err
	}

	jsonResp, err := ParseJSONResponse[response[GetVarResp]](resp)
	if err != nil {
		return nil, err
	}

	if jsonResp.Code == errorcode.OK {
		return &jsonResp.Data, nil
	}

	return nil, jsonResp.ToError()
}
