package cdssdk

import (
	"database/sql/driver"
	"fmt"
	"time"

	"gitlink.org.cn/cloudream/common/pkgs/types"
	"gitlink.org.cn/cloudream/common/utils/serder"
)

const (
	ObjectPathSeparator = "/"
)

type NodeID int64

type PackageID int64

type ObjectID int64

type UserID int64

type BucketID int64

type StorageID int64

type LocationID int64

/// TODO 将分散在各处的公共结构体定义集中到这里来

type Redundancy interface {
	driver.Valuer
}

var RedundancyUnion = serder.UseTypeUnionInternallyTagged(types.Ref(types.NewTypeUnion[Redundancy](
	(*NoneRedundancy)(nil),
	(*RepRedundancy)(nil),
	(*ECRedundancy)(nil),
)), "type")

type NoneRedundancy struct {
	serder.Metadata `union:"none"`
	Type            string `json:"type"`
}

func NewNoneRedundancy() *NoneRedundancy {
	return &NoneRedundancy{
		Type: "none",
	}
}
func (b *NoneRedundancy) Value() (driver.Value, error) {
	return serder.ObjectToJSONEx[Redundancy](b)
}

var DefaultRepRedundancy = *NewRepRedundancy(2)

type RepRedundancy struct {
	serder.Metadata `union:"rep"`
	Type            string `json:"type"`
	RepCount        int    `json:"repCount"`
}

func NewRepRedundancy(repCount int) *RepRedundancy {
	return &RepRedundancy{
		Type:     "rep",
		RepCount: repCount,
	}
}
func (b *RepRedundancy) Value() (driver.Value, error) {
	return serder.ObjectToJSONEx[Redundancy](b)
}

var DefaultECRedundancy = *NewECRedundancy(2, 3, 1024*1024*5)

type ECRedundancy struct {
	serder.Metadata `union:"ec"`
	Type            string `json:"type"`
	K               int    `json:"k"`
	N               int    `json:"n"`
	ChunkSize       int    `json:"chunkSize"`
}

func NewECRedundancy(k int, n int, chunkSize int) *ECRedundancy {
	return &ECRedundancy{
		Type:      "ec",
		K:         k,
		N:         n,
		ChunkSize: chunkSize,
	}
}
func (b *ECRedundancy) Value() (driver.Value, error) {
	return serder.ObjectToJSONEx[Redundancy](b)
}

const (
	PackageStateNormal  = "Normal"
	PackageStateDeleted = "Deleted"
)

type Package struct {
	PackageID PackageID `db:"PackageID" json:"packageID"`
	Name      string    `db:"Name" json:"name"`
	BucketID  BucketID  `db:"BucketID" json:"bucketID"`
	State     string    `db:"State" json:"state"`
}

type Object struct {
	ObjectID   ObjectID   `db:"ObjectID" json:"objectID"`
	PackageID  PackageID  `db:"PackageID" json:"packageID"`
	Path       string     `db:"Path" json:"path"`
	Size       int64      `db:"Size" json:"size,string"`
	FileHash   string     `db:"FileHash" json:"fileHash"`
	Redundancy Redundancy `db:"Redundancy" json:"redundancy"`
	CreateTime time.Time  `db:"CreateTime" json:"createTime"`
	UpdateTime time.Time  `db:"UpdateTime" json:"updateTime"`
}

type Node struct {
	NodeID           NodeID     `db:"NodeID" json:"nodeID"`
	Name             string     `db:"Name" json:"name"`
	LocalIP          string     `db:"LocalIP" json:"localIP"`
	ExternalIP       string     `db:"ExternalIP" json:"externalIP"`
	LocalGRPCPort    int        `db:"LocalGRPCPort" json:"localGRPCPort"`
	ExternalGRPCPort int        `db:"ExternalGRPCPort" json:"externalGRPCPort"`
	LocationID       LocationID `db:"LocationID" json:"locationID"`
	State            string     `db:"State" json:"state"`
	LastReportTime   *time.Time `db:"LastReportTime" json:"lastReportTime"`
}

type PinnedObject struct {
	ObjectID   ObjectID  `db:"ObjectID" json:"objectID"`
	NodeID     NodeID    `db:"NodeID" json:"nodeID"`
	CreateTime time.Time `db:"CreateTime" json:"createTime"`
}

type Bucket struct {
	BucketID  BucketID `db:"BucketID" json:"bucketID"`
	Name      string   `db:"Name" json:"name"`
	CreatorID UserID   `db:"CreatorID" json:"creatorID"`
}

type NodeConnectivity struct {
	FromNodeID NodeID    `db:"FromNodeID" json:"fromNodeID"`
	ToNodeID   NodeID    `db:"ToNodeID" json:"ToNodeID"`
	Delay      *float32  `db:"Delay" json:"delay"`
	TestTime   time.Time `db:"TestTime" json:"testTime"`
}

type NodePackageCachingInfo struct {
	NodeID      NodeID `json:"nodeID"`
	FileSize    int64  `json:"fileSize"`
	ObjectCount int64  `json:"objectCount"`
}

type PackageCachingInfo struct {
	NodeInfos   []NodePackageCachingInfo `json:"nodeInfos"`
	PackageSize int64                    `json:"packageSize"`
}

func NewPackageCachingInfo(nodeInfos []NodePackageCachingInfo, packageSize int64) PackageCachingInfo {
	return PackageCachingInfo{
		NodeInfos:   nodeInfos,
		PackageSize: packageSize,
	}
}

type CodeError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *CodeError) Error() string {
	return fmt.Sprintf("code: %s, message: %s", e.Code, e.Message)
}
