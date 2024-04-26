package pcmsdk

type ParticipantID int64

type Participant struct {
	ID   ParticipantID `json:"id"`
	Name string        `json:"name"`
	Type string        `json:"type"`
}

type ImageID string

type Image struct {
	ImageID     ImageID `json:"imageID"`
	ImageName   string  `json:"imageName"`
	ImageStatus string  `json:"imageStatus"`
}

type ResourceID string

type Resource struct {
	ParticipantID   ParticipantID `json:"participantID"`
	ParticipantName string        `json:"participantName"`
	SpecName        string        `json:"specName"`
	SpecID          ResourceID    `json:"specId"`
	SpecPrice       float64       `json:"specPrice"`
}

type TaskID string

type TaskStatus string

const (
	TaskStatusPending TaskStatus = "Pending"
	TaskStatusRunning TaskStatus = "Running"
	TaskStatusSuccess TaskStatus = "succeeded"
	TaskStatusFailed  TaskStatus = "failed"
)
