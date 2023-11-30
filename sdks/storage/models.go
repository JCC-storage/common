package cdssdk

import (
	"database/sql/driver"
	"fmt"

	"gitlink.org.cn/cloudream/common/pkgs/types"
	myreflect "gitlink.org.cn/cloudream/common/utils/reflect"
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

type RedundancyBase struct{}

func (b *RedundancyBase) Value() (driver.Value, error) {
	return serder.ObjectToJSONEx[Redundancy](b)
}

var RedundancyUnion = serder.UseTypeUnionInternallyTagged(types.Ref(types.NewTypeUnion[Redundancy](
	(*RepRedundancy)(nil),
	(*ECRedundancy)(nil),
)), "type")

type RepRedundancy struct {
	RedundancyBase
	serder.Metadata `union:"rep"`
	Type            string `json:"type"`
	RepCount        int    `json:"repCount"`
}

func NewRepRedundancy(repCount int) RepRedundancy {
	return RepRedundancy{
		Type:     "rep",
		RepCount: repCount,
	}
}

type ECRedundancy struct {
	RedundancyBase
	serder.Metadata `union:"ec"`
	Type            string `json:"type"`
	K               int    `json:"k"`
	N               int    `json:"n"`
	ChunkSize       int    `json:"chunkSize"`
}

func NewECRedundancy(k int, n int, chunkSize int) ECRedundancy {
	return ECRedundancy{
		Type:      "ec",
		K:         k,
		N:         n,
		ChunkSize: chunkSize,
	}
}

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
	Redundancy Redundancy `db:"Redundancy" json:"redundancy"`
}

func (i *Object) Scan(src interface{}) error {
	data, ok := src.([]uint8)
	if !ok {
		return fmt.Errorf("unknow src type: %v", myreflect.TypeOfValue(data))
	}

	obj, err := serder.JSONToObjectEx[*Object](data)
	if err != nil {
		return err
	}

	*i = *obj
	return nil
}

type ObjectCacheInfo struct {
	Object   Object `json:"object"`
	FileHash string `json:"fileHash"`
}

func NewObjectCacheInfo(object Object, fileHash string) ObjectCacheInfo {
	return ObjectCacheInfo{
		Object:   object,
		FileHash: fileHash,
	}
}

type NodePackageCachingInfo struct {
	NodeID      NodeID `json:"nodeID"`
	FileSize    int64  `json:"fileSize"`
	ObjectCount int64  `json:"objectCount"`
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

type CodeError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *CodeError) Error() string {
	return fmt.Sprintf("code: %s, message: %s", e.Code, e.Message)
}
