package mq

import (
	"bytes"
	"fmt"
	"reflect"
	"unsafe"

	"github.com/google/uuid"
	jsoniter "github.com/json-iterator/go"
	myreflect "gitlink.org.cn/cloudream/common/utils/reflect"
	"gitlink.org.cn/cloudream/common/utils/serder"
)

type Message struct {
	Headers map[string]string `json:"headers"`
	Body    MessageBodyTypes  `json:"body"`
}

type MessageBodyTypes interface{}

func (m *Message) GetRequestID() string {
	return m.Headers["requestID"]
}

func (m *Message) SetRequestID(id string) {
	m.Headers["requestID"] = id
}

func (m *Message) MakeRequestID() string {
	id := uuid.NewString()
	m.Headers["requestID"] = id
	return id
}

func (m *Message) SetCodeMessage(code string, msg string) {
	m.Headers["responseCode"] = code
	m.Headers["responseMessage"] = msg
}

func (m *Message) GetCodeMessage() (string, string) {
	return m.Headers["responseCode"], m.Headers["responseMessage"]
}

func MakeMessage(body MessageBodyTypes) Message {
	msg := Message{
		Headers: make(map[string]string),
		Body:    body,
	}

	return msg
}

type typeSet struct {
	TopType      myreflect.Type
	ElementTypes serder.TypeNameResolver
}

var typeSets map[myreflect.Type]typeSet = make(map[reflect.Type]typeSet)
var messageTypeSet *typeSet

// 所有新定义的Message都需要在init中调用此函数
func RegisterMessage[T any]() {
	messageTypeSet.ElementTypes.Register(myreflect.TypeOf[T]())
}

// 如果对一个类型T调用了此函数，那么在序列化结构体中包含的T类型字段时，
// 会将字段值的实际类型保存在序列化后的结果中（作为一个字段@type），
// 在反序列化时，会根据类型信息重建原本的字段值。
//
// 只会处理types指定的类型。
func RegisterTypeSet[T any](types ...myreflect.Type) *typeSet {
	set := typeSet{
		TopType:      myreflect.TypeOf[T](),
		ElementTypes: serder.NewTypeNameResolver(true),
	}

	for _, t := range types {
		set.ElementTypes.Register(t)
	}

	typeSets[set.TopType] = set

	jsoniter.RegisterTypeEncoderFunc(myreflect.TypeOf[T]().String(),
		func(ptr unsafe.Pointer, stream *jsoniter.Stream) {
			val := *((*T)(ptr))
			var ifVal any = val

			if ifVal != nil {
				stream.WriteArrayStart()
				typeStr, err := set.ElementTypes.TypeToString(myreflect.TypeOfValue(val))
				if err != nil {
					stream.Error = err
					return
				}
				stream.WriteString(typeStr)
				stream.WriteRaw(",")
				stream.WriteVal(val)
				stream.WriteArrayEnd()
			} else {
				stream.WriteNil()
			}
		},
		func(p unsafe.Pointer) bool {
			return false
		})

	jsoniter.RegisterTypeDecoderFunc(myreflect.TypeOf[T]().String(),
		func(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
			vp := (*T)(ptr)

			nextTkType := iter.WhatIsNext()
			if nextTkType == jsoniter.NilValue {
				iter.ReadNil()
				var zero T
				*vp = zero
			} else if nextTkType == jsoniter.ArrayValue {
				iter.ReadArray()
				typeStr := iter.ReadString()
				iter.ReadArray()

				typ, err := set.ElementTypes.StringToType(typeStr)
				if err != nil {
					iter.ReportError("get type from string", err.Error())
					return
				}

				val := reflect.New(typ)
				iter.ReadVal(val.Interface())
				*vp = val.Elem().Interface().(T)

				iter.ReadArray()
			} else {
				iter.ReportError("parse TypeSet field", fmt.Sprintf("unknow next token type %v", nextTkType))
				return
			}
		})

	return &set
}

func Serialize(msg Message) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	enc := jsoniter.NewEncoder(buf)
	err := enc.Encode(msg)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func Deserialize(data []byte) (*Message, error) {
	dec := jsoniter.NewDecoder(bytes.NewBuffer(data))

	var msg Message
	err := dec.Decode(&msg)
	if err != nil {
		return nil, err
	}

	return &msg, nil
}

func init() {
	messageTypeSet = RegisterTypeSet[MessageBodyTypes]()
}
