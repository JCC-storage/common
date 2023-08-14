package logger

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_FormatStruct(t *testing.T) {
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
		Struct2
	}

	st := Struct{
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
		Struct2: Struct2{
			Int: 789,
		},
	}

	fmtedStr := "len(Arr): 4, NilArr: <nil>, len(FixedArr): 5, St2: <Struct2>, St2Ptr: &<Struct2>, NilSt2Ptr: <nil>, Int: 789"

	Convey("基本格式", t, func() {
		So(fmt.Sprintf("%v", FormatStruct(st)), ShouldEqual, fmtedStr)
	})

	Convey("指针", t, func() {
		So(fmt.Sprintf("%v", FormatStruct(&st)), ShouldEqual, fmtedStr)
	})

	Convey("interface", t, func() {
		var ift any = st
		So(fmt.Sprintf("%v", FormatStruct(ift)), ShouldEqual, fmtedStr)
	})
}
