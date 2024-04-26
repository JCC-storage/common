package serder

import (
	"fmt"
	"reflect"
	"unsafe"

	jsoniter "github.com/json-iterator/go"
	"github.com/modern-go/reflect2"
	"gitlink.org.cn/cloudream/common/pkgs/types"

	ref2 "gitlink.org.cn/cloudream/common/utils/reflect2"
)

type anyTypeUnionExternallyTagged struct {
	Union          *types.AnyTypeUnion
	TypeNameToType map[string]reflect.Type
}

type TypeUnionExternallyTagged[T any] struct {
	anyTypeUnionExternallyTagged
	TUnion *types.TypeUnion[T]
}

// 遇到TypeUnion的基类（UnionType）的字段时，将其实际值的类型信息也编码到JSON中，反序列化时也会根据解析出类型信息，还原出真实的类型。
// Externally Tagged的格式是：{ "类型名": {...对象内容...} }
//
// 可以通过内嵌Metadata结构体，并在它身上增加"union"Tag来指定类型名称，如果没有指定，则默认使用系统类型名（包括包路径）。
func UseTypeUnionExternallyTagged[T any](union *types.TypeUnion[T]) *TypeUnionExternallyTagged[T] {
	eu := &TypeUnionExternallyTagged[T]{
		anyTypeUnionExternallyTagged: anyTypeUnionExternallyTagged{
			Union:          union.ToAny(),
			TypeNameToType: make(map[string]reflect.Type),
		},
		TUnion: union,
	}

	for _, eleType := range union.ElementTypes {
		eu.Add(eleType)
	}

	unionHandler.externallyTagged[union.UnionType] = &eu.anyTypeUnionExternallyTagged

	return eu
}

func (u *TypeUnionExternallyTagged[T]) Add(typ reflect.Type) error {
	err := u.TUnion.Add(typ)
	if err != nil {
		return nil
	}

	u.TypeNameToType[makeDerefFullTypeName(typ)] = typ
	return nil
}

func (u *TypeUnionExternallyTagged[T]) AddT(nilValue T) error {
	u.Add(reflect.TypeOf(nilValue))
	return nil
}

type anyTypeUnionInternallyTagged struct {
	Union     *types.AnyTypeUnion
	TagField  string
	TagToType map[string]reflect.Type
}

type TypeUnionInternallyTagged[T any] struct {
	anyTypeUnionInternallyTagged
	TUnion *types.TypeUnion[T]
}

// 遇到TypeUnion的基类（UnionType）的字段时，将其实际值的类型信息也编码到JSON中，反序列化时也会解析出类型信息，还原出真实的类型。
// Internally Tagged的格式是：{ "类型字段": "类型名", ...对象内容...}，JSON中的类型字段名需要指定。
// 注：对象定义需要包含类型字段，而且在序列化之前需要手动赋值，目前不支持自动设置。
//
// 可以通过内嵌Metadata结构体，并在它身上增加"union"Tag来指定类型名称，如果没有指定，则默认使用系统类型名（包括包路径）。
func UseTypeUnionInternallyTagged[T any](union *types.TypeUnion[T], tagField string) *TypeUnionInternallyTagged[T] {
	iu := &TypeUnionInternallyTagged[T]{
		anyTypeUnionInternallyTagged: anyTypeUnionInternallyTagged{
			Union:     union.ToAny(),
			TagField:  tagField,
			TagToType: make(map[string]reflect.Type),
		},
		TUnion: union,
	}

	for _, eleType := range union.ElementTypes {
		iu.Add(eleType)
	}

	unionHandler.internallyTagged[union.UnionType] = &iu.anyTypeUnionInternallyTagged
	return iu
}

