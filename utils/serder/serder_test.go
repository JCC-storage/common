package serder

import (
	"fmt"
	"reflect"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	myreflect "gitlink.org.cn/cloudream/common/utils/reflect"
)

type FromAnyString struct {
	Str string
}

func (a *FromAnyString) FromAny(val any) (bool, error) {
	if str, ok := val.(string); ok {
		a.Str = "@" + str
		return true, nil
	}

	return false, nil
}

type ToAnyString struct {
	Str string
}

func (a *ToAnyString) ToAny(typ reflect.Type) (val any, ok bool, err error) {
	if typ == myreflect.TypeOf[map[string]any]() {
		return map[string]any{
			"str": "@" + a.Str,
		}, true, nil
	}

	return nil, false, nil
}

type FromAnySt struct {
	Value string
}

func (a *FromAnySt) FromAny(val any) (bool, error) {
	if st, ok := val.(ToAnySt); ok {
		a.Value = "From:" + st.Value
		return true, nil
	}

	return false, nil
}

type ToAnySt struct {
	Value string
}

func (a *ToAnySt) ToAny(typ reflect.Type) (val any, ok bool, err error) {
	if typ == myreflect.TypeOf[FromAnySt]() {
		return FromAnySt{
			Value: "To:" + a.Value,
		}, true, nil
	}

	return nil, false, nil
}

type DirToAnySt struct {
	Value string
}

func (a DirToAnySt) ToAny(typ reflect.Type) (val any, ok bool, err error) {
	if typ == myreflect.TypeOf[FromAnySt]() {
		return FromAnySt{
			Value: "DirTo:" + a.Value,
		}, true, nil
	}

	return nil, false, nil
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

		err := AnyToAny(mp, &st)
		So(err, ShouldBeNil)

		So(st.A, ShouldEqual, "a")
		So(st.B, ShouldEqual, 1)
		So(st.C, ShouldEqual, 1234)
	})

	Convey("只有FromAny", t, func() {
		type Struct struct {
			Special FromAnyString `json:"str"`
		}

		mp := map[string]any{
			"str": "test",
		}

		var ret Struct
		err := AnyToAny(mp, &ret)
		So(err, ShouldBeNil)

		So(ret.Special.Str, ShouldEqual, "@test")
	})

	Convey("字段类型直接实现了FromAny", t, func() {
		type Struct struct {
			Special *FromAnyString `json:"str"`
		}

		mp := map[string]any{
			"str": "test",
		}

		var ret Struct
		err := AnyToAny(mp, &ret)
		So(err, ShouldBeNil)

		So(ret.Special.Str, ShouldEqual, "@test")
	})

	Convey("只有ToAny", t, func() {
		st := struct {
			Special ToAnyString `json:"str"`
		}{
			Special: ToAnyString{
				Str: "test",
			},
		}

		ret := map[string]any{}

		err := AnyToAny(st, &ret)
		So(err, ShouldBeNil)

		So(ret["str"].(map[string]any)["str"], ShouldEqual, "@test")
	})

	Convey("优先使用ToAny", t, func() {
		st1 := ToAnySt{
			Value: "test",
		}

		st2 := FromAnySt{}

		err := AnyToAny(st1, &st2)
		So(err, ShouldBeNil)

		So(st2.Value, ShouldEqual, "To:test")
	})

	Convey("使用Convertor", t, func() {
		type Struct1 struct {
			Value string
		}

		type Struct2 struct {
			Value string
		}

		st1 := Struct1{
			Value: "test",
		}

		st2 := Struct2{}

		err := AnyToAny(st1, &st2, AnyToAnyOption{
			Converters: []Converter{func(srcType reflect.Type, dstType reflect.Type, data interface{}) (interface{}, error) {
				if srcType == myreflect.TypeOf[Struct1]() && dstType == myreflect.TypeOf[Struct2]() {
					s1 := data.(Struct1)
					return Struct2{
						Value: "@" + s1.Value,
					}, nil
				}

				return nil, fmt.Errorf("should not arrive here!")
			}},
		})
		So(err, ShouldBeNil)

		So(st2.Value, ShouldEqual, "@test")
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
