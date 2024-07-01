package time2

import (
	"fmt"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_Duration(t *testing.T) {
	Convey("从字符串解析", t, func() {
		dur := Duration{}
		_, err := fmt.Sscanf("10s", "%v", &dur)
		So(err, ShouldEqual, nil)
		So(dur.Std(), ShouldEqual, 10*time.Second)
	})

	Convey("包含空白字符", t, func() {
		dur := Duration{}
		_, err := fmt.Sscanf(" 10s\t\n\r", "%v", &dur)
		So(err, ShouldEqual, nil)
		So(dur.Std(), ShouldEqual, 10*time.Second)
	})
}
