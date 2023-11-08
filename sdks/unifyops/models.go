package uopsdk

import (
	"gitlink.org.cn/cloudream/common/pkgs/mq"
	"gitlink.org.cn/cloudream/common/pkgs/types"
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

type SlwNodeID int64

type SlwNode struct {
	ID          SlwNodeID `json:"id"`
	Name        string    `json:"name"`
	SlwRegionID int64     `json:"slwRegionID"`
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

type UnitValue[T any] struct {
	Unit  string `json:"unit"`
	Value T      `json:"value"`
}

type CPUResourceData struct {
	ResourceDataBase
	Name      ResourceType     `json:"name" union:"CPU"`
	Total     UnitValue[int64] `json:"total"`
	Available UnitValue[int64] `json:"available"`
}

func NewCPUResourceData(total UnitValue[int64], available UnitValue[int64]) *CPUResourceData {
	return &CPUResourceData{
		Name:      ResourceTypeCPU,
		Total:     total,
		Available: available,
	}
}

type NPUResourceData struct {
	ResourceDataBase
	Name      ResourceType     `json:"name" union:"NPU"`
	Total     UnitValue[int64] `json:"total"`
	Available UnitValue[int64] `json:"available"`
}

func NewNPUResourceData(total UnitValue[int64], available UnitValue[int64]) *NPUResourceData {
	return &NPUResourceData{
		Name:      ResourceTypeNPU,
		Total:     total,
		Available: available,
	}
}

type GPUResourceData struct {
	ResourceDataBase
	Name      ResourceType     `json:"name" union:"GPU"`
	Total     UnitValue[int64] `json:"total"`
	Available UnitValue[int64] `json:"available"`
}

func NewGPUResourceData(total UnitValue[int64], available UnitValue[int64]) *GPUResourceData {
	return &GPUResourceData{
		Name:      ResourceTypeGPU,
		Total:     total,
		Available: available,
	}
}

type MLUResourceData struct {
	ResourceDataBase
	Name      ResourceType     `json:"name" union:"MLU"`
	Total     UnitValue[int64] `json:"total"`
	Available UnitValue[int64] `json:"available"`
}

func NewMLUResourceData(total UnitValue[int64], available UnitValue[int64]) *MLUResourceData {
	return &MLUResourceData{
		Name:      ResourceTypeMLU,
		Total:     total,
		Available: available,
	}
}

type StorageResourceData struct {
	ResourceDataBase
	Name      ResourceType       `json:"name" union:"STORAGE"`
	Total     UnitValue[float64] `json:"total"`
	Available UnitValue[float64] `json:"available"`
}

func NewStorageResourceData(total UnitValue[float64], available UnitValue[float64]) *StorageResourceData {
	return &StorageResourceData{
		Name:      ResourceTypeStorage,
		Total:     total,
		Available: available,
	}
}

type MemoryResourceData struct {
	ResourceDataBase
	Name      ResourceType       `json:"name" union:"MEMORY"`
	Total     UnitValue[float64] `json:"total"`
	Available UnitValue[float64] `json:"available"`
}

func NewMemoryResourceData(total UnitValue[float64], available UnitValue[float64]) *MemoryResourceData {
	return &MemoryResourceData{
		Name:      ResourceTypeMemory,
		Total:     total,
		Available: available,
	}
}
