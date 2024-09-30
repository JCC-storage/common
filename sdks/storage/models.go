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
	(*LRCRedundancy)(nil),
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

func (b *ECRedundancy) StripSize() int64 {
	return int64(b.ChunkSize) * int64(b.K)
}

var DefaultLRCRedundancy = *NewLRCRedundancy(2, 4, []int{2}, 1024*1024*5)

type LRCRedundancy struct {
	serder.Metadata `union:"lrc"`
	Type            string `json:"type"`
	K               int    `json:"k"`
	N               int    `json:"n"`
	Groups          []int  `json:"groups"`
	ChunkSize       int    `json:"chunkSize"`
}

func NewLRCRedundancy(k int, n int, groups []int, chunkSize int) *LRCRedundancy {
	return &LRCRedundancy{
		Type:      "lrc",
		K:         k,
		N:         n,
		Groups:    groups,
		ChunkSize: chunkSize,
	}
}
func (b *LRCRedundancy) Value() (driver.Value, error) {
	return serder.ObjectToJSONEx[Redundancy](b)
}

// 判断指定块属于哪个组。如果都不属于，则返回-1。
func (b *LRCRedundancy) FindGroup(idx int) int {
	if idx >= b.N-len(b.Groups) {
		return idx - (b.N - len(b.Groups))
	}

	for i, group := range b.Groups {
		if idx < group {
			return i
		}
		idx -= group
	}

	return -1
}

// M = N - len(Groups)，即数据块+校验块的总数，不包括组校验块。
func (b *LRCRedundancy) M() int {
	return b.N - len(b.Groups)
}

func (b *LRCRedundancy) GetGroupElements(grp int) []int {
	var idxes []int

	grpStart := 0
	for i := 0; i < grp; i++ {
		grpStart += b.Groups[i]
	}

	for i := 0; i < b.Groups[grp]; i++ {
		idxes = append(idxes, grpStart+i)
	}

	idxes = append(idxes, b.N-len(b.Groups)+grp)
	return idxes
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

func (n Node) String() string {
	return fmt.Sprintf("%v(%v)", n.Name, n.NodeID)
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

type Storage struct {
	StorageID  StorageID `db:"StorageID" json:"storageID"`
	Name       string    `db:"Name" json:"name"`
	NodeID     NodeID    `db:"NodeID" json:"nodeID"`
	LocalBase  string    `db:"LocalBase" json:"localBase"`   // 存储服务挂载在代理节点的目录
	RemoteBase string    `db:"RemoteBase" json:"remoteBase"` // 挂载在本地的目录对应存储服务的哪个路径
	State      string    `db:"State" json:"state"`
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
