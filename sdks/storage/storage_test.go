package cdssdk

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
		createResp, err := cli.Package().Create(PackageCreateReq{
			UserID:   1,
			BucketID: 1,
			Name:     pkgName,
		})
		So(err, ShouldBeNil)

		_, err = cli.Object().Upload(ObjectUploadReq{
			ObjectUploadInfo: ObjectUploadInfo{
				UserID:    1,
				PackageID: createResp.PackageID,
			},
			Files: iterator.Array(
				&IterObjectUpload{
					Path: "test",
					File: io.NopCloser(bytes.NewBuffer(fileData)),
				},
				&IterObjectUpload{
					Path: "test2",
					File: io.NopCloser(bytes.NewBuffer(fileData)),
				},
			),
		})
		So(err, ShouldBeNil)

		getResp, err := cli.Package().Get(PackageGetReq{
			UserID:    1,
			PackageID: createResp.PackageID,
		})
		So(err, ShouldBeNil)

		So(getResp.PackageID, ShouldEqual, createResp.PackageID)
		So(getResp.Package.Name, ShouldEqual, pkgName)

		err = cli.Package().Delete(PackageDeleteReq{
			UserID:    1,
			PackageID: createResp.PackageID,
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

		nodeAff := NodeID(2)

		pkgName := uuid.NewString()
		createResp, err := cli.Package().Create(PackageCreateReq{
			UserID:   1,
			BucketID: 1,
			Name:     pkgName,
		})
		So(err, ShouldBeNil)

		_, err = cli.Object().Upload(ObjectUploadReq{
			ObjectUploadInfo: ObjectUploadInfo{
				UserID:       1,
				PackageID:    createResp.PackageID,
				NodeAffinity: &nodeAff,
			},
			Files: iterator.Array(
				&IterObjectUpload{
					Path: "test",
					File: io.NopCloser(bytes.NewBuffer(fileData)),
				},
				&IterObjectUpload{
					Path: "test2",
					File: io.NopCloser(bytes.NewBuffer(fileData)),
				},
			),
		})
		So(err, ShouldBeNil)

		// downFs, err := cli.ObjectDownload(ObjectDownloadReq{
		// 	UserID:   1,
		// 	ObjectID: upResp.ObjectID,
		// })
		// So(err, ShouldBeNil)
		//
		// downFileData, err := io.ReadAll(downFs)
		// So(err, ShouldBeNil)
		// So(downFileData, ShouldResemble, fileData)
		// downFs.Close()

		err = cli.Package().Delete(PackageDeleteReq{
			UserID:    1,
			PackageID: createResp.PackageID,
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

		pkgName := uuid.NewString()
		createResp, err := cli.Package().Create(PackageCreateReq{
			UserID:   1,
			BucketID: 1,
			Name:     pkgName,
		})
		So(err, ShouldBeNil)

		_, err = cli.Object().Upload(ObjectUploadReq{
			ObjectUploadInfo: ObjectUploadInfo{
				UserID:    1,
				PackageID: createResp.PackageID,
			},
			Files: iterator.Array(
				&IterObjectUpload{
					Path: "test",
					File: io.NopCloser(bytes.NewBuffer(fileData)),
				},
				&IterObjectUpload{
					Path: "test2",
					File: io.NopCloser(bytes.NewBuffer(fileData)),
				},
			),
		})
		So(err, ShouldBeNil)

		_, err = cli.StorageLoadPackage(StorageLoadPackageReq{
			UserID:    1,
			PackageID: createResp.PackageID,
			StorageID: 1,
		})
		So(err, ShouldBeNil)

		err = cli.Package().Delete(PackageDeleteReq{
			UserID:    1,
			PackageID: createResp.PackageID,
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

		pkgName := uuid.NewString()
		createResp, err := cli.Package().Create(PackageCreateReq{
			UserID:   1,
			BucketID: 1,
			Name:     pkgName,
		})
		So(err, ShouldBeNil)

		_, err = cli.Object().Upload(ObjectUploadReq{
			ObjectUploadInfo: ObjectUploadInfo{
				UserID:    1,
				PackageID: createResp.PackageID,
			},
			Files: iterator.Array(
				&IterObjectUpload{
					Path: "test.txt",
					File: io.NopCloser(bytes.NewBuffer(fileData)),
				},
				&IterObjectUpload{
					Path: "test2.txt",
					File: io.NopCloser(bytes.NewBuffer(fileData)),
				},
			),
		})
		So(err, ShouldBeNil)

		_, err = cli.CacheMovePackage(CacheMovePackageReq{
			UserID:    1,
			PackageID: createResp.PackageID,
			NodeID:    1,
		})
		So(err, ShouldBeNil)

		err = cli.Package().Delete(PackageDeleteReq{
			UserID:    1,
			PackageID: createResp.PackageID,
		})
		So(err, ShouldBeNil)
	})
}

func Test_GetNodeInfos(t *testing.T) {
	Convey("测试获取node信息", t, func() {
		cli := NewClient(&Config{
			URL: "http://localhost:7890",
		})
		resp1, err := cli.Package().GetCachedNodes(PackageGetCachedNodesReq{
			PackageID: 11,
			UserID:    1,
		})
		So(err, ShouldBeNil)
		fmt.Printf("resp1: %v\n", resp1)

		resp2, err := cli.Package().GetLoadedNodes(PackageGetLoadedNodesReq{
			PackageID: 11,
			UserID:    1,
		})
		So(err, ShouldBeNil)
		fmt.Printf("resp2: %v\n", resp2)
	})
}
