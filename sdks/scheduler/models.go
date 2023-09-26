package schsdk

import (
	"gitlink.org.cn/cloudream/common/pkgs/types"
	stgsdk "gitlink.org.cn/cloudream/common/sdks/storage"
	myreflect "gitlink.org.cn/cloudream/common/utils/reflect"
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

type ImageID string

type JobSetInfo struct {
	Jobs []JobInfo `json:"jobs"`
}

type JobInfo interface {
	GetLocalJobID() string
}

var JobInfoTypeUnion = types.NewTypeUnion[JobInfo](
	myreflect.TypeOf[NormalJobInfo](),
	myreflect.TypeOf[ResourceJobInfo](),
)
var JobInfoTaggedTypeUnion = serder.NewTaggedTypeUnion(JobInfoTypeUnion, "Type", "type")

type JobInfoBase struct {
	LocalJobID string `json:"localJobID"`
}

func (i *JobInfoBase) GetLocalJobID() string {
	return i.LocalJobID
}

type NormalJobInfo struct {
	JobInfoBase
	Type      string           `json:"type" union:"Normal"`
	Files     JobFilesInfo     `json:"files"`
	Runtime   JobRuntimeInfo   `json:"runtime"`
	Resources JobResourcesInfo `json:"resources"`
}

type ResourceJobInfo struct {
	JobInfoBase
	Type             string                     `json:"type" union:"Resource"`
	BucketID         int64                      `json:"bucketID"`
	Redundancy       stgsdk.TypedRedundancyInfo `json:"redundancy"`
	TargetLocalJobID string                     `json:"targetLocalJobID"`
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
	myreflect.TypeOf[PackageJobFileInfo](),
	myreflect.TypeOf[LocalJobFileInfo](),
	myreflect.TypeOf[ResourceJobFileInfo](),
	myreflect.TypeOf[ImageJobFileInfo](),
)
var FileInfoTaggedTypeUnion = serder.NewTaggedTypeUnion(FileInfoTypeUnion, "Type", "type")

type JobFileInfoBase struct{}

func (i *JobFileInfoBase) Noop() {}

type PackageJobFileInfo struct {
	JobFileInfoBase
	Type      string `json:"type" union:"Package"`
	PackageID int64  `json:"packageID"`
}

type LocalJobFileInfo struct {
	JobFileInfoBase
	Type      string `json:"type" union:"LocalFile"`
	LocalPath string `json:"localPath"`
}

type ResourceJobFileInfo struct {
	JobFileInfoBase
	Type               string `json:"type" union:"Resource"`
	ResourceLocalJobID string `json:"resourceLocalJobID"`
}

type ImageJobFileInfo struct {
	JobFileInfoBase
	Type    string  `json:"type" union:"Image"`
	ImageID ImageID `json:"imageID"`
}

type JobRuntimeInfo struct {
	Command string   `json:"command"`
	Envs    []EnvVar `json:"envs"`
}

type EnvVar struct {
	Var   string `json:"var"`
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

func JobSetInfoFromJSON(data []byte) (*JobSetInfo, error) {
	mp := make(map[string]any)
	if err := serder.JSONToObject(data, &mp); err != nil {
		return nil, err
	}

	var ret JobSetInfo
	err := serder.MapToObject(mp, &ret, serder.MapToObjectOption{
		UnionTypes: []serder.TaggedUnionType{
			JobInfoTaggedTypeUnion,
			FileInfoTaggedTypeUnion,
		},
	})
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

type JobSetFilesUploadScheme struct {
	LocalFileSchemes []LocalFileUploadScheme `json:"localFileUploadSchemes"`
}

type LocalFileUploadScheme struct {
	LocalPath         string `json:"localPath"`
	UploadToStgNodeID *int64 `json:"uploadToStgNodeID"`
}
