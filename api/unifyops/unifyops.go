package unifyops

import (
	"fmt"
	"net/url"
	"strings"

	myhttp "gitlink.org.cn/cloudream/common/utils/http"
	"gitlink.org.cn/cloudream/common/utils/serder"
)

const CORRECT_CODE int = 200

type SlwNode struct {
	ID          int64  `json:"ID"`
	Name        string `json:"name"`
	SlwRegionID int64  `json:"slwRegionID"`
}

func (c *Client) GetSlwNodeInfo() (*[]SlwNode, error) {
	url, err := url.JoinPath(c.baseURL, "/cmdb/resApi/getSlwNodeInfo")
	if err != nil {
		return nil, err
	}
	resp, err := myhttp.GetJSON(url, myhttp.RequestParam{})
	if err != nil {
		return nil, err
	}
	contType := resp.Header.Get("Content-Type")
	if strings.Contains(contType, myhttp.ContentTypeJSON) {

		var codeResp response[[]SlwNode]
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

type Node struct {
	NodeId int64 `json:"nodeId"`
}

type ResourceData[T any] struct {
	Name      string        `json:"name"`
	Total     DetailType[T] `json:"total"`
	Available DetailType[T] `json:"available"`
}

type DetailType[T any] struct {
	Unit  string `json:"unit"`
	Value T      `json:"value"`
}

func (c *Client) GetCPUData(node Node) (*ResourceData[int64], error) {
	//TODO 整合成一个接口，参数增加一个resourceType，根据type调用不同的接口
	url, err := url.JoinPath(c.baseURL, "/cmdb/resApi/getCPUData")
	if err != nil {
		return nil, err
	}
	resp, err := myhttp.PostJSON(url, myhttp.RequestParam{
		Body: node,
	})
	if err != nil {
		return nil, err
	}

	contType := resp.Header.Get("Content-Type")
	if strings.Contains(contType, myhttp.ContentTypeJSON) {

		var codeResp response[ResourceData[int64]]
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

func (c *Client) GetNPUData(node Node) (*ResourceData[int64], error) {
	url, err := url.JoinPath(c.baseURL, "/cmdb/resApi/getNPUData")
	if err != nil {
		return nil, err
	}
	resp, err := myhttp.PostJSON(url, myhttp.RequestParam{
		Body: node,
	})
	if err != nil {
		return nil, err
	}

	contType := resp.Header.Get("Content-Type")
	if strings.Contains(contType, myhttp.ContentTypeJSON) {

		var codeResp response[ResourceData[int64]]
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

func (c *Client) GetGPUData(node Node) (*ResourceData[int64], error) {
	url, err := url.JoinPath(c.baseURL, "/cmdb/resApi/getGPUData")
	if err != nil {
		return nil, err
	}
	resp, err := myhttp.PostJSON(url, myhttp.RequestParam{
		Body: node,
	})
	if err != nil {
		return nil, err
	}

	contType := resp.Header.Get("Content-Type")
	if strings.Contains(contType, myhttp.ContentTypeJSON) {

		var codeResp response[ResourceData[int64]]
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

func (c *Client) GetMLUData(node Node) (*ResourceData[int64], error) {
	url, err := url.JoinPath(c.baseURL, "/cmdb/resApi/getMLUData")
	if err != nil {
		return nil, err
	}
	resp, err := myhttp.PostJSON(url, myhttp.RequestParam{
		Body: node,
	})
	if err != nil {
		return nil, err
	}

	contType := resp.Header.Get("Content-Type")
	if strings.Contains(contType, myhttp.ContentTypeJSON) {

		var codeResp response[ResourceData[int64]]
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

func (c *Client) GetStorageData(node Node) (*ResourceData[float64], error) {
	url, err := url.JoinPath(c.baseURL, "/cmdb/resApi/getStorageData")
	if err != nil {
		return nil, err
	}
	resp, err := myhttp.PostJSON(url, myhttp.RequestParam{
		Body: node,
	})
	if err != nil {
		return nil, err
	}

	contType := resp.Header.Get("Content-Type")
	if strings.Contains(contType, myhttp.ContentTypeJSON) {

		var codeResp response[ResourceData[float64]]
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

func (c *Client) GetMemoryData(node Node) (*ResourceData[float64], error) {
	url, err := url.JoinPath(c.baseURL, "/cmdb/resApi/getMemoryData")
	if err != nil {
		return nil, err
	}
	resp, err := myhttp.PostJSON(url, myhttp.RequestParam{
		Body: node,
	})
	if err != nil {
		return nil, err
	}

	contType := resp.Header.Get("Content-Type")
	if strings.Contains(contType, myhttp.ContentTypeJSON) {

		var codeResp response[ResourceData[float64]]
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

func (c *Client) GetIndicatorData(node Node) (*[]ResourceData[any], error) {
	url, err := url.JoinPath(c.baseURL, "/cmdb/resApi/getIndicatorData")
	if err != nil {
		return nil, err
	}
	resp, err := myhttp.PostJSON(url, myhttp.RequestParam{
		Body: node,
	})
	if err != nil {
		return nil, err
	}

	contType := resp.Header.Get("Content-Type")
	if strings.Contains(contType, myhttp.ContentTypeJSON) {

		var codeResp response[[]ResourceData[any]]
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
