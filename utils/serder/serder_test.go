package serder

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	myreflect "gitlink.org.cn/cloudream/common/utils/reflect"
)

func Test_MapToObject(t *testing.T) {
	Convey("包含用字符串保存的int数据", t, func() {
		type Struct struct {
			A string `json:"a"`
			B int    `json:"b"`
			C int64  `json:"c,string"`
		}

		mp := map[string]any{
			"a": "a",
			"b": 1,
			"c": "1234",
		}

		var st Struct

		err := MapToObject(mp, &st)
		So(err, ShouldBeNil)

		So(st.A, ShouldEqual, "a")
		So(st.B, ShouldEqual, 1)
		So(st.C, ShouldEqual, 1234)
	})

}

func Test_TypedMapToObject(t *testing.T) {
	type Struct struct {
		A string `json:"a"`
		B int    `json:"b"`
		C int64  `json:"c,string"`
	}

	nameResovler := NewTypeNameResolver(true)
	nameResovler.Register(myreflect.GetGenericType[Struct]())

	Convey("结构体", t, func() {
		st := Struct{
			A: "a",
			B: 1,
			C: 2,
		}

		mp, err := ObjectToTypedMap(st, TypedSerderOption{
			TypeResolver:  &nameResovler,
			TypeFieldName: "@type",
		})

		So(err, ShouldBeNil)

		st2Ptr, err := TypedMapToObject(mp, TypedSerderOption{
			TypeResolver:  &nameResovler,
			TypeFieldName: "@type",
		})
		So(err, ShouldBeNil)

		st2, ok := st2Ptr.(Struct)
		So(ok, ShouldBeTrue)
		So(st2, ShouldHaveSameTypeAs, st)
		So(st2, ShouldResemble, st)
	})

}
