package stgsdk

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/google/uuid"
	. "github.com/smartystreets/goconvey/convey"
	"gitlink.org.cn/cloudream/common/pkgs/iterator"
)

func Test_PackageGet(t *testing.T) {
	Convey("上传后获取Package信息", t, func() {
		cli := NewClient(&Config{
			URL: "http://localhost:7890",
		})

		fileData := make([]byte, 4096)
		for i := 0; i < len(fileData); i++ {
			fileData[i] = byte(i)
		}

		pkgName := uuid.NewString()
		upResp, err := cli.PackageUpload(PackageUploadReq{
			UserID:   0,
			BucketID: 1,
			Name:     pkgName,
			Redundancy: TypedRedundancyInfo{
				Type: RedundancyRep,
				Info: NewRepRedundancyInfo(1),
			},
			Files: iterator.Array(
				&IterPackageUploadFile{
					Path: "test",
					File: io.NopCloser(bytes.NewBuffer(fileData)),
				},
				&IterPackageUploadFile{
					Path: "test2",
					File: io.NopCloser(bytes.NewBuffer(fileData)),
				},
			),
		})
		So(err, ShouldBeNil)

		getResp, err := cli.PackageGet(PackageGetReq{
			UserID:    0,
			PackageID: upResp.PackageID,
		})
		So(err, ShouldBeNil)

		So(getResp.PackageID, ShouldEqual, upResp.PackageID)
		So(getResp.Package.Name, ShouldEqual, pkgName)

		err = cli.PackageDelete(PackageDeleteReq{
			UserID:    0,
			PackageID: upResp.PackageID,
		})
		So(err, ShouldBeNil)
	})
}

func Test_Object(t *testing.T) {
	Convey("上传，下载，删除", t, func() {
		cli := NewClient(&Config{
			URL: "http://localhost:7890",
		})

		fileData := make([]byte, 4096)
		for i := 0; i < len(fileData); i++ {
			fileData[i] = byte(i)
		}

		nodeAff := int64(2)
		upResp, err := cli.PackageUpload(PackageUploadReq{
			UserID:   0,
			BucketID: 1,
			Name:     uuid.NewString(),
			Redundancy: TypedRedundancyInfo{
				Type: RedundancyRep,
				Info: NewRepRedundancyInfo(1),
			},
			NodeAffinity: &nodeAff,
			Files: iterator.Array(
				&IterPackageUploadFile{
					Path: "test",
					File: io.NopCloser(bytes.NewBuffer(fileData)),
				},
				&IterPackageUploadFile{
					Path: "test2",
					File: io.NopCloser(bytes.NewBuffer(fileData)),
				},
			),
		})
		So(err, ShouldBeNil)

		// downFs, err := cli.ObjectDownload(ObjectDownloadReq{
		// 	UserID:   0,
		// 	ObjectID: upResp.ObjectID,
		// })
		// So(err, ShouldBeNil)
		//
		// downFileData, err := io.ReadAll(downFs)
		// So(err, ShouldBeNil)
		// So(downFileData, ShouldResemble, fileData)
		// downFs.Close()

		err = cli.PackageDelete(PackageDeleteReq{
			UserID:    0,
			PackageID: upResp.PackageID,
		})
		So(err, ShouldBeNil)
	})
}

func Test_Storage(t *testing.T) {
	Convey("上传后调度文件", t, func() {
		cli := NewClient(&Config{
			URL: "http://localhost:7890",
		})

		fileData := make([]byte, 4096)
		for i := 0; i < len(fileData); i++ {
			fileData[i] = byte(i)
		}

		upResp, err := cli.PackageUpload(PackageUploadReq{
			UserID:   0,
			BucketID: 1,
			Name:     uuid.NewString(),
			Redundancy: TypedRedundancyInfo{
				Type: RedundancyRep,
				Info: NewRepRedundancyInfo(1),
			},
			Files: iterator.Array(
				&IterPackageUploadFile{
					Path: "test",
					File: io.NopCloser(bytes.NewBuffer(fileData)),
				},
				&IterPackageUploadFile{
					Path: "test2",
					File: io.NopCloser(bytes.NewBuffer(fileData)),
				},
			),
		})
		So(err, ShouldBeNil)

		_, err = cli.StorageLoadPackage(StorageLoadPackageReq{
			UserID:    0,
			PackageID: upResp.PackageID,
			StorageID: 1,
		})
		So(err, ShouldBeNil)

		err = cli.PackageDelete(PackageDeleteReq{
			UserID:    0,
			PackageID: upResp.PackageID,
		})
		So(err, ShouldBeNil)
	})
}

func Test_Cache(t *testing.T) {
	Convey("上传后移动文件", t, func() {
		cli := NewClient(&Config{
			URL: "http://localhost:7890",
		})

		fileData := make([]byte, 4096)
		for i := 0; i < len(fileData); i++ {
			fileData[i] = byte(i)
		}

		upResp, err := cli.PackageUpload(PackageUploadReq{
			UserID:   0,
			BucketID: 1,
			Name:     uuid.NewString(),
			Redundancy: TypedRedundancyInfo{
				Type: RedundancyRep,
				Info: NewRepRedundancyInfo(1),
			},
			Files: iterator.Array(
				&IterPackageUploadFile{
					Path: "test.txt",
					File: io.NopCloser(bytes.NewBuffer(fileData)),
				},
				&IterPackageUploadFile{
					Path: "test2.txt",
					File: io.NopCloser(bytes.NewBuffer(fileData)),
				},
			),
		})
		So(err, ShouldBeNil)

		cacheMoveResp, err := cli.CacheMovePackage(CacheMovePackageReq{
			UserID:    0,
			PackageID: upResp.PackageID,
			NodeID:    1,
		})
		So(err, ShouldBeNil)

		cacheInfoResp, err := cli.GetPackageObjectCacheInfos(GetPackageObjectCacheInfosReq{
			UserID:    0,
			PackageID: upResp.PackageID,
		})
		So(err, ShouldBeNil)

		So(cacheInfoResp.Infos, ShouldResemble, cacheMoveResp.CacheInfos)

		err = cli.PackageDelete(PackageDeleteReq{
			UserID:    0,
			PackageID: upResp.PackageID,
		})
		So(err, ShouldBeNil)
	})
}

func Test_GetNodeInfos(t *testing.T) {
	Convey("测试获取node信息", t, func() {
		cli := NewClient(&Config{
			URL: "http://localhost:7890",
		})
		resp1, err := cli.PackageGetCachedNodes(PackageGetCachedNodesReq{
			PackageID: 11,
			UserID:    0,
		})
		So(err, ShouldBeNil)
		fmt.Printf("resp1: %v\n", resp1)

		resp2, err := cli.PackageGetLoadedNodes(PackageGetLoadedNodesReq{
			PackageID: 11,
			UserID:    0,
		})
		So(err, ShouldBeNil)
		fmt.Printf("resp2: %v\n", resp2)
	})
}
