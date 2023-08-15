package models

/// TODO 将分散在各处的公共结构体定义集中到这里来

const (
	RedundancyRep = "rep"
	RedundancyEC  = "ec"
)

// 冗余模式的描述信息。
// 注：如果在mq中的消息结构体使用了此类型，记得使用RegisterTypeSet注册相关的类型。
type RedundancyInfo interface{}
type RedundancyInfoConst interface {
	RepRedundancyInfo | ECRedundancyInfo | RedundancyInfo
}
type RepRedundancyInfo struct {
	RepCount int `json:"repCount"`
}

func NewRepRedundancyInfo(repCount int) RepRedundancyInfo {
	return RepRedundancyInfo{
		RepCount: repCount,
	}
}

type ECRedundancyInfo struct {
}
