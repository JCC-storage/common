package mq

import (
	"bytes"
	"fmt"
	"reflect"
	"unsafe"

	jsoniter "github.com/json-iterator/go"
	"gitlink.org.cn/cloudream/common/pkgs/types"
	myreflect "gitlink.org.cn/cloudream/common/utils/reflect"
)

const (
	MessageTypeAppData   = "AppData"
	MessageTypeHeartbeat = "Heartbeat"
)

type Message struct {
	Type    string         `json:"type"`
	Headers map[string]any `json:"headers"`
	Body    MessageBody    `json:"body"`
}

type MessageBody interface{}

func (m *Message) GetRequestID() string {
	reqID, _ := m.Headers["requestID"].(string)
	return reqID
}

func (m *Message) SetRequestID(id string) {
	m.Headers["requestID"] = id
}

func (m *Message) GetKeepAlive() int {
	timeoutMs, _ := m.Headers["keepAliveTimeout"].(float64)
	return int(timeoutMs)
}

func (m *Message) SetKeepAlive(timeoutMs int) {
	m.Headers["keepAliveTimeout"] = timeoutMs
}

func (m *Message) SetCodeMessage(code string, msg string) {
	m.Headers["responseCode"] = code
	m.Headers["responseMessage"] = msg
}

func (m *Message) GetCodeMessage() (string, string) {
	code, _ := m.Headers["responseCode"].(string)
	msg, _ := m.Headers["responseMessage"].(string)
	return code, msg
}

func MakeAppDataMessage(body MessageBody) Message {
	msg := Message{
		Type:    MessageTypeAppData,
		Headers: make(map[string]any),
		Body:    body,
	}

	return msg
}

func MakeHeartbeatMessage() Message {
	msg := Message{
		Type:    MessageTypeHeartbeat,
		Headers: make(map[string]any),
	}

	return msg
}

type TypeUnionWithTypeName struct {
	Union          types.TypeUnion
	TypeNameToType map[string]myreflect.Type
}

func (u *TypeUnionWithTypeName) Register(typ myreflect.Type) {
	u.Union.ElementTypes = append(msgBodyTypeUnion.Union.ElementTypes, typ)
	u.TypeNameToType[makeFullTypeName(typ)] = typ
}

var msgBodyTypeUnion *TypeUnionWithTypeName

// 所有新定义的Message都需要在init中调用此函数
func RegisterMessage[T any]() {
	msgBodyTypeUnion.Register(myreflect.TypeOf[T]())
}

// 在序列化结构体中包含的UnionType类型字段时，会将字段值的实际类型保存在序列化后的结果中。
// 在反序列化时，会根据类型信息重建原本的字段值。
func RegisterUnionType(union types.TypeUnion) *TypeUnionWithTypeName {
	myUnion := &TypeUnionWithTypeName{
		Union:          union,
		TypeNameToType: make(map[string]reflect.Type),
	}

	for _, typ := range union.ElementTypes {
		myUnion.TypeNameToType[makeFullTypeName(typ)] = typ
	}

	jsoniter.RegisterTypeEncoderFunc(union.UnionType.String(),
		func(ptr unsafe.Pointer, stream *jsoniter.Stream) {
			// 此处无法变成*UnionType，只能强转为*any
			val := *(*any)(ptr)
			if val != nil {
				stream.WriteArrayStart()

				valType := myreflect.TypeOfValue(val)
				if !myUnion.Union.Include(valType) {
					stream.Error = fmt.Errorf("type %v is not in union %v", valType, union.UnionType)
					return
				}

				stream.WriteString(makeFullTypeName(valType))
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

	jsoniter.RegisterTypeDecoderFunc(union.UnionType.String(),
		func(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
			// 此处无法变成*UnionType，只能强转为*any
			vp := (*any)(ptr)

			nextTkType := iter.WhatIsNext()
			if nextTkType == jsoniter.NilValue {
				iter.ReadNil()
				*vp = nil

			} else if nextTkType == jsoniter.ArrayValue {
				iter.ReadArray()
				typeStr := iter.ReadString()
				iter.ReadArray()

				typ, ok := myUnion.TypeNameToType[typeStr]
				if !ok {
					iter.ReportError("decode UnionType", fmt.Sprintf("unknow type string %s under %v", typeStr, union.UnionType))
					return
				}

				val := reflect.New(typ)
				iter.ReadVal(val.Interface())
				*vp = val.Elem().Interface()

				iter.ReadArray()
			} else {
				iter.ReportError("decode UnionType", fmt.Sprintf("unknow next token type %v", nextTkType))
				return
			}
		})

	return myUnion
}

func makeFullTypeName(typ myreflect.Type) string {
	return fmt.Sprintf("%s.%s", typ.PkgPath(), typ.Name())
}

/*
// 如果对一个类型T调用了此函数，那么在序列化结构体中包含的T类型字段时，
// 会将字段值的实际类型保存在序列化后的结果中
// 在反序列化时，会根据类型信息重建原本的字段值。
//
// 只会处理types指定的类型。
func RegisterTypeSet[T any](types ...myreflect.Type) *serder.UnionTypeInfo {
	eleTypes := serder.NewTypeNameResolver(true)
	set := serder.UnionTypeInfo{
		UnionType:    myreflect.TypeOf[T](),
		ElementTypes: eleTypes,
	}

	for _, t := range types {
		eleTypes.Register(t)
	}

		TODO 暂时保留这一段代码，如果RegisterUnionType中的非泛型版本出了问题，则重新使用这一部分的代码
			unionTypes[set.UnionType] = set

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
	RegisterUnionType(serder.NewTypeUnion[T]("", serder.NewTypeNameResolver(true)))
	return &set
}
*/

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
	msgBodyTypeUnion = RegisterUnionType(types.NewTypeUnion[MessageBody]())
}
