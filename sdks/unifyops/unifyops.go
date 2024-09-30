package uopsdk

import (
	"fmt"
	"net/url"
	"strings"

	myhttp "gitlink.org.cn/cloudream/common/utils/http"
	"gitlink.org.cn/cloudream/common/utils/serder"
)

const CORRECT_CODE int = 200

func (c *Client) GetAllSlwNodeInfo() ([]SlwNode, error) {
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
			return codeResp.Data, nil
		}

		return nil, codeResp.ToError()
	}

	return nil, fmt.Errorf("unknow response content type: %s", contType)
}

type GetOneResourceDataReq struct {
	SlwNodeID SlwNodeID `json:"nodeId"`
}

func (c *Client) GetCPUData(node GetOneResourceDataReq) (*CPUResourceData, error) {
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

		var codeResp response[CPUResourceData]
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

func (c *Client) GetNPUData(node GetOneResourceDataReq) (*NPUResourceData, error) {
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

		var codeResp response[NPUResourceData]
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

func (c *Client) GetGPUData(node GetOneResourceDataReq) (*GPUResourceData, error) {
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

		var codeResp response[GPUResourceData]
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

func (c *Client) GetMLUData(node GetOneResourceDataReq) (*MLUResourceData, error) {
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

		var codeResp response[MLUResourceData]
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

func (c *Client) GetStorageData(node GetOneResourceDataReq) (*StorageResourceData, error) {
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

		var codeResp response[StorageResourceData]
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

func (c *Client) GetMemoryData(node GetOneResourceDataReq) (*MemoryResourceData, error) {
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

		var codeResp response[MemoryResourceData]
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

func (c *Client) GetIndicatorData(node GetOneResourceDataReq) (*[]ResourceData, error) {
	switch node.SlwNodeID {
	case 1:
		return shuguang()
	case 2:
		return modelarts()
	case 3:
		return hanwuji()
	}
	return nil, nil
}

//func (c *Client) GetIndicatorData(node GetOneResourceDataReq) (*[]ResourceData, error) {
//url, err := url.JoinPath(c.baseURL, "/cmdb/resApi/getIndicatorData")
//if err != nil {
//	return nil, err
//}
//resp, err := myhttp.PostJSON(url, myhttp.RequestParam{
//	Body: node,
//})
//if err != nil {
//	return nil, err
//}
//
//contType := resp.Header.Get("Content-Type")
//if strings.Contains(contType, myhttp.ContentTypeJSON) {
//
//	var codeResp response[[]map[string]any]
//	if err := serder.JSONToObjectStream(resp.Body, &codeResp); err != nil {
//		return nil, fmt.Errorf("parsing response: %w", err)
//	}
//
//	if codeResp.Code != CORRECT_CODE {
//		return nil, codeResp.ToError()
//	}
//
//	var ret []ResourceData
//	for _, mp := range codeResp.Data {
//		var data ResourceData
//		err := serder.MapToObject(mp, &data)
//		if err != nil {
//			return nil, err
//		}
//		ret = append(ret, data)
//	}
//
//	return &ret, nil
//}
//
//return nil, fmt.Errorf("unknow response content type: %s", contType)
//}
