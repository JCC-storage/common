package unifyops

import (
	"fmt"
	"net/url"
	"strings"

	"gitlink.org.cn/cloudream/common/models"
	myhttp "gitlink.org.cn/cloudream/common/utils/http"
	"gitlink.org.cn/cloudream/common/utils/serder"
)

const CORRECT_CODE int = 200

type SlwNodeInfo struct {
	Nodes []models.SlwNode `json:"nodes"`
}

func (c *Client) GetSlwNodeInfo() (*SlwNodeInfo, error) {
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

		var codeResp response[SlwNodeInfo]
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

func (c *Client) GetCPUData(node Node) (*models.CPUResourceData, error) {
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

		var codeResp response[models.CPUResourceData]
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

func (c *Client) GetNPUData(node Node) (*models.NPUResourceData, error) {
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

		var codeResp response[models.NPUResourceData]
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

func (c *Client) GetGPUData(node Node) (*models.GPUResourceData, error) {
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

		var codeResp response[models.GPUResourceData]
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

func (c *Client) GetMLUData(node Node) (*models.MLUResourceData, error) {
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

		var codeResp response[models.MLUResourceData]
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

func (c *Client) GetStorageData(node Node) (*models.StorageResourceData, error) {
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

		var codeResp response[models.StorageResourceData]
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

func (c *Client) GetMemoryData(node Node) (*models.MemoryResourceData, error) {
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

		var codeResp response[models.MemoryResourceData]
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

func (c *Client) GetIndicatorData(node Node) (*[]models.ResourceData, error) {
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

		var codeResp response[[]map[string]any]
		if err := serder.JSONToObjectStream(resp.Body, &codeResp); err != nil {
			return nil, fmt.Errorf("parsing response: %w", err)
		}

		if codeResp.Code != CORRECT_CODE {
			return nil, codeResp.ToError()
		}

		mapToObjOpt := serder.MapToObjectOption{
			UnionTypes: []serder.TaggedUnionType{
				models.ResourceDataTaggedTypeUnion,
			},
		}
		var ret []models.ResourceData
		for _, mp := range codeResp.Data {
			var data models.ResourceData
			err := serder.MapToObject(mp, &data, mapToObjOpt)
			if err != nil {
				return nil, err
			}
			ret = append(ret, data)
		}

		return &ret, nil
	}

	return nil, fmt.Errorf("unknow response content type: %s", contType)
}
