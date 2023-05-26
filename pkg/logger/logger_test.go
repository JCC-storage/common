package logger

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_FormatStruct(t *testing.T) {
	Convey("检查输出格式", t, func() {

		type Struct2 struct {
			Int int
		}

		type Struct struct {
			Arr       []int
			NilArr    []int
			FixedArr  [5]int
			St2       Struct2
			St2Ptr    *Struct2
			NilSt2Ptr *Struct2
		}

		So(fmt.Sprintf("%v", FormatStruct(Struct{
			Arr:      []int{1, 2, 3, 4},
			NilArr:   nil,
			FixedArr: [5]int{1, 2, 3, 4, 5},
			St2: Struct2{
				Int: 123,
			},
			St2Ptr: &Struct2{
				Int: 456,
			},
			NilSt2Ptr: nil,
		})), ShouldEqual, "len(Arr): 4, len(NilArr): 0, len(FixedArr): 5, St2: <Struct2>, St2Ptr: &<Struct2>, NilSt2Ptr: <nil>")
	})
}
