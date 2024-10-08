package uopsdk

import (
	"fmt"
	"net/url"
	"strings"

	"gitlink.org.cn/cloudream/common/utils/http2"
	"gitlink.org.cn/cloudream/common/utils/serder"
)

const CORRECT_CODE int = 200

func (c *Client) GetAllSlwNodeInfo() ([]SlwNode, error) {
	url, err := url.JoinPath(c.baseURL, "/cmdb/resApi/getSlwNodeInfo")
	if err != nil {
		return nil, err
	}
	resp, err := http2.GetJSON(url, http2.RequestParam{})
	if err != nil {
		return nil, err
	}
	contType := resp.Header.Get("Content-Type")
	if strings.Contains(contType, http2.ContentTypeJSON) {

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
	resp, err := http2.PostJSON(url, http2.RequestParam{
		Body: node,
	})
	if err != nil {
		return nil, err
	}

	contType := resp.Header.Get("Content-Type")
	if strings.Contains(contType, http2.ContentTypeJSON) {

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
	resp, err := http2.PostJSON(url, http2.RequestParam{
		Body: node,
	})
	if err != nil {
		return nil, err
	}

	contType := resp.Header.Get("Content-Type")
	if strings.Contains(contType, http2.ContentTypeJSON) {

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
	resp, err := http2.PostJSON(url, http2.RequestParam{
		Body: node,
	})
	if err != nil {
		return nil, err
	}

	contType := resp.Header.Get("Content-Type")
	if strings.Contains(contType, http2.ContentTypeJSON) {

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
	resp, err := http2.PostJSON(url, http2.RequestParam{
		Body: node,
	})
	if err != nil {
		return nil, err
	}

	contType := resp.Header.Get("Content-Type")
	if strings.Contains(contType, http2.ContentTypeJSON) {

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
	resp, err := http2.PostJSON(url, http2.RequestParam{
		Body: node,
	})
	if err != nil {
		return nil, err
	}

	contType := resp.Header.Get("Content-Type")
	if strings.Contains(contType, http2.ContentTypeJSON) {

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
	resp, err := http2.PostJSON(url, http2.RequestParam{
		Body: node,
	})
	if err != nil {
		return nil, err
	}

	contType := resp.Header.Get("Content-Type")
	if strings.Contains(contType, http2.ContentTypeJSON) {

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
	if node.SlwNodeID == 1 {
		return mockData1()
	}

	if node.SlwNodeID == 2 {
		return mockData2()
	}

	return mockData3()
}

func mockData1() (*[]ResourceData, error) {
	var ret []ResourceData

	cpuResourceData := CPUResourceData{
		Name: ResourceTypeCPU,
		Total: UnitValue[int64]{
			Value: 100,
			Unit:  "",
		},
		Available: UnitValue[int64]{
			Value: 100,
			Unit:  "",
		},
	}
	ret = append(ret, &cpuResourceData)

	npuResourceData := NPUResourceData{
		Name: ResourceTypeNPU,
		Total: UnitValue[int64]{
			Value: 0,
			Unit:  "",
		},
		Available: UnitValue[int64]{
			Value: 0,
			Unit:  "",
		},
	}
	ret = append(ret, &npuResourceData)

	gpuResourceData := GPUResourceData{
		Name: ResourceTypeGPU,
		Total: UnitValue[int64]{
			Value: 0,
			Unit:  "",
		},
		Available: UnitValue[int64]{
			Value: 0,
			Unit:  "",
		},
	}
	ret = append(ret, &gpuResourceData)

	mluResourceData := MLUResourceData{
		Name: ResourceTypeMLU,
		Total: UnitValue[int64]{
			Value: 100,
			Unit:  "",
		},
		Available: UnitValue[int64]{
			Value: 100,
			Unit:  "",
		},
	}
	ret = append(ret, &mluResourceData)

	storageResourceData := StorageResourceData{
		Name: ResourceTypeStorage,
		Total: UnitValue[float64]{
			Value: 100,
			Unit:  "GB",
		},
		Available: UnitValue[float64]{
			Value: 100,
			Unit:  "GB",
		},
	}
	ret = append(ret, &storageResourceData)

	memoryResourceData := MemoryResourceData{
		Name: ResourceTypeMemory,
		Total: UnitValue[float64]{
			Value: 100,
			Unit:  "GB",
		},
		Available: UnitValue[float64]{
			Value: 100,
			Unit:  "GB",
		},
	}
	ret = append(ret, &memoryResourceData)

	return &ret, nil
}

func mockData2() (*[]ResourceData, error) {
	var ret []ResourceData

	cpuResourceData := CPUResourceData{
		Name: ResourceTypeCPU,
		Total: UnitValue[int64]{
			Value: 100,
			Unit:  "",
		},
		Available: UnitValue[int64]{
			Value: 100,
			Unit:  "",
		},
	}
	ret = append(ret, &cpuResourceData)

	npuResourceData := NPUResourceData{
		Name: ResourceTypeNPU,
		Total: UnitValue[int64]{
			Value: 100,
			Unit:  "",
		},
		Available: UnitValue[int64]{
			Value: 100,
			Unit:  "",
		},
	}
	ret = append(ret, &npuResourceData)

	gpuResourceData := GPUResourceData{
		Name: ResourceTypeGPU,
		Total: UnitValue[int64]{
			Value: 0,
			Unit:  "",
		},
		Available: UnitValue[int64]{
			Value: 0,
			Unit:  "",
		},
	}
	ret = append(ret, &gpuResourceData)

	mluResourceData := MLUResourceData{
		Name: ResourceTypeMLU,
		Total: UnitValue[int64]{
			Value: 0,
			Unit:  "",
		},
		Available: UnitValue[int64]{
			Value: 0,
			Unit:  "",
		},
	}
	ret = append(ret, &mluResourceData)

	storageResourceData := StorageResourceData{
		Name: ResourceTypeStorage,
		Total: UnitValue[float64]{
			Value: 100,
			Unit:  "GB",
		},
		Available: UnitValue[float64]{
			Value: 100,
			Unit:  "GB",
		},
	}
	ret = append(ret, &storageResourceData)

	memoryResourceData := MemoryResourceData{
		Name: ResourceTypeMemory,
		Total: UnitValue[float64]{
			Value: 100,
			Unit:  "GB",
		},
		Available: UnitValue[float64]{
			Value: 100,
			Unit:  "GB",
		},
	}
	ret = append(ret, &memoryResourceData)

	return &ret, nil
}

func mockData3() (*[]ResourceData, error) {
	var ret []ResourceData

	cpuResourceData := CPUResourceData{
		Name: ResourceTypeCPU,
		Total: UnitValue[int64]{
			Value: 100,
			Unit:  "",
		},
		Available: UnitValue[int64]{
			Value: 100,
			Unit:  "",
		},
	}
	ret = append(ret, &cpuResourceData)

	npuResourceData := NPUResourceData{
		Name: ResourceTypeNPU,
		Total: UnitValue[int64]{
			Value: 0,
			Unit:  "",
		},
		Available: UnitValue[int64]{
			Value: 0,
			Unit:  "",
		},
	}
	ret = append(ret, &npuResourceData)

	gpuResourceData := GPUResourceData{
		Name: ResourceTypeGPU,
		Total: UnitValue[int64]{
			Value: 100,
			Unit:  "",
		},
		Available: UnitValue[int64]{
			Value: 100,
			Unit:  "",
		},
	}
	ret = append(ret, &gpuResourceData)

	mluResourceData := MLUResourceData{
		Name: ResourceTypeMLU,
		Total: UnitValue[int64]{
			Value: 0,
			Unit:  "",
		},
		Available: UnitValue[int64]{
			Value: 0,
			Unit:  "",
		},
	}
	ret = append(ret, &mluResourceData)

	storageResourceData := StorageResourceData{
		Name: ResourceTypeStorage,
		Total: UnitValue[float64]{
			Value: 100,
			Unit:  "GB",
		},
		Available: UnitValue[float64]{
			Value: 100,
			Unit:  "GB",
		},
	}
	ret = append(ret, &storageResourceData)

	memoryResourceData := MemoryResourceData{
		Name: ResourceTypeMemory,
		Total: UnitValue[float64]{
			Value: 100,
			Unit:  "GB",
		},
		Available: UnitValue[float64]{
			Value: 100,
			Unit:  "GB",
		},
	}
	ret = append(ret, &memoryResourceData)

	return &ret, nil
}