func (u *TypeUnionInternallyTagged[T]) Add(typ reflect.Type) error {
	err := u.Union.Add(typ)
	if err != nil {
		return nil
	}

	// 解引用直到得到结构体类型
	structType := typ
	for structType.Kind() == reflect.Pointer {
		structType = structType.Elem()
	}

	// 要求内嵌Metadata结构体，那么结构体中的字段名就会是Metadata，
	field, ok := structType.FieldByName(ref2.TypeNameOf[Metadata]())
	if !ok {
		u.TagToType[makeDerefFullTypeName(structType)] = typ
		return nil
	}

	// 为防同名，检查类型是不是也是Metadata
	if field.Type != ref2.TypeOf[Metadata]() {
		u.TagToType[makeDerefFullTypeName(structType)] = typ
		return nil
	}

	tag := field.Tag.Get("union")
	if tag == "" {
		u.TagToType[makeDerefFullTypeName(structType)] = typ
		return nil
	}

	u.TagToType[tag] = typ
	return nil
}

func (u *TypeUnionInternallyTagged[T]) AddT(nilValue T) error {
	u.Add(reflect.TypeOf(nilValue))
	return nil
}

type UnionHandler struct {
	internallyTagged map[reflect.Type]*anyTypeUnionInternallyTagged
	externallyTagged map[reflect.Type]*anyTypeUnionExternallyTagged
}

func (h *UnionHandler) UpdateStructDescriptor(structDescriptor *jsoniter.StructDescriptor) {

}

func (h *UnionHandler) CreateMapKeyDecoder(typ reflect2.Type) jsoniter.ValDecoder {
	return nil
}

func (h *UnionHandler) CreateMapKeyEncoder(typ reflect2.Type) jsoniter.ValEncoder {
	return nil
}

func (h *UnionHandler) CreateDecoder(typ reflect2.Type) jsoniter.ValDecoder {
	typ1 := typ.Type1()
	if it, ok := h.internallyTagged[typ1]; ok {
		return &InternallyTaggedDecoder{
			union: it,
		}
	}

	if et, ok := h.externallyTagged[typ1]; ok {
		return &ExternallyTaggedDecoder{
			union: et,
		}
	}

	return nil
}

func (h *UnionHandler) CreateEncoder(typ reflect2.Type) jsoniter.ValEncoder {
	typ1 := typ.Type1()
	if it, ok := h.internallyTagged[typ1]; ok {
		return &InternallyTaggedEncoder{
			union: it,
		}
	}

	if et, ok := h.externallyTagged[typ1]; ok {
		return &ExternallyTaggedEncoder{
			union: et,
		}
	}
	return nil
}

func (h *UnionHandler) DecorateDecoder(typ reflect2.Type, decoder jsoniter.ValDecoder) jsoniter.ValDecoder {
	return decoder
}

func (h *UnionHandler) DecorateEncoder(typ reflect2.Type, encoder jsoniter.ValEncoder) jsoniter.ValEncoder {
	return encoder
}

// 以下Encoder/Decoder都是在传入类型/目标类型是TypeUnion的基类（UnionType）时使用
type InternallyTaggedEncoder struct {
	union *anyTypeUnionInternallyTagged
}

func (e *InternallyTaggedEncoder) IsEmpty(ptr unsafe.Pointer) bool {
	return false
}

func (e *InternallyTaggedEncoder) Encode(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	var val any

	if e.union.Union.UnionType.NumMethod() == 0 {
		// 无方法的interface底层都是eface结构体，所以可以直接转*any
		val = *(*any)(ptr)
	} else {
		// 有方法的interface底层都是iface结构体，可以将其转成eface，转换后不损失类型信息
		val = reflect2.IFaceToEFace(ptr)
	}

	// 可以考虑检查一下Type字段有没有赋值，没有赋值则将其赋值为union Tag指定的值
	stream.WriteVal(val)
}

type InternallyTaggedDecoder struct {
	union *anyTypeUnionInternallyTagged
}

