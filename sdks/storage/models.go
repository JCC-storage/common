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

// 文件的SHA256哈希值，全大写的16进制字符串格式
type FileHash string

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
	PackageID PackageID `gorm:"column:PackageID; primaryKey; type:bigint; autoIncrement" json:"packageID"`
	Name      string    `gorm:"column:Name; type:varchar(255); not null" json:"name"`
	BucketID  BucketID  `gorm:"column:BucketID; type:bigint; not null" json:"bucketID"`
	State     string    `gorm:"column:State; type:varchar(255); not null" json:"state"`
}

func (Package) TableName() string {
	return "Package"
}

type Object struct {
	ObjectID   ObjectID   `json:"objectID"  gorm:"column:ObjectID; primaryKey; type:bigint; autoIncrement" `
	PackageID  PackageID  `json:"packageID" gorm:"column:PackageID; type:bigint; not null"`
	Path       string     `json:"path" gorm:"column:Path; type:varchar(1024); not null"`
	Size       int64      `json:"size,string" gorm:"column:Size; type:bigint; not null"`
	FileHash   FileHash   `json:"fileHash" gorm:"column:FileHash; type:char(64); not null"`
	Redundancy Redundancy `json:"redundancy" gorm:"column:Redundancy; type: json; serializer:union"`
	CreateTime time.Time  `json:"createTime" gorm:"column:CreateTime; type:datetime; not null"`
	UpdateTime time.Time  `json:"updateTime" gorm:"column:UpdateTime; type:datetime; not null"`
}

func (Object) TableName() string {
	return "Object"
}

type Node struct {
	NodeID         NodeID          `gorm:"column:NodeID; primaryKey; type:bigint; autoIncrement" json:"nodeID"`
	Name           string          `gorm:"column:Name; type:varchar(255); not null" json:"name"`
	Address        NodeAddressInfo `gorm:"column:Address; type:json; serializer:union" json:"address"`
	LocationID     LocationID      `gorm:"column:LocationID; type:bigint; not null" json:"locationID"`
	State          string          `gorm:"column:State; type:varchar(255); not null" json:"state"`
	LastReportTime *time.Time      `gorm:"column:LastReportTime; type:datetime" json:"lastReportTime"`
}

func (Node) TableName() string {
	return "Node"
}

type NodeAddressInfo interface {
}

var NodeAddressUnion = types.NewTypeUnion[NodeAddressInfo](
	(*GRPCAddressInfo)(nil),
	(*HttpAddressInfo)(nil),
)

var _ = serder.UseTypeUnionInternallyTagged(&NodeAddressUnion, "type")

type GRPCAddressInfo struct {
	serder.Metadata  `union:"GRPC"`
	Type             string `json:"type"`
	LocalIP          string `json:"localIP"`
	ExternalIP       string `json:"externalIP"`
	LocalGRPCPort    int    `json:"localGRPCPort"`
	ExternalGRPCPort int    `json:"externalGRPCPort"`
}

type HttpAddressInfo struct {
	serder.Metadata `union:"HTTP"`
	Type            string `json:"type"`
	LocalIP         string `json:"localIP"`
	ExternalIP      string `json:"externalIP"`
	Port            int    `json:"port"`
}

func (n Node) String() string {
	return fmt.Sprintf("%v(%v)", n.Name, n.NodeID)
}

type PinnedObject struct {
	ObjectID   ObjectID  `gorm:"column:ObjectID; primaryKey; type:bigint" json:"objectID"`
	StorageID  StorageID `gorm:"column:StorageID; primaryKey; type:bigint" json:"storageID"`
	CreateTime time.Time `gorm:"column:CreateTime; type:datetime; not null" json:"createTime"`
}

func (PinnedObject) TableName() string {
	return "PinnedObject"
}

type Bucket struct {
	BucketID  BucketID `gorm:"column:BucketID; primaryKey; type:bigint; autoIncrement" json:"bucketID"`
	Name      string   `gorm:"column:Name; type:varchar(255); not null" json:"name"`
	CreatorID UserID   `gorm:"column:CreatorID; type:bigint; not null" json:"creatorID"`
}

func (Bucket) TableName() string {
	return "Bucket"
}

type NodeConnectivity struct {
	FromNodeID NodeID    `gorm:"column:FromNodeID; primaryKey; type:bigint" json:"fromNodeID"`
	ToNodeID   NodeID    `gorm:"column:ToNodeID; primaryKey; type:bigint" json:"ToNodeID"`
	Delay      *float32  `gorm:"column:Delay; type:float" json:"delay"`
	TestTime   time.Time `gorm:"column:TestTime; type:datetime" json:"testTime"`
}

func (NodeConnectivity) TableName() string {
	return "NodeConnectivity"
}

type StoragePackageCachingInfo struct {
	StorageID   StorageID `json:"storageID"`
	FileSize    int64     `json:"fileSize"`
	ObjectCount int64     `json:"objectCount"`
}

type PackageCachingInfo struct {
	StorageInfos []StoragePackageCachingInfo `json:"stgInfos"`
	PackageSize  int64                       `json:"packageSize"`
}

func NewPackageCachingInfo(stgInfos []StoragePackageCachingInfo, packageSize int64) PackageCachingInfo {
	return PackageCachingInfo{
		StorageInfos: stgInfos,
		PackageSize:  packageSize,
	}
}

type CodeError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *CodeError) Error() string {
	return fmt.Sprintf("code: %s, message: %s", e.Code, e.Message)
}
