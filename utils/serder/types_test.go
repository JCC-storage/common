package serder

import (
	"encoding/json"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_Timestamp(t *testing.T) {
	Convey("秒级时间戳", t, func() {
		str := "1698894747"

		var ts TimestampSecond

		err := json.Unmarshal([]byte(str), &ts)
		So(err, ShouldBeNil)

		t := time.Time(ts)
		So(t.Unix(), ShouldEqual, 1698894747)
	})

	Convey("毫秒级时间戳", t, func() {
		str := "1698895130651"

		var ts TimestampMilliSecond

		err := json.Unmarshal([]byte(str), &ts)
		So(err, ShouldBeNil)

		t := time.Time(ts)
		So(t.UnixMilli(), ShouldEqual, 1698895130651)
	})
}
