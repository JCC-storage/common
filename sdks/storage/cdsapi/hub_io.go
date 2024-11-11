package cdsapi

import (
	"fmt"
	"io"
	"net/url"

	"gitlink.org.cn/cloudream/common/consts/errorcode"
	"gitlink.org.cn/cloudream/common/pkgs/ioswitch/exec"
	"gitlink.org.cn/cloudream/common/utils/http2"
	"gitlink.org.cn/cloudream/common/utils/io2"
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

	cr := http2.NewChunkedReader(resp.Body)
	_, str, err := cr.NextPart()
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	return io2.DelegateReadCloser(str, func() error {
		cr.Close()
		return nil
	}), nil
}

const SendStreamPath = "/hubIO/sendStream"

type SendStreamReq struct {
	SendStreamInfo
	Stream io.ReadCloser
}
type SendStreamInfo struct {
	PlanID exec.PlanID `json:"planID"`
	VarID  exec.VarID  `json:"varID"`
}

func (c *Client) SendStream(req SendStreamReq) error {
	targetUrl, err := url.JoinPath(c.baseURL, SendStreamPath)
	if err != nil {
		return err
	}

	pr, pw := io.Pipe()
	errCh := make(chan error, 1)
	go func() {
		cw := http2.NewChunkedWriter(pw)

		infoJSON, err := serder.ObjectToJSONEx(req)
		if err != nil {
			cw.Abort(fmt.Sprintf("info to json: %v", err))
			errCh <- fmt.Errorf("info to json: %w", err)
			return
		}

		if err := cw.WriteDataPart("info", infoJSON); err != nil {
			cw.Close()
			errCh <- fmt.Errorf("write info: %w", err)
			return
		}

		_, err = cw.WriteStreamPart("stream", req.Stream)
		if err != nil {
			cw.Close()
			errCh <- fmt.Errorf("write stream: %w", err)
			return
		}

		err = cw.Finish()
		if err != nil {
			errCh <- fmt.Errorf("finish chunked writer: %w", err)
			return
		}
	}()

	resp, err := http2.PostChunked2(targetUrl, http2.Chunked2RequestParam{
		Body: pr,
	})
	if err != nil {
		return err
	}

	err = <-errCh
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
