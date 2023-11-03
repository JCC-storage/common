package pcmsdk

import schsdk "gitlink.org.cn/cloudream/common/sdks/scheduler"

type Participant struct {
	ID   schsdk.SlwNodeID `json:"id"`
	Name string           `json:"name"`
	Type string           `json:"type"`
}

type Image struct {
	ImageID     schsdk.SlwNodeImageID `json:"imageID"`
	ImageName   string                `json:"imageName"`
	ImageStatus string                `json:"imageStatus"`
}

type ResourceID string

type Resource struct {
	ParticipantID   schsdk.SlwNodeID `json:"participantID"`
	ParticipantName string           `json:"participantName"`
	SpecName        string           `json:"specName"`
	SpecID          ResourceID       `json:"specId"`
	SpecPrice       float64          `json:"specPrice"`
}

type TaskID string

type TaskStatus string

const (
	TaskStatusPending TaskStatus = "Pending"
	TaskStatusRunning TaskStatus = "Running"
	TaskStatusSuccess TaskStatus = "Success"
	TaskStatuFailed   TaskStatus = "Failed"
)
