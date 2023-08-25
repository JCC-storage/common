package unifyops

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

		slwNodeInfo, err := cli.GetSlwNodeInfo()
		So(err, ShouldBeNil)

		sNodes := *slwNodeInfo
		cpuData, err := cli.GetCPUData(Node{
			NodeId: sNodes[0].ID,
		})
		So(err, ShouldBeNil)
		fmt.Printf("cpuData: %v\n", cpuData)

		gpuData, err := cli.GetGPUData(Node{
			NodeId: sNodes[0].ID,
		})
		So(err, ShouldBeNil)
		fmt.Printf("gpuData: %v\n", gpuData)

		npuData, err := cli.GetNPUData(Node{
			NodeId: sNodes[0].ID,
		})
		So(err, ShouldBeNil)
		fmt.Printf("npuData: %v\n", npuData)

		mluData, err := cli.GetMLUData(Node{
			NodeId: sNodes[0].ID,
		})
		So(err, ShouldBeNil)
		fmt.Printf("mluData: %v\n", mluData)

		storageData, err := cli.GetStorageData(Node{
			NodeId: sNodes[0].ID,
		})
		So(err, ShouldBeNil)
		fmt.Printf("storageData: %v\n", storageData)

		memoryData, err := cli.GetMemoryData(Node{
			NodeId: sNodes[0].ID,
		})
		So(err, ShouldBeNil)
		fmt.Printf("memoryData: %v\n", memoryData)

		indicatorData, err := cli.GetIndicatorData(Node{
			NodeId: sNodes[0].ID,
		})
		So(err, ShouldBeNil)
		fmt.Printf("indicatorData: %v\n", indicatorData)

	})
}