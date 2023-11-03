package pcmsdk

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	schsdk "gitlink.org.cn/cloudream/common/sdks/scheduler"
)

func Test_SubmitTask(t *testing.T) {
	Convey("提交任务，查询任务", t, func() {
		cli := NewClient(&Config{
			URL: "http://localhost:8889",
		})

		submitResp, err := cli.SubmitTask(SubmitTaskReq{
			PartID:     1711652475901054976,
			ImageID:    "1d1769857cd64c03928c8a1a4ee4a23f",
			ResourceID: "6388d3c27f654fa5b11439a3d6098dbc",
			CMD:        "echo $asd",
			Envs: []schsdk.KVPair{{
				Key:   "asd",
				Value: "hello",
			}},
			Params: []schsdk.KVPair{},
		})
		So(err, ShouldBeNil)

		t.Logf("taskID: %s", submitResp.TaskID)

		taskResp, err := cli.GetTask(GetTaskReq{
			PartID: 1711652475901054976,
			TaskID: submitResp.TaskID,
		})
		So(err, ShouldBeNil)

		<-time.After(time.Second * 3)

		t.Logf("taskName: %s, taskStatus: %s, startedAt: %v", taskResp.TaskName, taskResp.TaskStatus, taskResp.StartedAt)
	})

}

func Test_GetImageList(t *testing.T) {
	Convey("查询镜像列表", t, func() {
		cli := NewClient(&Config{
			URL: "http://localhost:8889",
		})

		getReps, err := cli.GetImageList(GetImageListReq{
			PartID: 1711652475901054976,
		})
		So(err, ShouldBeNil)

		t.Logf("imageList: %v", getReps.Images)
	})
}
