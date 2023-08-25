package models

import (
	"fmt"

	myreflect "gitlink.org.cn/cloudream/common/utils/reflect"
	"gitlink.org.cn/cloudream/common/utils/serder"
)

/// TODO 将分散在各处的公共结构体定义集中到这里来

const (
	RedundancyRep = "rep"
	RedundancyEC  = "ec"
)

// 冗余模式的描述信息。
// 注：如果在mq中的消息结构体使用了此类型，记得使用RegisterTypeSet注册相关的类型。
type RedundancyInfo interface{}
type RedundancyInfoConst interface {
	RepRedundancyInfo | ECRedundancyInfo
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
	ECName     string `json:"ecName"`
	PacketSize int64  `json:"packetSize"`
}

func NewECRedundancyInfo(ecName string, packetSize int64) ECRedundancyInfo {
	return ECRedundancyInfo{
		ECName:     ecName,
		PacketSize: packetSize,
	}
}

type TypedRedundancyInfo struct {
	Type string         `json:"type"`
	Info RedundancyInfo `json:"info"`
}

func NewTypedRedundancyInfo[T RedundancyInfoConst](info T) TypedRedundancyInfo {
	var typ string

	if myreflect.TypeOf[T]() == myreflect.TypeOf[RepRedundancyInfo]() {
		typ = RedundancyRep
	} else if myreflect.TypeOf[T]() == myreflect.TypeOf[ECRedundancyInfo]() {
		typ = RedundancyEC
	}

	return TypedRedundancyInfo{
		Type: typ,
		Info: info,
	}
}
func NewTypedRepRedundancyInfo(repCount int) TypedRedundancyInfo {
	return TypedRedundancyInfo{
		Type: RedundancyRep,
		Info: RepRedundancyInfo{
			RepCount: repCount,
		},
	}
}

func NewTypedECRedundancyInfo(ecName string, packetSize int64) TypedRedundancyInfo {
	return TypedRedundancyInfo{
		Type: RedundancyRep,
		Info: ECRedundancyInfo{
			ECName:     ecName,
			PacketSize: packetSize,
		},
	}
}

func (i *TypedRedundancyInfo) IsRepInfo() bool {
	return i.Type == RedundancyRep
}

func (i *TypedRedundancyInfo) IsECInfo() bool {
	return i.Type == RedundancyEC
}

func (i *TypedRedundancyInfo) ToRepInfo() (RepRedundancyInfo, error) {
	var info RepRedundancyInfo
	err := serder.AnyToAny(i.Info, &info)
	return info, err
}

func (i *TypedRedundancyInfo) ToECInfo() (ECRedundancyInfo, error) {
	var info ECRedundancyInfo
	err := serder.AnyToAny(i.Info, &info)
	return info, err
}

func (i *TypedRedundancyInfo) Scan(src interface{}) error {
	data, ok := src.([]uint8)
	if !ok {
		return fmt.Errorf("unknow src type: %v", myreflect.TypeOfValue(data))
	}

	return serder.JSONToObject(data, i)
}

type NodePackageCachingInfo struct {
	NodeID      int64 `json:"nodeID"`
	FileSize    int64 `json:"fileSize"`
	ObjectCount int64 `json:"objectCount"`
}

type PackageCachingInfo struct {
	NodeInfos     []NodePackageCachingInfo `json:"nodeInfos"`
	PackageSize   int64                    `json:"packageSize"`
	RedunancyType string                   `json:"redunancyType"`
}

func NewPackageCachingInfo(nodeInfos []NodePackageCachingInfo, packageSize int64, redunancyType string) PackageCachingInfo {
	return PackageCachingInfo{
		NodeInfos:     nodeInfos,
		PackageSize:   packageSize,
		RedunancyType: redunancyType,
	}
}
