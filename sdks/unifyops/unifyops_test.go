package uopsdk

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_UnifyOps(t *testing.T) {
	Convey("测试获取SlwNode信息", t, func() {
		cli := NewClient(&Config{
			URL: "http://101.201.215.165:6000",
		})

		slwNodeInfos, err := cli.GetAllSlwNodeInfo()
		So(err, ShouldBeNil)

		cpuData, err := cli.GetCPUData(GetOneResourceDataReq{
			SlwNodeID: slwNodeInfos.Nodes[0].ID,
		})
		So(err, ShouldBeNil)
		fmt.Printf("cpuData: %v\n", cpuData)

		gpuData, err := cli.GetGPUData(GetOneResourceDataReq{
			SlwNodeID: slwNodeInfos.Nodes[0].ID,
		})
		So(err, ShouldBeNil)
		fmt.Printf("gpuData: %v\n", gpuData)

		npuData, err := cli.GetNPUData(GetOneResourceDataReq{
			SlwNodeID: slwNodeInfos.Nodes[0].ID,
		})
		So(err, ShouldBeNil)
		fmt.Printf("npuData: %v\n", npuData)

		mluData, err := cli.GetMLUData(GetOneResourceDataReq{
			SlwNodeID: slwNodeInfos.Nodes[0].ID,
		})
		So(err, ShouldBeNil)
		fmt.Printf("mluData: %v\n", mluData)

		storageData, err := cli.GetStorageData(GetOneResourceDataReq{
			SlwNodeID: slwNodeInfos.Nodes[0].ID,
		})
		So(err, ShouldBeNil)
		fmt.Printf("storageData: %v\n", storageData)

		memoryData, err := cli.GetMemoryData(GetOneResourceDataReq{
			SlwNodeID: slwNodeInfos.Nodes[0].ID,
		})
		So(err, ShouldBeNil)
		fmt.Printf("memoryData: %v\n", memoryData)

		indicatorData, err := cli.GetIndicatorData(GetOneResourceDataReq{
			SlwNodeID: slwNodeInfos.Nodes[0].ID,
		})
		So(err, ShouldBeNil)
		fmt.Printf("indicatorData: %v\n", indicatorData)

	})
}
