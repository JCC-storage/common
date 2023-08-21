package unifyops

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_Unify_Ops(t *testing.T) {
	Convey("测试获取SlwNode信息", t, func() {
		cli := NewClient("http://101.201.215.165:6000")

		slwNodeInfo, err := cli.getSlwNodeInfo()
		So(err, ShouldBeNil)

		sNodes := *slwNodeInfo
		cpuData, err := cli.getCPUData(Node{
			NodeId: sNodes[0].ID,
		})
		So(err, ShouldBeNil)
		fmt.Printf("cpuData: %v\n", cpuData)

		gpuData, err := cli.getGPUData(Node{
			NodeId: sNodes[0].ID,
		})
		So(err, ShouldBeNil)
		fmt.Printf("gpuData: %v\n", gpuData)

		npuData, err := cli.getNPUData(Node{
			NodeId: sNodes[0].ID,
		})
		So(err, ShouldBeNil)
		fmt.Printf("npuData: %v\n", npuData)

		mluData, err := cli.getMLUData(Node{
			NodeId: sNodes[0].ID,
		})
		So(err, ShouldBeNil)
		fmt.Printf("mluData: %v\n", mluData)

		storageData, err := cli.getStorageData(Node{
			NodeId: sNodes[0].ID,
		})
		So(err, ShouldBeNil)
		fmt.Printf("storageData: %v\n", storageData)

		memoryData, err := cli.getMemoryData(Node{
			NodeId: sNodes[0].ID,
		})
		So(err, ShouldBeNil)
		fmt.Printf("memoryData: %v\n", memoryData)

		indicatorData, err := cli.getIndicatorData(Node{
			NodeId: sNodes[0].ID,
		})
		So(err, ShouldBeNil)
		fmt.Printf("indicatorData: %v\n", indicatorData)

	})
}
