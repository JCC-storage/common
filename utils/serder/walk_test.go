package serder

import (
	"fmt"
	"reflect"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	myreflect "gitlink.org.cn/cloudream/common/utils/reflect"
)

func Test_WalkValue(t *testing.T) {
	type Base struct {
		Int    int
		Bool   bool
		String string
		Float  float32
	}
	type ArraryStruct struct {
		IntArr []int
		StArr  []Base
		ArrArr [][]int
		Nil    []Base
	}
	type MapStruct struct {
		StrMap map[string]string
		StMap  map[string]Base
		MapMap map[string]map[string]string
		Nil    map[string]Base
	}

	type Top struct {
		ArrSt  ArraryStruct
		MapSt  *MapStruct
		BaseIf any
		NilPtr *Base
	}

	isBaseDataType := func(val reflect.Value) bool {
		typ := val.Type()
		return typ == myreflect.TypeOf[int]() || typ == myreflect.TypeOf[bool]() ||
			typ == myreflect.TypeOf[string]() || typ == myreflect.TypeOf[float32]() || val.IsZero()
	}

	toString := func(val any) string {
		return fmt.Sprintf("%v", val)
	}

	Convey("遍历", t, func() {
		val := Top{
			ArrSt: ArraryStruct{
				IntArr: []int{1, 2, 3},
				StArr: []Base{
					{
						Int:    1,
						Bool:   true,
						String: "test",
						Float:  1,
					},
					{
						Int:    2,
						Bool:   false,
						String: "test2",
						Float:  2,
					},
				},
				ArrArr: [][]int{
					{1, 2, 3},
					{},
					nil,
				},
				Nil: nil,
			},
			MapSt: &MapStruct{
				StrMap: map[string]string{
					"a": "1",
					"b": "2",
				},
				StMap: map[string]Base{
					"a": {
						Int:    1,
						Bool:   true,
						String: "test",
						Float:  1,
					},
					"b": {
						Int:    2,
						Bool:   false,
						String: "test2",
						Float:  2,
					},
				},
				MapMap: map[string]map[string]string{
					"a": {
						"a": "1",
						"b": "2",
					},
					"b": nil,
				},
				Nil: nil,
			},
			BaseIf: Base{
				Int:    1,
				Bool:   true,
				String: "test",
				Float:  1,
			},
			NilPtr: nil,
		}

		var trace []string
		WalkValue(val, func(ctx *WalkContext, event WalkEvent) WalkingOp {
			switch e := event.(type) {
			case StructBeginEvent:
				trace = append(trace, "StructBeginEvent")
			case StructArriveFieldEvent:
				trace = append(trace, "StructFieldEvent", e.Info.Name)
				if isBaseDataType(e.Value) {
					trace = append(trace, toString(e.Value.Interface()))
				}
			case StructEndEvent:
				trace = append(trace, "StructEndEvent")

			case MapBeginEvent:
				trace = append(trace, "MapBeginEvent")
			case MapArriveEntryEvent:
				trace = append(trace, "MapEntryEvent", e.Key.String())
				if isBaseDataType(e.Value) {
					trace = append(trace, toString(e.Value.Interface()))
				}
			case MapEndEvent:
				trace = append(trace, "MapEndEvent")

			case ArrayBeginEvent:
				trace = append(trace, "ArrayBeginEvent")
			case ArrayArriveElementEvent:
				trace = append(trace, "ArrayElementEvent", fmt.Sprintf("%d", e.Index))
				if isBaseDataType(e.Value) {
					trace = append(trace, toString(e.Value.Interface()))
				}
			case ArrayEndEvent:
				trace = append(trace, "ArrayEndEvent")
			}

			return Next
		})

		So(trace, ShouldResemble, []string{})
	})
}
