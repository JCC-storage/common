package models

/// TODO 将分散在各处的公共结构体定义集中到这里来

const (
	RedundancyRep = "rep"
	RedundancyEC  = "ec"
)

// 冗余模式的描述信息
type RedundancyConfigTypes interface{}
type RedundancyConfigTypesConst interface {
	RepRedundancyConfig | ECRedundancyConfig
}
type RepRedundancyConfig struct {
	RepCount int `json:"repCount"`
}

func NewRepRedundancyConfig(repCount int) RepRedundancyConfig {
	return RepRedundancyConfig{
		RepCount: repCount,
	}
}

type ECRedundancyConfig struct {
}
