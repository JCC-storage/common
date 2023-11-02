package uopsdk

import (
	"gitlink.org.cn/cloudream/common/pkgs/mq"
	"gitlink.org.cn/cloudream/common/pkgs/types"
	schsdk "gitlink.org.cn/cloudream/common/sdks/scheduler"
	"gitlink.org.cn/cloudream/common/utils/serder"
)

type ResourceType string

const (
	ResourceTypeCPU     ResourceType = "CPU"
	ResourceTypeNPU     ResourceType = "NPU"
	ResourceTypeGPU     ResourceType = "GPU"
	ResourceTypeMLU     ResourceType = "MLU"
	ResourceTypeStorage ResourceType = "STORAGE"
	ResourceTypeMemory  ResourceType = "MEMORY"
)

type SlwNode struct {
	ID          schsdk.SlwNodeID `json:"ID"`
	Name        string           `json:"name"`
	SlwRegionID int64            `json:"slwRegionID"`
	StgNodeID   int64            `json:"stgNodeID"`
	StorageID   int64            `json:"StorageID"`
}

type ResourceData interface {
	Noop()
}

var ResourceDataTypeUnion = types.NewTypeUnion[ResourceData](
	(*CPUResourceData)(nil),
	(*NPUResourceData)(nil),
	(*GPUResourceData)(nil),
	(*MLUResourceData)(nil),
	(*StorageResourceData)(nil),
	(*MemoryResourceData)(nil),
)
var _ = serder.RegisterNewTaggedTypeUnion(ResourceDataTypeUnion, "Name", "name")
var _ = mq.RegisterUnionType(ResourceDataTypeUnion)

type ResourceDataBase struct{}

func (d *ResourceDataBase) Noop() {}

type DetailType[T any] struct {
	Unit  string `json:"unit"`
	Value T      `json:"value"`
}

type CPUResourceData struct {
	ResourceDataBase
	Name      ResourceType      `json:"name" union:"CPU"`
	Total     DetailType[int64] `json:"total"`
	Available DetailType[int64] `json:"available"`
}

func NewCPUResourceData(total DetailType[int64], available DetailType[int64]) *CPUResourceData {
	return &CPUResourceData{
		Name:      ResourceTypeCPU,
		Total:     total,
		Available: available,
	}
}

type NPUResourceData struct {
	ResourceDataBase
	Name      ResourceType      `json:"name" union:"NPU"`
	Total     DetailType[int64] `json:"total"`
	Available DetailType[int64] `json:"available"`
}

func NewNPUResourceData(total DetailType[int64], available DetailType[int64]) *NPUResourceData {
	return &NPUResourceData{
		Name:      ResourceTypeNPU,
		Total:     total,
		Available: available,
	}
}

type GPUResourceData struct {
	ResourceDataBase
	Name      ResourceType      `json:"name" union:"GPU"`
	Total     DetailType[int64] `json:"total"`
	Available DetailType[int64] `json:"available"`
}

func NewGPUResourceData(total DetailType[int64], available DetailType[int64]) *GPUResourceData {
	return &GPUResourceData{
		Name:      ResourceTypeGPU,
		Total:     total,
		Available: available,
	}
}

type MLUResourceData struct {
	ResourceDataBase
	Name      ResourceType      `json:"name" union:"MLU"`
	Total     DetailType[int64] `json:"total"`
	Available DetailType[int64] `json:"available"`
}

func NewMLUResourceData(total DetailType[int64], available DetailType[int64]) *MLUResourceData {
	return &MLUResourceData{
		Name:      ResourceTypeMLU,
		Total:     total,
		Available: available,
	}
}

type StorageResourceData struct {
	ResourceDataBase
	Name      ResourceType        `json:"name" union:"STORAGE"`
	Total     DetailType[float64] `json:"total"`
	Available DetailType[float64] `json:"available"`
}

func NewStorageResourceData(total DetailType[float64], available DetailType[float64]) *StorageResourceData {
	return &StorageResourceData{
		Name:      ResourceTypeStorage,
		Total:     total,
		Available: available,
	}
}

type MemoryResourceData struct {
	ResourceDataBase
	Name      ResourceType        `json:"name" union:"MEMORY"`
	Total     DetailType[float64] `json:"total"`
	Available DetailType[float64] `json:"available"`
}

func NewMemoryResourceData(total DetailType[float64], available DetailType[float64]) *MemoryResourceData {
	return &MemoryResourceData{
		Name:      ResourceTypeMemory,
		Total:     total,
		Available: available,
	}
}
