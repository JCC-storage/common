package sch

import (
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

	CODE    = "code"
	DATASET = "dataset"
	IMAGE   = "image"
	MODEL   = "model"
)

type TaskID int64

type ClusterDetail struct {
	// 集群ID
	ClusterId schsdk.ClusterID `json:"clusterID"`
	// 集群功能类型：云算，智算，超算
	ClusterType string `json:"clusterType"`
	// 集群地区：华东地区、华南地区、华北地区、华中地区、西南地区、西北地区、东北地区
	Region string `json:"region"`
	// 资源类型
	Resources2 []ResourceData `json:"resources1,omitempty"`
	//Resources []ResourceData `json:"resources"`
	Resources []TmpResourceData `json:"resources"`
}

type TmpResourceData struct {
	Type      ResourceType       `json:"type"`
	Name      string             `json:"name"`
	Total     UnitValue[float64] `json:"total"`
	Available UnitValue[float64] `json:"available"`
}

type ResourceData interface {
	Noop()
}

var ResourceDataTypeUnion = types.NewTypeUnion[ResourceData](
	(*CPUResourceData)(nil),
	(*NPUResourceData)(nil),
	(*GPUResourceData)(nil),
	(*MLUResourceData)(nil),
	(*DCUResourceData)(nil),
	(*GCUResourceData)(nil),
	(*GPGPUResourceData)(nil),
	(*StorageResourceData)(nil),
	(*MemoryResourceData)(nil),
)
var _ = serder.UseTypeUnionInternallyTagged(&ResourceDataTypeUnion, "type")

type ResourceDataBase struct{}

func (d *ResourceDataBase) Noop() {}

type UnitValue[T any] struct {
	Unit  string `json:"unit"`
	Value T      `json:"value"`
}

type CPUResourceData struct {
	serder.Metadata `union:"CPU"`
	ResourceDataBase
	Type      string           `json:"type"`
	Name      ResourceType     `json:"name"`
	Total     UnitValue[int64] `json:"total"`
	Available UnitValue[int64] `json:"available"`
}

type NPUResourceData struct {
	serder.Metadata `union:"NPU"`
	ResourceDataBase
	Type      string           `json:"type"`
	Name      ResourceType     `json:"name"`
	Total     UnitValue[int64] `json:"total"`
	Available UnitValue[int64] `json:"available"`
}

type GPUResourceData struct {
	serder.Metadata `union:"GPU"`
	ResourceDataBase
	Type      string           `json:"type"`
	Name      ResourceType     `json:"name"`
	Total     UnitValue[int64] `json:"total"`
	Available UnitValue[int64] `json:"available"`
}

type MLUResourceData struct {
	serder.Metadata `union:"MLU"`
	ResourceDataBase
	Type      string           `json:"type"`
	Name      ResourceType     `json:"name"`
	Total     UnitValue[int64] `json:"total"`
	Available UnitValue[int64] `json:"available"`
}

type DCUResourceData struct {
	serder.Metadata `union:"DCU"`
	ResourceDataBase
	Type      string           `json:"type"`
	Name      ResourceType     `json:"name"`
	Total     UnitValue[int64] `json:"total"`
	Available UnitValue[int64] `json:"available"`
}

type GCUResourceData struct {
	serder.Metadata `union:"GCU"`
	ResourceDataBase
	Type      string           `json:"type"`
	Name      ResourceType     `json:"name"`
	Total     UnitValue[int64] `json:"total"`
	Available UnitValue[int64] `json:"available"`
}

type GPGPUResourceData struct {
	serder.Metadata `union:"ILUVATAR-GPGPU"`
	ResourceDataBase
	Type      string           `json:"type"`
	Name      ResourceType     `json:"name"`
	Total     UnitValue[int64] `json:"total"`
	Available UnitValue[int64] `json:"available"`
}

type StorageResourceData struct {
	serder.Metadata `union:"STORAGE"`
	ResourceDataBase
	Type      string             `json:"type"`
	Name      ResourceType       `json:"name"`
	Total     UnitValue[float64] `json:"total"`
	Available UnitValue[float64] `json:"available"`
}

type MemoryResourceData struct {
	serder.Metadata `union:"MEMORY"`
	ResourceDataBase
	Type      string             `json:"type"`
	Name      ResourceType       `json:"name"`
	Total     UnitValue[float64] `json:"total"`
	Available UnitValue[float64] `json:"available"`
}

type ResourcePriority interface {
	Noop()
}

type ResourcePriorityBase struct {
}

var ResourcePriorityTypeUnion = types.NewTypeUnion[ResourcePriority](
	(*RegionPriority)(nil),
	(*ChipPriority)(nil),
	(*BiasPriority)(nil),
)

var _ = serder.UseTypeUnionInternallyTagged(&ResourcePriorityTypeUnion, "type")

func (d *ResourcePriorityBase) Noop() {}

type RegionPriority struct {
	serder.Metadata `union:"region"`
	ResourcePriorityBase
	Type    string   `json:"type"`
	Options []string `json:"options"`
}

type ChipPriority struct {
	serder.Metadata `union:"chip"`
	ResourcePriorityBase
	Type    string   `json:"type"`
	Options []string `json:"options"`
}

type BiasPriority struct {
	serder.Metadata `union:"bias"`
	ResourcePriorityBase
	Type    string   `json:"type"`
	Options []string `json:"options"`
}

type UploadParams struct {
	DataType       string         `json:"dataType"`
	DataName       string         `json:"dataName"`
	UploadInfo     UploadInfo     `json:"uploadInfo"`
	UploadPriority UploadPriority `json:"uploadPriority"`
}

type UploadInfo interface {
	Noop()
}

var UploadInfoTypeUnion = types.NewTypeUnion[UploadInfo](
	(*LocalUploadInfo)(nil),
	(*RemoteUploadInfo)(nil),
)

var _ = serder.UseTypeUnionInternallyTagged(&UploadInfoTypeUnion, "type")

type LocalUploadInfo struct {
	serder.Metadata `union:"local"`
	UploadInfoBase
	Type      string `json:"type"`
	LocalPath string `json:"localPath"`
}

type RemoteUploadInfo struct {
	serder.Metadata `union:"url"`
	UploadInfoBase
	Type           string             `json:"type"`
	Url            string             `json:"url"`
	TargetClusters []schsdk.ClusterID `json:"targetClusters"`
}

type UploadInfoBase struct{}

func (d *UploadInfoBase) Noop() {}

type UploadPriority interface {
	Noop()
}

var UploadPriorityTypeUnion = types.NewTypeUnion[UploadPriority](
	(*Preferences)(nil),
	(*SpecifyCluster)(nil),
)

var _ = serder.UseTypeUnionInternallyTagged(&UploadPriorityTypeUnion, "type")

type Preferences struct {
	serder.Metadata `union:"preference"`
	UploadPriorityBase
	Type               string             `json:"type"`
	ResourcePriorities []ResourcePriority `json:"priorities"`
}

type SpecifyCluster struct {
	serder.Metadata `union:"specify"`
	UploadPriorityBase
	Type     string             `json:"type"`
	Clusters []schsdk.ClusterID `json:"clusters"`
}

type UploadPriorityBase struct{}

func (d *UploadPriorityBase) Noop() {}
