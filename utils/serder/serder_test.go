package serder

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	myreflect "gitlink.org.cn/cloudream/common/utils/reflect"
)

type SpecialString struct {
	Str string
}

func (a *SpecialString) FromAny(val any) (bool, error) {
	if str, ok := val.(string); ok {
		a.Str = "@" + str
		return true, nil
	}

	return false, nil
}

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

	Convey("包含Time，先从结构体转为JSON，再从JSON转为Map，最后变回结构体", t, func() {
		type Struct struct {
			Time    time.Time
			NilTime *time.Time
		}

		var st = Struct{
			Time:    time.Now(),
			NilTime: nil,
		}

		data, err := ObjectToJSON(st)
		So(err, ShouldBeNil)

		var mp map[string]any
		err = JSONToObject(data, &mp)
		So(err, ShouldBeNil)

		var st2 Struct
		err = MapToObject(mp, &st2)
		So(err, ShouldBeNil)

		So(st.Time, ShouldEqual, st2.Time)
		So(st.NilTime, ShouldEqual, st2.NilTime)
	})

	Convey("使用FromAny", t, func() {
		type Struct struct {
			Special SpecialString `json:"str"`
		}

		mp := map[string]any{
			"str": "test",
		}

		var ret Struct
		err := AnyToAny(mp, &ret)
		So(err, ShouldBeNil)

		So(ret.Special.Str, ShouldEqual, "@test")
	})
}

func Test_TypedMapToObject(t *testing.T) {
	type Struct struct {
		A string `json:"a"`
		B int    `json:"b"`
		C int64  `json:"c,string"`
	}

	nameResovler := NewTypeNameResolver(true)
	nameResovler.Register(myreflect.TypeOf[Struct]())

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
