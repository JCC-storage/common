package models

/// TODO 将分散在各处的公共结构体定义集中到这里来

const (
	RedundancyRep = "rep"
	RedundancyEC  = "ec"
)

type RedundancyConfigTypes interface{}
type RedundancyConfigTypesConst interface {
	RepRedundancyConfig | ECRedundancyConfig
}
type RepRedundancyConfig struct {
	RepCount int `json:"repCount"`
}

type ECRedundancyConfig struct {
}