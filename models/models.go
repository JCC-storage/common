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

type RedundancyDataTypes interface{}
type RedundancyDataTypesConst interface {
	RepRedundancyData | ECRedundancyData
}
type RepRedundancyData struct {
	FileHash string `json:"fileHash"`
}

func NewRedundancyRepData(fileHash string) RepRedundancyData {
	return RepRedundancyData{
		FileHash: fileHash,
	}
}

type ECRedundancyData struct {
	Blocks []ObjectBlock `json:"blocks"`
}

func NewECRedundancyData(blocks []ObjectBlock) ECRedundancyData {
	return ECRedundancyData{
		Blocks: blocks,
	}
}

type ObjectBlock struct {
	Index    int    `json:"index"`
	FileHash string `json:"fileHash"`
}

func NewObjectBlock(index int, fileHash string) ObjectBlock {
	return ObjectBlock{
		Index:    index,
		FileHash: fileHash,
	}
}
