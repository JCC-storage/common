package schsdk

// 这个结构体无任何字段，但实现了Noop，每种MessageBody都要内嵌这个结构体
type MessageBodyBase struct{}

// 此处的receiver是指针
func (b *MessageBodyBase) Noop() {}

type AvailableNodesResp struct {
	MessageBodyBase
	AvailableNodes map[ModelID]AvailableNodes `json:"allNode"`
}

type AvailableNodes struct {
	MessageBodyBase
	//ModelID ModelID    `json:"modelID"`
	JobID JobID      `json:"jobID"`
	Nodes []NodeInfo `json:"nodes"`
}

type NodeInfo struct {
	MessageBodyBase
	InstanceID JobID   `json:"instanceID"`
	NodeID     NodeID  `json:"nodeID"`
	Address    Address `json:"address"`
}
