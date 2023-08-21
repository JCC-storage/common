package storage

import (
	"bytes"
	"io"
	"testing"

	"github.com/google/uuid"
	. "github.com/smartystreets/goconvey/convey"
	"gitlink.org.cn/cloudream/common/models"
)

func Test_Object(t *testing.T) {
	Convey("上传，下载，删除", t, func() {
		cli := NewClient("http://localhost:7890")

		fileData := make([]byte, 4096)
		for i := 0; i < len(fileData); i++ {
			fileData[i] = byte(i)
		}

		upResp, err := cli.ObjectUpload(ObjectUploadReq{
			UserID:     0,
			BucketID:   1,
			FileSize:   4096,
			ObjectName: uuid.NewString(),
			Redundancy: models.TypedRedundancyInfo{
				Type: models.RedundancyRep,
				Info: models.NewRepRedundancyInfo(1),
			},
			File: bytes.NewBuffer(fileData),
		})
		So(err, ShouldBeNil)

		downFs, err := cli.ObjectDownload(ObjectDownloadReq{
			UserID:   0,
			ObjectID: upResp.ObjectID,
		})
		So(err, ShouldBeNil)

		downFileData, err := io.ReadAll(downFs)
		So(err, ShouldBeNil)
		So(downFileData, ShouldResemble, fileData)
		downFs.Close()

		err = cli.ObjectDelete(ObjectDeleteReq{
			UserID:   0,
			ObjectID: upResp.ObjectID,
		})
		So(err, ShouldBeNil)
	})
}

func Test_Storage(t *testing.T) {
	Convey("上传后调度文件", t, func() {
		cli := NewClient("http://localhost:7890")

		fileData := make([]byte, 4096)
		for i := 0; i < len(fileData); i++ {
			fileData[i] = byte(i)
		}

		upResp, err := cli.ObjectUpload(ObjectUploadReq{
			UserID:     0,
			BucketID:   1,
			FileSize:   4096,
			ObjectName: uuid.NewString(),
			Redundancy: models.TypedRedundancyInfo{
				Type: models.RedundancyRep,
				Info: models.NewRepRedundancyInfo(1),
			},
			File: bytes.NewBuffer(fileData),
		})
		So(err, ShouldBeNil)

		err = cli.StorageMoveObject(StorageMoveObjectReq{
			UserID:    0,
			ObjectID:  upResp.ObjectID,
			StorageID: 1,
		})
		So(err, ShouldBeNil)

		err = cli.ObjectDelete(ObjectDeleteReq{
			UserID:   0,
			ObjectID: upResp.ObjectID,
		})
		So(err, ShouldBeNil)
	})
}
