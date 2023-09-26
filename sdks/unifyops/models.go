package uopsdk

import (
	"gitlink.org.cn/cloudream/common/pkgs/types"
	myreflect "gitlink.org.cn/cloudream/common/utils/reflect"
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

type SlwNodeImageID int64

type SlwNode struct {
	ID          SlwNodeID `json:"ID"`
	Name        string    `json:"name"`
	SlwRegionID int64     `json:"slwRegionID"`
	StgNodeID   int64     `json:"stgNodeID"`
	StorageID   int64     `json:"StorageID"`
}

type ResourceData interface {
	Noop()
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

func NewCPUResourceData(name ResourceType, total DetailType[int64], available DetailType[int64]) *CPUResourceData {
	return &CPUResourceData{
		Name:      name,
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

func NewNPUResourceData(name ResourceType, total DetailType[int64], available DetailType[int64]) *NPUResourceData {
	return &NPUResourceData{
		Name:      name,
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

func NewGPUResourceData(name ResourceType, total DetailType[int64], available DetailType[int64]) *GPUResourceData {
	return &GPUResourceData{
		Name:      name,
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

func NewMLUResourceData(name ResourceType, total DetailType[int64], available DetailType[int64]) *MLUResourceData {
	return &MLUResourceData{
		Name:      name,
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

func NewStorageResourceData(name ResourceType, total DetailType[float64], available DetailType[float64]) *StorageResourceData {
	return &StorageResourceData{
		Name:      name,
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

func NewMemoryResourceData(name ResourceType, total DetailType[float64], available DetailType[float64]) *MemoryResourceData {
	return &MemoryResourceData{
		Name:      name,
		Total:     total,
		Available: available,
	}
}