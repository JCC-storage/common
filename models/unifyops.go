package models

import (
	"gitlink.org.cn/cloudream/common/pkgs/types"
	myreflect "gitlink.org.cn/cloudream/common/utils/reflect"
	"gitlink.org.cn/cloudream/common/utils/serder"
)

const (
	ResourceTypeCPU     = "CPU"
	ResourceTypeNPU     = "NPU"
	ResourceTypeGPU     = "GPU"
	ResourceTypeMLU     = "MLU"
	ResourceTypeStorage = "STORAGE"
	ResourceTypeMemory  = "MEMORY"
)

type SlwNode struct {
	ID           int64  `json:"ID"`
	Name         string `json:"name"`
	SlwRegionID  int64  `json:"slwRegionID"`
	StgNodeID    int64  `json:"stgNodeID"`
	StorageID    int64  `json:"StorageID"`
}

type ResourceData interface{}
type ResourceDataConst interface {
	ResourceData | CPUResourceData | NPUResourceData | GPUResourceData | MLUResourceData | StorageResourceData | MemoryResourceData
}

var ResourceDataTypeUnion = types.NewTypeUnion[ResourceData](
	myreflect.TypeOf[CPUResourceData](),
	myreflect.TypeOf[NPUResourceData](),
	myreflect.TypeOf[GPUResourceData](),
	myreflect.TypeOf[MLUResourceData](),
	myreflect.TypeOf[StorageResourceData](),
	myreflect.TypeOf[MemoryResourceData](),
)
var ResourceDataTaggedTypeUnion = serder.NewTaggedTypeUnion(ResourceDataTypeUnion, "Name", "name")

type DetailType[T any] struct {
	Unit  string `json:"unit"`
	Value T      `json:"value"`
}

type CPUResourceData struct {
	Name      string            `json:"name" union:"CPU"`
	Total     DetailType[int64] `json:"total"`
	Available DetailType[int64] `json:"available"`
}

func NewCPUResourceData(name string, total DetailType[int64], available DetailType[int64]) CPUResourceData {
	return CPUResourceData{
		Name:      name,
		Total:     total,
		Available: available,
	}
}

type NPUResourceData struct {
	Name      string            `json:"name" union:"NPU"`
	Total     DetailType[int64] `json:"total"`
	Available DetailType[int64] `json:"available"`
}

func NewNPUResourceData(name string, total DetailType[int64], available DetailType[int64]) NPUResourceData {
	return NPUResourceData{
		Name:      name,
		Total:     total,
		Available: available,
	}
}

type GPUResourceData struct {
	Name      string            `json:"name" union:"GPU"`
	Total     DetailType[int64] `json:"total"`
	Available DetailType[int64] `json:"available"`
}

func NewGPUResourceData(name string, total DetailType[int64], available DetailType[int64]) GPUResourceData {
	return GPUResourceData{
		Name:      name,
		Total:     total,
		Available: available,
	}
}

type MLUResourceData struct {
	Name      string            `json:"name" union:"MLU"`
	Total     DetailType[int64] `json:"total"`
	Available DetailType[int64] `json:"available"`
}

func NewMLUResourceData(name string, total DetailType[int64], available DetailType[int64]) MLUResourceData {
	return MLUResourceData{
		Name:      name,
		Total:     total,
		Available: available,
	}
}

type StorageResourceData struct {
	Name      string              `json:"name" union:"STORAGE"`
	Total     DetailType[float64] `json:"total"`
	Available DetailType[float64] `json:"available"`
}

func NewStorageResourceData(name string, total DetailType[float64], available DetailType[float64]) StorageResourceData {
	return StorageResourceData{
		Name:      name,
		Total:     total,
		Available: available,
	}
}

type MemoryResourceData struct {
	Name      string              `json:"name" union:"MEMORY"`
	Total     DetailType[float64] `json:"total"`
	Available DetailType[float64] `json:"available"`
}

func NewMemoryResourceData(name string, total DetailType[float64], available DetailType[float64]) MemoryResourceData {
	return MemoryResourceData{
		Name:      name,
		Total:     total,
		Available: available,
	}
}