func (e *InternallyTaggedDecoder) Decode(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
	nextTokenKind := iter.WhatIsNext()
	if nextTokenKind == jsoniter.NilValue {
		iter.Skip()
		return
	}

	raw := iter.ReadAny()
	if raw.LastError() != nil {
		iter.ReportError("decode TaggedUnionType", "getting object raw:"+raw.LastError().Error())
		return
	}

	tagField := raw.Get(e.union.TagField)
	if tagField.LastError() != nil {
		iter.ReportError("decode TaggedUnionType", "getting type tag field:"+tagField.LastError().Error())
		return
	}

	typeTag := tagField.ToString()
	if typeTag == "" {
		iter.ReportError("decode TaggedUnionType", "type tag is empty")
		return
	}

	typ, ok := e.union.TagToType[typeTag]
	if !ok {
		iter.ReportError("decode TaggedUnionType", fmt.Sprintf("unknow type tag %s in union %s", typeTag, e.union.Union.UnionType.Name()))
		return
	}

	// 如果目标类型已经是个指针类型*T，那么在New的时候就需要使用T，
	// 否则New出来的是会是**T，这将导致后续的反序列化出问题
	if typ.Kind() == reflect.Pointer {
		val := reflect.New(typ.Elem())
		raw.ToVal(val.Interface())

		retVal := reflect.NewAt(e.union.Union.UnionType, ptr)
		retVal.Elem().Set(val)

	} else {
		val := reflect.New(typ)
		raw.ToVal(val.Interface())

		retVal := reflect.NewAt(e.union.Union.UnionType, ptr)
		retVal.Elem().Set(val.Elem())
	}
}

type ExternallyTaggedEncoder struct {
	union *anyTypeUnionExternallyTagged
}

func (e *ExternallyTaggedEncoder) IsEmpty(ptr unsafe.Pointer) bool {
	return false
}

func (e *ExternallyTaggedEncoder) Encode(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	var val any

	if e.union.Union.UnionType.NumMethod() == 0 {
		// 无方法的interface底层都是eface结构体，所以可以直接转*any
		val = *(*any)(ptr)
	} else {
		// 有方法的interface底层都是iface结构体，可以将其转成eface，转换后不损失类型信息
		val = reflect2.IFaceToEFace(ptr)
	}

	if val == nil {
		stream.WriteNil()
		return
	}

	stream.WriteObjectStart()
	valType := ref2.TypeOfValue(val)
	if !e.union.Union.Include(valType) {
		stream.Error = fmt.Errorf("type %v is not in union %v", valType, e.union.Union.UnionType)
		return
	}
	stream.WriteObjectField(makeDerefFullTypeName(valType))
	stream.WriteVal(val)
	stream.WriteObjectEnd()
}

type ExternallyTaggedDecoder struct {
	union *anyTypeUnionExternallyTagged
}

func (e *ExternallyTaggedDecoder) Decode(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
	nextTkType := iter.WhatIsNext()

	if nextTkType == jsoniter.NilValue {
		iter.Skip()
		return
	}

	if nextTkType != jsoniter.ObjectValue {
		iter.ReportError("decode UnionType", fmt.Sprintf("unknow next token type %v", nextTkType))
		return
	}

	typeStr := iter.ReadObject()
	if typeStr == "" {
		iter.ReportError("decode UnionType", "type string is empty")
	}

	typ, ok := e.union.TypeNameToType[typeStr]
	if !ok {
		iter.ReportError("decode UnionType", fmt.Sprintf("unknow type string %s in union %v", typeStr, e.union.Union.UnionType))
		return
	}

	// 如果目标类型已经是个指针类型*T，那么在New的时候就需要使用T，
	// 否则New出来的是会是**T，这将导致后续的反序列化出问题
	if typ.Kind() == reflect.Pointer {
		val := reflect.New(typ.Elem())
		iter.ReadVal(val.Interface())

		retVal := reflect.NewAt(e.union.Union.UnionType, ptr)
		retVal.Elem().Set(val)

	} else {
		val := reflect.New(typ)
		iter.ReadVal(val.Interface())

		retVal := reflect.NewAt(e.union.Union.UnionType, ptr)
		retVal.Elem().Set(val.Elem())
	}

	if iter.ReadObject() != "" {
		iter.ReportError("decode UnionType", "there should be only one fields in the json object")
	}
}

func makeDerefFullTypeName(typ reflect.Type) string {
	realType := typ
	for realType.Kind() == reflect.Pointer {
		realType = realType.Elem()
	}
	return fmt.Sprintf("%s.%s", realType.PkgPath(), realType.Name())
}
