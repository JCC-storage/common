package schsdk

import (
	"gitlink.org.cn/cloudream/common/pkgs/types"
	cdssdk "gitlink.org.cn/cloudream/common/sdks/storage"
	"gitlink.org.cn/cloudream/common/utils/serder"
)

const (
	JobTypeNormal   = "Normal"
	JobTypeResource = "Resource"

	FileInfoTypePackage   = "Package"
	FileInfoTypeLocalFile = "LocalFile"
	FileInfoTypeResource  = "Resource"
	FileInfoTypeImage     = "Image"
)

type JobID string

type JobSetID string

type ImageID int64

// 计算中心ID
type CCID int64

type JobSetInfo struct {
	Jobs []JobInfo `json:"jobs"`
}

type JobInfo interface {
	GetLocalJobID() string
}

var JobInfoTypeUnion = types.NewTypeUnion[JobInfo](
	(*NormalJobInfo)(nil),
	(*DataReturnJobInfo)(nil),
	(*MultiInstanceJobInfo)(nil),
)
var _ = serder.UseTypeUnionInternallyTagged(&JobInfoTypeUnion, "type")

type JobInfoBase struct {
	LocalJobID string `json:"localJobID"`
}

func (i *JobInfoBase) GetLocalJobID() string {
	return i.LocalJobID
}

type NormalJobInfo struct {
	serder.Metadata `union:"Normal"`
	JobInfoBase
	Type      string           `json:"type"`
	Files     JobFilesInfo     `json:"files"`
	Runtime   JobRuntimeInfo   `json:"runtime"`
	Resources JobResourcesInfo `json:"resources"`
	Services  JobServicesInfo  `json:"services"`
}

type CommonJobInfo struct {
	serder.Metadata `union:"Normal"`
	JobInfoBase
	Type      string           `json:"type"`
	Files     JobFilesInfo     `json:"files"`
	Runtime   JobRuntimeInfo   `json:"runtime"`
	Resources JobResourcesInfo `json:"resources"`
	Services  JobServicesInfo  `json:"services"`
}

type DataReturnJobInfo struct {
	serder.Metadata `union:"DataReturn"`
	JobInfoBase
	Type             string          `json:"type"`
	BucketID         cdssdk.BucketID `json:"bucketID"`
	TargetLocalJobID string          `json:"targetLocalJobID"`
}

type MultiInstanceJobInfo struct {
	serder.Metadata `union:"MultiInstance"`
	JobInfoBase
	Type      string           `json:"type"`
	Files     JobFilesInfo     `json:"files"`
	Runtime   JobRuntimeInfo   `json:"runtime"`
	Resources JobResourcesInfo `json:"resources"`
}

type InstanceJobInfo struct {
	serder.Metadata `union:"SubJob"`
	JobInfoBase
	Type       string           `json:"type"`
	LocalJobID string           `json:"multiInstJobID"`
	Files      JobFilesInfo     `json:"files"`
	Runtime    JobRuntimeInfo   `json:"runtime"`
	Resources  JobResourcesInfo `json:"resources"`
}

type JobFilesInfo struct {
	Dataset JobFileInfo `json:"dataset"`
	Code    JobFileInfo `json:"code"`
	Image   JobFileInfo `json:"image"`
}

type JobFileInfo interface {
	Noop()
}

var FileInfoTypeUnion = types.NewTypeUnion[JobFileInfo](
	(*PackageJobFileInfo)(nil),
	(*LocalJobFileInfo)(nil),
	(*DataReturnJobFileInfo)(nil),
	(*ImageJobFileInfo)(nil),
)
var _ = serder.UseTypeUnionInternallyTagged(&FileInfoTypeUnion, "type")

type JobFileInfoBase struct{}

func (i *JobFileInfoBase) Noop() {}

type PackageJobFileInfo struct {
	serder.Metadata `union:"Package"`
	JobFileInfoBase
	Type      string           `json:"type"`
	PackageID cdssdk.PackageID `json:"packageID"`
}

type LocalJobFileInfo struct {
	serder.Metadata `union:"LocalFile"`
	JobFileInfoBase
	Type      string `json:"type"`
	LocalPath string `json:"localPath"`
}

type DataReturnJobFileInfo struct {
	serder.Metadata `union:"DataReturn"`
	JobFileInfoBase
	Type                 string `json:"type"`
	DataReturnLocalJobID string `json:"dataReturnLocalJobID"`
}

type ImageJobFileInfo struct {
	serder.Metadata `union:"Image"`
	JobFileInfoBase
	Type    string  `json:"type"`
	ImageID ImageID `json:"imageID"`
}

type JobRuntimeInfo struct {
	Command string   `json:"command"`
	Envs    []KVPair `json:"envs"`
}

type KVPair struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// CPU、GPU、NPU、MLU单位为：核
// Storage、Memory单位为：字节
type JobResourcesInfo struct {
	CPU     float64 `json:"cpu"`
	GPU     float64 `json:"gpu"`
	NPU     float64 `json:"npu"`
	MLU     float64 `json:"mlu"`
	Storage int64   `json:"storage"`
	Memory  int64   `json:"memory"`
}

type JobSetFilesUploadScheme struct {
	LocalFileSchemes []LocalFileUploadScheme `json:"localFileUploadSchemes"`
}

type JobFilesUploadScheme struct {
	LocalFileSchemes []LocalFileUploadScheme `json:"localFileUploadSchemes"`
}

type LocalFileUploadScheme struct {
	LocalPath         string         `json:"localPath"`
	UploadToCDSNodeID *cdssdk.NodeID `json:"uploadToCDSNodeID"`
}

type JobServicesInfo struct {
	ServicePortInfos []ServicePortInfo `json:"servicePortInfos"`
}

type ServicePortInfo struct {
	Name string `json:"name"`
	Port int64  `json:"port"`
}

type JobSetServiceInfo struct {
	Name       string         `json:"name"`
	Port       int64          `json:"port"`
	CDSNodeID  *cdssdk.NodeID `json:"cdsNodeID"`
	LocalJobID string         `json:"localJobID"`
}
