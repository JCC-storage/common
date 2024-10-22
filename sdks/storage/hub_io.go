package cdssdk

import (
	"bytes"
	"fmt"
	"gitlink.org.cn/cloudream/common/consts/errorcode"
	"gitlink.org.cn/cloudream/common/pkgs/ioswitch/exec"
	"gitlink.org.cn/cloudream/common/utils/http2"
	"gitlink.org.cn/cloudream/common/utils/serder"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
)

const GetStreamPath = "/hubIO/getStream"

type GetStreamReq struct {
	PlanID exec.PlanID     `json:"planID"`
	VarID  exec.VarID      `json:"varID"`
	Signal *exec.SignalVar `json:"signal"`
}

func (c *Client) GetStream(planID exec.PlanID, id exec.VarID, signal *exec.SignalVar) (io.ReadCloser, error) {
	targetUrl, err := url.JoinPath(c.baseURL, GetStreamPath)
	if err != nil {
		return nil, err
	}

	req := &GetStreamReq{
		PlanID: planID,
		VarID:  id,
		Signal: signal,
	}

	resp, err := http2.GetJSON(targetUrl, http2.RequestParam{
		Body: req,
	})
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		// 读取错误信息
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("error response from server: %s", string(body))
	}

	return resp.Body, nil
}

const SendStreamPath = "/hubIO/sendStream"

type SendStreamReq struct {
	PlanID exec.PlanID   `json:"planID"`
	VarID  exec.VarID    `json:"varID"`
	Stream io.ReadCloser `json:"stream"`
}

func (c *Client) SendStream(planID exec.PlanID, varID exec.VarID, str io.Reader) error {
	targetUrl, err := url.JoinPath(c.baseURL, SendStreamPath)
	if err != nil {
		return err
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	_ = writer.WriteField("plan_id", string(planID))
	_ = writer.WriteField("var_id", string(rune(varID)))
	fileWriter, err := writer.CreateFormFile("file", "data")
	if err != nil {
		return fmt.Errorf("creating form file: %w", err)
	}

	// 将读取的数据写入 multipart 的文件部分
	_, err = io.Copy(fileWriter, str)
	if err != nil {
		return fmt.Errorf("copying stream data: %w", err)
	}

	// 关闭 writer 以结束 multipart
	err = writer.Close()
	if err != nil {
		return fmt.Errorf("closing writer: %w", err)
	}

	// 创建 HTTP 请求
	req, err := http.NewRequest("POST", targetUrl, body)
	if err != nil {
		return fmt.Errorf("creating HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// 发送请求
	cli := http.Client{}
	resp, err := cli.Do(req)
	if err != nil {
		return fmt.Errorf("sending HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned non-200 status: %d", resp.StatusCode)
	}

	return nil
}

const ExecuteIOPlanPath = "/hubIO/executeIOPlan"

type ExecuteIOPlanReq struct {
	Plan exec.Plan `json:"plan"`
}

func (c *Client) ExecuteIOPlan(plan exec.Plan) error {
	targetUrl, err := url.JoinPath(c.baseURL, ExecuteIOPlanPath)
	if err != nil {
		return err
	}

	req := &ExecuteIOPlanReq{
		Plan: plan,
	}

	resp, err := http2.PostJSON(targetUrl, http2.RequestParam{
		Body: req,
	})
	if err != nil {
		return err
	}

	contType := resp.Header.Get("Content-Type")
	if strings.Contains(contType, http2.ContentTypeJSON) {
		var codeResp response[any]
		if err := serder.JSONToObjectStream(resp.Body, &codeResp); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		if codeResp.Code == errorcode.OK {
			return nil
		}

		return codeResp.ToError()
	}

	return fmt.Errorf("unknow response content type: %s", contType)
}

const SendVarPath = "/hubIO/sendVar"

type SendVarReq struct {
	PlanID exec.PlanID `json:"planID"`
	Var    exec.Var    `json:"var"`
}

func (c *Client) SendVar(id exec.PlanID, v exec.Var) error {
	targetUrl, err := url.JoinPath(c.baseURL, SendVarPath)
	if err != nil {
		return err
	}

	req := &SendVarReq{
		PlanID: id,
		Var:    v,
	}

	resp, err := http2.PostJSON(targetUrl, http2.RequestParam{
		Body: req,
	})
	if err != nil {
		return err
	}

	contType := resp.Header.Get("Content-Type")
	if strings.Contains(contType, http2.ContentTypeJSON) {
		var codeResp response[any]
		if err := serder.JSONToObjectStream(resp.Body, &codeResp); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		if codeResp.Code == errorcode.OK {
			return nil
		}

		return codeResp.ToError()
	}

	return fmt.Errorf("unknow response content type: %s", contType)
}

const GetVarPath = "/hubIO/getVar"

type GetVarReq struct {
	PlanID exec.PlanID     `json:"planID"`
	Var    exec.Var        `json:"var"`
	Signal *exec.SignalVar `json:"signal"`
}

func (c *Client) GetVar(id exec.PlanID, v exec.Var, signal *exec.SignalVar) error {
	targetUrl, err := url.JoinPath(c.baseURL, GetVarPath)
	if err != nil {
		return err
	}

	req := &GetVarReq{
		PlanID: id,
		Var:    v,
		Signal: signal,
	}

	resp, err := http2.GetJSON(targetUrl, http2.RequestParam{
		Body: req,
	})
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		// 读取错误信息
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return fmt.Errorf("error response from server: %s", string(body))
	}

	return nil
}
