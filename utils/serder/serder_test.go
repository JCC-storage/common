package serder

import (
	"fmt"
	"reflect"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"gitlink.org.cn/cloudream/common/pkgs/types"
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

func Test_AnyToAny(t *testing.T) {
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
			Converters: []Converter{func(from reflect.Value, to reflect.Value) (interface{}, error) {
				if from.Type() == myreflect.TypeOf[Struct1]() && to.Type() == myreflect.TypeOf[Struct2]() {
					s1 := from.Interface().(Struct1)
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

func Test_MapToObject(t *testing.T) {
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
	Convey("结构体递归转换成map[string]any", t, func() {
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

		retMp, err := ObjectToMap(val)
		So(err, ShouldBeNil)

		exceptMap := map[string]any{
			"ArrSt": map[string]any{
				"IntArr": []any{1, 2, 3},
				"StArr": []any{
					map[string]any{
						"Int":    1,
						"Bool":   true,
						"String": "test",
						"Float":  1,
					},
					map[string]any{
						"Int":    2,
						"Bool":   false,
						"String": "test2",
						"Float":  2,
					},
				},
				"ArrArr": []any{
					[]any{1, 2, 3},
					[]any{},
					[]int(nil),
				},
				"Nil": []Base(nil),
			},
			"MapSt": map[string]any{
				"StrMap": map[string]any{
					"a": "1",
					"b": "2",
				},
				"StMap": map[string]any{
					"a": map[string]any{
						"Int":    1,
						"Bool":   true,
						"String": "test",
						"Float":  1,
					},
					"b": map[string]any{
						"Int":    2,
						"Bool":   false,
						"String": "test2",
						"Float":  2,
					},
				},
				"MapMap": map[string]any{
					"a": map[string]any{
						"a": "1",
						"b": "2",
					},
					"b": map[string]string(nil),
				},
				"Nil": map[string]Base(nil),
			},
			"BaseIf": map[string]any{
				"Int":    1,
				"Bool":   true,
				"String": "test",
				"Float":  1,
			},
			"NilPtr": (*Base)(nil),
		}

		mpRetJson, err := ObjectToJSON(retMp)
		So(err, ShouldBeNil)

		exceptMapJson, err := ObjectToJSON(exceptMap)
		So(err, ShouldBeNil)

		So(string(mpRetJson), ShouldEqualJSON, string(exceptMapJson))
	})

	Convey("包含UnionType", t, func() {
		type UnionType interface{}

		type EleType1 struct {
			Type   string `json:"type" union:"1"`
			Value1 string `json:"value1"`
		}

		type EleType2 struct {
			Type   string `json:"type" union:"2"`
			Value2 int    `json:"value2"`
		}

		type St struct {
			Us []UnionType `json:"us"`
		}

		mp := map[string]any{
			"us": []map[string]any{
				{
					"type":   "1",
					"value1": "1",
				},
				{
					"type":   "2",
					"value2": 2,
				},
			},
		}

		var ret St
		err := MapToObject(mp, &ret, MapToObjectOption{
			UnionTypes: []TaggedUnionType{
				NewTaggedTypeUnion(types.NewTypeUnion[UnionType](
					myreflect.TypeOf[EleType1](),
					myreflect.TypeOf[EleType2](),
				),
					"Type",
					"type",
				),
			},
		})

		So(err, ShouldBeNil)

		So(ret.Us, ShouldResemble, []UnionType{
			&EleType1{Type: "1", Value1: "1"},
			&EleType2{Type: "2", Value2: 2},
		})
	})

	Convey("要转换到的结构体就是一个UnionType", t, func() {
		type UnionType interface{}

		type EleType1 struct {
			Type   string `json:"type" union:"1"`
			Value1 string `json:"value1"`
		}

		type EleType2 struct {
			Type   string `json:"type" union:"2"`
			Value2 int    `json:"value2"`
		}

		mp := map[string]any{
			"type":   "1",
			"value1": "1",
		}

		var ret UnionType
		err := MapToObject(mp, &ret, MapToObjectOption{
			UnionTypes: []TaggedUnionType{
				NewTaggedTypeUnion(types.NewTypeUnion[UnionType](
					myreflect.TypeOf[EleType1](),
					myreflect.TypeOf[EleType2](),
				),
					"Type",
					"type",
				),
			},
		})

		So(err, ShouldBeNil)

		So(ret, ShouldResemble, &EleType1{Type: "1", Value1: "1"})
	})
}
