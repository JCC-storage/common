package mq

import (
	"bytes"
	"fmt"
	"testing"
	"unsafe"

	jsoniter "github.com/json-iterator/go"
	. "github.com/smartystreets/goconvey/convey"
	myreflect "gitlink.org.cn/cloudream/common/utils/reflect"
	"gitlink.org.cn/cloudream/common/utils/serder"
)

func TestMessage(t *testing.T) {
	Convey("测试jsoniter", t, func() {

		type MyAny interface{}

		type Struct1 struct {
			Value string
		}

		type Struct2 struct {
			Value string
		}

		type Top struct {
			A1  Struct1
			A2  MyAny
			Nil MyAny
		}

		top := Top{
			A1: Struct1{
				Value: "s1",
			},
			A2: Struct2{
				Value: "s2",
			},
			Nil: nil,
		}

		jsoniter.RegisterTypeEncoderFunc(myreflect.TypeOf[MyAny]().String(),
			func(ptr unsafe.Pointer, stream *jsoniter.Stream) {
				val := *((*MyAny)(ptr))

				stream.WriteArrayStart()
				if val != nil {
					stream.WriteString(myreflect.TypeOfValue(val).String())
					stream.WriteRaw(",")
					stream.WriteVal(val)
				}
				stream.WriteArrayEnd()
			},
			func(p unsafe.Pointer) bool {
				return false
			})

		jsoniter.RegisterTypeDecoderFunc(myreflect.TypeOf[MyAny]().String(),
			func(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
				vp := (*MyAny)(ptr)

				nextTkType := iter.WhatIsNext()
				if nextTkType == jsoniter.NilValue {
					*vp = nil
				} else if nextTkType == jsoniter.ArrayValue {
					iter.ReadArray()
					typ := iter.ReadString()
					iter.ReadArray()

					if typ == "message.Struct1" {
						var st Struct1
						iter.ReadVal(&st)
						*vp = st
					} else if typ == "message.Struct2" {
						var st Struct2
						iter.ReadVal(&st)
						*vp = st
					}

					iter.ReadArray()
				}
			})

		buf := bytes.NewBuffer(nil)
		enc := jsoniter.NewEncoder(buf)
		err := enc.Encode(top)
		So(err, ShouldBeNil)

		dec := jsoniter.NewDecoder(buf)
		var newTop Top
		dec.Decode(&newTop)

		fmt.Printf("%s\n", buf.String())
		fmt.Printf("%#+v", newTop)
	})

	Convey("body中包含nil数组", t, func() {
		type Body struct {
			NilArr []string
		}
		RegisterMessage[Body]()

		msg := MakeAppDataMessage(Body{})
		data, err := Serialize(msg)
		So(err, ShouldBeNil)

		retMsg, err := Deserialize(data)
		So(err, ShouldBeNil)

		So(retMsg.Body.(Body).NilArr, ShouldBeNil)
	})

	Convey("body中包含匿名结构体", t, func() {
		type Emb struct {
			Value string `json:"value"`
		}
		type Body struct {
			Emb
		}
		RegisterMessage[Body]()

		msg := MakeAppDataMessage(Body{Emb: Emb{Value: "test"}})
		data, err := Serialize(msg)
		So(err, ShouldBeNil)

		retMsg, err := Deserialize(data)
		So(err, ShouldBeNil)

		So(retMsg.Body.(Body).Value, ShouldEqual, "test")
	})

	Convey("使用TypeSet类型，但字段值为nil", t, func() {
		type MyTypeSet interface {
			Test()
		}

		type Body struct {
			Value MyTypeSet
		}
		RegisterMessage[Body]()
		RegisterUnionType(serder.NewTypeUnion[MyTypeSet]("", serder.NewTypeNameResolver(true)))

		msg := MakeAppDataMessage(Body{Value: nil})
		data, err := Serialize(msg)
		So(err, ShouldBeNil)

		retMsg, err := Deserialize(data)
		So(err, ShouldBeNil)

		So(retMsg.Body.(Body).Value, ShouldBeNil)
	})

	Convey("字段实际类型不在TypeSet范围内", t, func() {
		type MyTypeSet interface{}

		type Body struct {
			Value MyTypeSet
		}
		RegisterMessage[Body]()
		RegisterUnionType(serder.NewTypeUnion[MyTypeSet]("", serder.NewTypeNameResolver(true)))

		msg := MakeAppDataMessage(Body{Value: struct{}{}})
		_, err := Serialize(msg)
		So(err, ShouldNotBeNil)
	})
}
