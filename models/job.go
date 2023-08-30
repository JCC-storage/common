package models

import (
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

type JobSetInfo struct {
	Jobs []JobInfo `json:"jobs"`
}

type JobInfo interface{}

var JobInfoTypeUnion = serder.NewTypeUnion[JobInfo]("type",
	serder.NewStringTypeResolver().
		Add(JobTypeNormal, myreflect.TypeOf[NormalJobInfo]()).
		Add(JobTypeResource, myreflect.TypeOf[ResourceJobInfo]()),
)

type NormalJobInfo struct {
	LocalJobID string           `json:"localJobID"`
	Type       string           `json:"type"`
	Files      JobFilesInfo     `json:"files"`
	Runtime    JobRuntimeInfo   `json:"runtime"`
	Resources  JobResourcesInfo `json:"resources"`
}

type ResourceJobInfo struct {
	LocalJobID       string `json:"localJobID"`
	Type             string `json:"type"`
	TargetLocalJobID string `json:"targetLocalJobID"`
}

type JobFilesInfo struct {
	Dateset FileInfo `json:"dataset"`
	Code    FileInfo `json:"code"`
	Image   FileInfo `json:"image"`
}

type FileInfo interface{}

var FileInfoTypeUnion = serder.NewTypeUnion[JobInfo]("type",
	serder.NewStringTypeResolver().
		Add(FileInfoTypePackage, myreflect.TypeOf[PackageFileInfo]()).
		Add(FileInfoTypeLocalFile, myreflect.TypeOf[LocalFileInfo]()).
		Add(FileInfoTypeResource, myreflect.TypeOf[ResourceFileInfo]()).
		Add(FileInfoTypeImage, myreflect.TypeOf[ImageFileInfo]()),
)

type PackageFileInfo struct {
	Type      string `json:"type"`
	PackageID int64  `json:"packageID"`
}

type LocalFileInfo struct {
	Type      string `json:"type"`
	LocalPath string `json:"localPath"`
}

type ResourceFileInfo struct {
	Type               string `json:"type"`
	ResourceLocalJobID string `json:"resourceLocalJobID"`
}

type ImageFileInfo struct {
	Type    string `json:"type"`
	ImageID string `json:"imageID"`
}

type JobRuntimeInfo struct {
	Command string   `json:"command"`
	Envs    []EnvVar `json:"envs"`
}

type EnvVar struct {
	Var   string `json:"var"`
	Value string `json:"value"`
}

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
		UnionTypes: []serder.UnionTypeInfo{
			JobInfoTypeUnion,
			FileInfoTypeUnion,
		},
	})
	if err != nil {
		return nil, err
	}

	return &ret, nil
}
