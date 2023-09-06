package scheduler

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"gitlink.org.cn/cloudream/common/models"
)

func Test_JobSet(t *testing.T) {
	Convey("提交任务集和设置LocalFile", t, func() {
		cli := NewClient(&Config{
			URL: "http://localhost:7891",
		})

		id, err := cli.JobSetSumbit(JobSetSumbitReq{
			JobSetInfo: models.JobSetInfo{
				Jobs: []models.JobInfo{
					models.ResourceJobInfo{
						Type: models.JobTypeResource,
					},
					models.NormalJobInfo{
						Type: models.JobTypeNormal,
						Files: models.JobFilesInfo{
							Dataset: models.PackageJobFileInfo{
								Type: models.FileInfoTypePackage,
							},
							Code: models.LocalJobFileInfo{
								Type:      models.FileInfoTypeLocalFile,
								LocalPath: "code",
							},
							Image: models.ImageJobFileInfo{
								Type: models.FileInfoTypeImage,
							},
						},
					},
				},
			},
		})
		So(err, ShouldBeNil)
		So(id.JobSetID, ShouldNotBeEmpty)

		err = cli.JobSetSetLocalFile(JobSetSetLocalFileReq{
			JobSetID:  id.JobSetID,
			LocalPath: "code",
			PackageID: 1,
		})
		So(err, ShouldBeNil)
	})
}
