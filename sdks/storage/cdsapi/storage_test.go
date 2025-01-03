package cdsapi

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/google/uuid"
	. "github.com/smartystreets/goconvey/convey"
	"gitlink.org.cn/cloudream/common/pkgs/iterator"
	cdssdk "gitlink.org.cn/cloudream/common/sdks/storage"
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
		createResp, err := cli.Package().Create(PackageCreate{
			UserID:   1,
			BucketID: 1,
			Name:     pkgName,
		})
		So(err, ShouldBeNil)

		_, err = cli.Object().Upload(ObjectUpload{
			ObjectUploadInfo: ObjectUploadInfo{
				UserID:    1,
				PackageID: createResp.Package.PackageID,
			},
			Files: iterator.Array(
				&UploadingObject{
					Path: "abc/test",
					File: io.NopCloser(bytes.NewBuffer(fileData)),
				},
				&UploadingObject{
					Path: "test2",
					File: io.NopCloser(bytes.NewBuffer(fileData)),
				},
			),
		})
		So(err, ShouldBeNil)

		getResp, err := cli.Package().Get(PackageGetReq{
			UserID:    1,
			PackageID: createResp.Package.PackageID,
		})
		So(err, ShouldBeNil)

		So(getResp.PackageID, ShouldEqual, createResp.Package.PackageID)
		So(getResp.Package.Name, ShouldEqual, pkgName)

		err = cli.Package().Delete(PackageDelete{
			UserID:    1,
			PackageID: createResp.Package.PackageID,
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

		stgAff := cdssdk.StorageID(2)

		pkgName := uuid.NewString()
		createResp, err := cli.Package().Create(PackageCreate{
			UserID:   1,
			BucketID: 1,
			Name:     pkgName,
		})
		So(err, ShouldBeNil)

		_, err = cli.Object().Upload(ObjectUpload{
			ObjectUploadInfo: ObjectUploadInfo{
				UserID:    1,
				PackageID: createResp.Package.PackageID,
				Affinity:  stgAff,
			},
			Files: iterator.Array(
				&UploadingObject{
					Path: "test",
					File: io.NopCloser(bytes.NewBuffer(fileData)),
				},
				&UploadingObject{
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

		err = cli.Package().Delete(PackageDelete{
			UserID:    1,
			PackageID: createResp.Package.PackageID,
		})
		So(err, ShouldBeNil)
	})
}

func Test_ObjectList(t *testing.T) {
	Convey("路径查询", t, func() {
		cli := NewClient(&Config{
			URL: "http://localhost:7890",
		})

		resp, err := cli.Object().List(ObjectList{
			UserID:    1,
			PackageID: 10,
			Path:      "100x100K/zexema",
		})
		So(err, ShouldBeNil)
		fmt.Printf("\n")
		fmt.Printf("%+v\n", resp.Objects[0])
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
		createResp, err := cli.Package().Create(PackageCreate{
			UserID:   1,
			BucketID: 1,
			Name:     pkgName,
		})
		So(err, ShouldBeNil)

		_, err = cli.Object().Upload(ObjectUpload{
			ObjectUploadInfo: ObjectUploadInfo{
				UserID:    1,
				PackageID: createResp.Package.PackageID,
			},
			Files: iterator.Array(
				&UploadingObject{
					Path: "test",
					File: io.NopCloser(bytes.NewBuffer(fileData)),
				},
				&UploadingObject{
					Path: "test2",
					File: io.NopCloser(bytes.NewBuffer(fileData)),
				},
			),
		})
		So(err, ShouldBeNil)

		_, err = cli.StorageLoadPackage(StorageLoadPackageReq{
			UserID:    1,
			PackageID: createResp.Package.PackageID,
			StorageID: 1,
		})
		So(err, ShouldBeNil)

		err = cli.Package().Delete(PackageDelete{
			UserID:    1,
			PackageID: createResp.Package.PackageID,
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
		createResp, err := cli.Package().Create(PackageCreate{
			UserID:   1,
			BucketID: 1,
			Name:     pkgName,
		})
		So(err, ShouldBeNil)

		_, err = cli.Object().Upload(ObjectUpload{
			ObjectUploadInfo: ObjectUploadInfo{
				UserID:    1,
				PackageID: createResp.Package.PackageID,
			},
			Files: iterator.Array(
				&UploadingObject{
					Path: "test.txt",
					File: io.NopCloser(bytes.NewBuffer(fileData)),
				},
				&UploadingObject{
					Path: "test2.txt",
					File: io.NopCloser(bytes.NewBuffer(fileData)),
				},
			),
		})
		So(err, ShouldBeNil)

		_, err = cli.CacheMovePackage(CacheMovePackageReq{
			UserID:    1,
			PackageID: createResp.Package.PackageID,
			StorageID: 1,
		})
		So(err, ShouldBeNil)

		err = cli.Package().Delete(PackageDelete{
			UserID:    1,
			PackageID: createResp.Package.PackageID,
		})
		So(err, ShouldBeNil)
	})
}
