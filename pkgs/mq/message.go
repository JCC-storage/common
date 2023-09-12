package mq

import (
	"bytes"
	"fmt"
	"reflect"
	"unsafe"

	jsoniter "github.com/json-iterator/go"
	"github.com/modern-go/reflect2"
	"gitlink.org.cn/cloudream/common/pkgs/types"
	myreflect "gitlink.org.cn/cloudream/common/utils/reflect"
)

const (
	MessageTypeAppData   = "AppData"
	MessageTypeHeartbeat = "Heartbeat"
)

type MessageBody interface {
	// 此方法无任何作用，仅用于避免MessageBody是一个空interface，从而导致任何类型的值都可以赋值给它
	// 与下方的MessageBodyBase配合使用：
	// IsMessageBody只让实现了此接口的类型能赋值给它，内嵌MessageBodyBase让类型必须是个指针类型，
	// 这就确保了Message.Body必是某个类型的指针类型，避免序列化、反序列化过程出错
	IsMessageBody()
}

// 这个结构体无任何字段，但实现了IsMessageBody，每种MessageBody都要内嵌这个结构体
type MessageBodyBase struct{}

// 此处的receiver是指针
func (b *MessageBodyBase) IsMessageBody() {}

type Message struct {
	Type    string         `json:"type"`
	Headers map[string]any `json:"headers"`
	Body    MessageBody    `json:"body"`
}

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
// 注：TypeUnion.UnionType必须是一个interface
func RegisterUnionType(union types.TypeUnion) *TypeUnionWithTypeName {
	myUnion := &TypeUnionWithTypeName{
		Union:          union,
		TypeNameToType: make(map[string]reflect.Type),
	}

	for _, typ := range union.ElementTypes {
		myUnion.TypeNameToType[makeFullTypeName(typ)] = typ
	}

	if union.UnionType.NumMethod() == 0 {
		registerForEFace(myUnion)
	} else {
		registerForIFace(myUnion)
	}

	return myUnion
}

// 无方法的interface类型
func registerForEFace(myUnion *TypeUnionWithTypeName) {
	jsoniter.RegisterTypeEncoderFunc(myUnion.Union.UnionType.String(),
		func(ptr unsafe.Pointer, stream *jsoniter.Stream) {
			// 无方法的interface底层数据结构都是eface类型，所以可以直接转*any
			val := *(*any)(ptr)
			if val != nil {
				stream.WriteArrayStart()

				valType := myreflect.TypeOfValue(val).Elem()
				if !myUnion.Union.Include(valType) {
					stream.Error = fmt.Errorf("type %v is not in union %v", valType, myUnion.Union.UnionType)
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

	jsoniter.RegisterTypeDecoderFunc(myUnion.Union.UnionType.String(),
		func(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
			// 无方法的interface底层都是eface结构体，所以可以直接转*any
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
					iter.ReportError("decode UnionType", fmt.Sprintf("unknow type string %s under %v", typeStr, myUnion.Union.UnionType))
					return
				}

				val := reflect.New(typ)
				iter.ReadVal(val.Interface())
				*vp = val.Interface()

				iter.ReadArray()
			} else {
				iter.ReportError("decode UnionType", fmt.Sprintf("unknow next token type %v", nextTkType))
				return
			}
		})
}

// 有方法的interface类型
func registerForIFace(myUnion *TypeUnionWithTypeName) {
	jsoniter.RegisterTypeEncoderFunc(myUnion.Union.UnionType.String(),
		func(ptr unsafe.Pointer, stream *jsoniter.Stream) {
			// 有方法的interface底层都是iface结构体，可以将其转成eface，转换后不损失类型信息
			val := reflect2.IFaceToEFace(ptr)
			if val != nil {
				stream.WriteArrayStart()

				// 此处肯定是指针类型，见MessageBody上的注释的分析
				valType := myreflect.TypeOfValue(val).Elem()
				if !myUnion.Union.Include(valType) {
					stream.Error = fmt.Errorf("type %v is not in union %v", valType, myUnion.Union.UnionType)
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

	jsoniter.RegisterTypeDecoderFunc(myUnion.Union.UnionType.String(),
		func(ptr unsafe.Pointer, iter *jsoniter.Iterator) {

			nextTkType := iter.WhatIsNext()
			if nextTkType == jsoniter.NilValue {
				iter.ReadNil()

			} else if nextTkType == jsoniter.ArrayValue {
				iter.ReadArray()
				typeStr := iter.ReadString()
				iter.ReadArray()

				typ, ok := myUnion.TypeNameToType[typeStr]
				if !ok {
					iter.ReportError("decode UnionType", fmt.Sprintf("unknow type string %s under %v", typeStr, myUnion.Union.UnionType))
					return
				}

				val := reflect.New(typ)
				iter.ReadVal(val.Interface())

				retVal := reflect.NewAt(myUnion.Union.UnionType, ptr)
				retVal.Elem().Set(val)

				iter.ReadArray()
			} else {
				iter.ReportError("decode UnionType", fmt.Sprintf("unknow next token type %v", nextTkType))
				return
			}
		})
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
