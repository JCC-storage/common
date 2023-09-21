package schsdk

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_JobSet(t *testing.T) {
	Convey("提交任务集和设置LocalFile", t, func() {
		cli := NewClient(&Config{
			URL: "http://localhost:7891",
		})

		id, err := cli.JobSetSumbit(JobSetSumbitReq{
			JobSetInfo: JobSetInfo{
				Jobs: []JobInfo{
					&ResourceJobInfo{
						Type: JobTypeResource,
					},
					&NormalJobInfo{
						Type: JobTypeNormal,
						Files: JobFilesInfo{
							Dataset: &PackageJobFileInfo{
								Type: FileInfoTypePackage,
							},
							Code: &LocalJobFileInfo{
								Type:      FileInfoTypeLocalFile,
								LocalPath: "code",
							},
							Image: &ImageJobFileInfo{
								Type: FileInfoTypeImage,
							},
						},
					},
				},
			},
		})
		So(err, ShouldBeNil)
		So(id.JobSetID, ShouldNotBeEmpty)

		err = cli.JobSetLocalFileUploaded(JobSetLocalFileUploadedReq{
			JobSetID:  id.JobSetID,
			LocalPath: "code",
			PackageID: 1,
		})
		So(err, ShouldBeNil)
	})
}
