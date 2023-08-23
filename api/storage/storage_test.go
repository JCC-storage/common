package storage

import (
	"bytes"
	"testing"

	"github.com/google/uuid"
	. "github.com/smartystreets/goconvey/convey"
	"gitlink.org.cn/cloudream/common/models"
)

func Test_Object(t *testing.T) {
	Convey("上传，下载，删除", t, func() {
		cli := NewClient(&Config{
			URL: "http://localhost:7890",
		})

		fileData := make([]byte, 4096)
		for i := 0; i < len(fileData); i++ {
			fileData[i] = byte(i)
		}

		_, err := cli.PackageUpload(PackageUploadReq{
			UserID:   0,
			BucketID: 1,
			Name:     uuid.NewString(),
			Redundancy: models.TypedRedundancyInfo{
				Type: models.RedundancyRep,
				Info: models.NewRepRedundancyInfo(1),
			},
			Files: []PackageUploadFile{
				{
					Path: "test",
					File: bytes.NewBuffer(fileData),
				},
				{
					Path: "test2",
					File: bytes.NewBuffer(fileData),
				},
			},
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

		//err = cli.PackageDelete(PackageDeleteReq{
		//	UserID:    0,
		//	PackageID: upResp.PackageID,
		//})
		//So(err, ShouldBeNil)
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
			Redundancy: models.TypedRedundancyInfo{
				Type: models.RedundancyRep,
				Info: models.NewRepRedundancyInfo(1),
			},
			Files: []PackageUploadFile{
				{
					Path: "test",
					File: bytes.NewBuffer(fileData),
				},
				{
					Path: "test2",
					File: bytes.NewBuffer(fileData),
				},
			},
		})
		So(err, ShouldBeNil)

		err = cli.StorageLoadPackage(StorageLoadPackageReq{
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
			Redundancy: models.TypedRedundancyInfo{
				Type: models.RedundancyRep,
				Info: models.NewRepRedundancyInfo(1),
			},
			Files: []PackageUploadFile{
				{
					Path: "test",
					File: bytes.NewBuffer(fileData),
				},
				{
					Path: "test3",
					File: bytes.NewBuffer(fileData),
				},
			},
		})
		So(err, ShouldBeNil)

		err = cli.CacheMovePackage(CacheMovePackageReq{
			UserID:    0,
			PackageID: upResp.PackageID,
			NodeID:    1,
		})
		So(err, ShouldBeNil)

		err = cli.PackageDelete(PackageDeleteReq{
			UserID:    0,
			PackageID: upResp.PackageID,
		})
		So(err, ShouldBeNil)
	})
}
