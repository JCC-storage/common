package imsdk

import (
	"io"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_IPFSRead(t *testing.T) {
	Convey("读取IPFS文件", t, func() {
		cli := NewClient(&Config{
			URL: "http://localhost:7893",
		})

		file, err := cli.IPFSRead(IPFSRead{
			FileHash: "QmcYsRZxmYGgSaydEiJwJRMsD8uWzS2x8gCt1iGMtsZKsU",
			Length:   2,
		})
		So(err, ShouldBeNil)
		defer file.Close()

		data, err := io.ReadAll(file)
		So(err, ShouldBeNil)
		So(len(data), ShouldEqual, 2)
	})
}

func Test_Package(t *testing.T) {
	Convey("获取Package文件列表", t, func() {
		cli := NewClient(&Config{
			URL: "http://localhost:7893",
		})

		_, err := cli.PackageGetWithObjects(PackageGetWithObjectsInfos{UserID: 0, PackageID: 13})
		So(err, ShouldBeNil)
	})
}
