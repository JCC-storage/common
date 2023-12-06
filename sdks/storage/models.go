package cdssdk

import (
	"database/sql/driver"
	"fmt"

	"gitlink.org.cn/cloudream/common/pkgs/types"
	"gitlink.org.cn/cloudream/common/utils/serder"
)

const (
	ObjectPathSeperator = "/"
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
	(*RepRedundancy)(nil),
	(*ECRedundancy)(nil),
)), "type")

type RepRedundancy struct {
	serder.Metadata `union:"rep"`
	Type            string `json:"type"`
}

func NewRepRedundancy() *RepRedundancy {
	return &RepRedundancy{
		Type: "rep",
	}
}
func (b *RepRedundancy) Value() (driver.Value, error) {
	return serder.ObjectToJSONEx[Redundancy](b)
}

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
