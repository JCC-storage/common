package serder

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"

	"gitlink.org.cn/cloudream/common/pkgs/types"
)

func ObjectToJSON(obj any) ([]byte, error) {
	return json.Marshal(obj)
}

func ObjectToJSONStream(obj any) io.ReadCloser {
	pr, pw := io.Pipe()
	enc := json.NewEncoder(pw)

	go func() {
		err := enc.Encode(obj)
		if err != nil && err != io.EOF {
			pw.CloseWithError(err)
		} else {
			pw.Close()
		}
	}()

	return pr
}

func JSONToObject(data []byte, obj any) error {
	return json.Unmarshal(data, obj)
}

func JSONToObjectStream(str io.Reader, obj any) error {
	dec := json.NewDecoder(str)
	err := dec.Decode(obj)
	if err != io.EOF {
		return err
	}

	return nil
}

type TypeResolver interface {
	TypeToString(typ reflect.Type) (string, error)
	StringToType(typeStr string) (reflect.Type, error)
}

var registeredTaggedTypeUnions []*TaggedUnionType

type TaggedUnionType struct {
	Union          types.TypeUnion
	StrcutTagField string
	JSONTagField   string
	TagToType      map[string]reflect.Type
}

// 根据指定的字段的值来区分不同的类型。值可以通过在字段增加“union”Tag来指定。如果没有指定，则使用类型名。
func NewTaggedTypeUnion(union types.TypeUnion, structTagField string, jsonTagField string) *TaggedUnionType {
	tagToType := make(map[string]reflect.Type)

	for _, typ := range union.ElementTypes {
		if structTagField == "" {
			tagToType[typ.Name()] = typ
			continue
		}

		// 如果ElementType是一个指向结构体的指针，那么就遍历结构体的字段（解引用）
		structType := typ
		for structType.Kind() == reflect.Pointer {
			structType = structType.Elem()
		}

		field, ok := structType.FieldByName(structTagField)
		if !ok {
			tagToType[typ.Name()] = typ
			continue
		}

		tag := field.Tag.Get("union")
		if tag == "" {
			tagToType[typ.Name()] = typ
			continue
		}

		tagToType[tag] = typ
	}

	return &TaggedUnionType{
		Union:          union,
		StrcutTagField: structTagField,
		JSONTagField:   jsonTagField,
		TagToType:      tagToType,
	}
}

// 注册一个TaggedTypeUnion
func RegisterTaggedTypeUnion(union *TaggedUnionType) *TaggedUnionType {
	registeredTaggedTypeUnions = append(registeredTaggedTypeUnions, union)
	return union
}

// 创建并注册一个TaggedTypeUnion
func RegisterNewTaggedTypeUnion(union types.TypeUnion, structTagField string, jsonTagField string) *TaggedUnionType {
	taggedUnion := NewTaggedTypeUnion(union, structTagField, jsonTagField)
	RegisterTaggedTypeUnion(taggedUnion)
	return taggedUnion
}

type MapToObjectOption struct {
	UnionTypes             []*TaggedUnionType // 转换过程中遇到这些类型时，会依据指定的字段的值，来决定转换后的实际类型
	NoRegisteredUnionTypes bool               // 是否不使用全局注册的UnionType
}

func MapToObject(m map[string]any, obj any, opt ...MapToObjectOption) error {
	var op MapToObjectOption
	if len(opt) > 0 {
		op = opt[0]
	}

	unionTypeMapping := make(map[reflect.Type]*TaggedUnionType)

	for _, u := range op.UnionTypes {
		unionTypeMapping[u.Union.UnionType] = u
	}

	if !op.NoRegisteredUnionTypes {
		for _, u := range registeredTaggedTypeUnions {
			unionTypeMapping[u.Union.UnionType] = u
		}
	}

	convs := []Converter{
		func(from reflect.Value, to reflect.Value) (interface{}, error) {
			toType := to.Type()
			info, ok := unionTypeMapping[toType]
			if !ok {
				return from.Interface(), nil
			}

			mp := from.Interface().(map[string]any)
			tag, ok := mp[info.JSONTagField]
			if !ok {
				return nil, fmt.Errorf("converting to %v: no tag field %s in map", toType, info.JSONTagField)
			}

			tagStr, ok := tag.(string)
			if !ok {
				return nil, fmt.Errorf("converting to %v: tag field %s value is %v, which is not a string", toType, info.JSONTagField, tag)
			}

			eleType, ok := info.TagToType[tagStr]
			if !ok {
				return nil, fmt.Errorf("converting to %v: unknow type tag %s", toType, tagStr)
			}

			to.Set(reflect.New(eleType).Elem())

			return from.Interface(), nil
		},
	}

	return AnyToAny(m, obj, AnyToAnyOption{
		Converters: convs,
	})
}

func ObjectToMap(obj any) (map[string]any, error) {
	ctx := WalkValue(obj, func(ctx *WalkContext, event WalkEvent) WalkingOp {
		switch e := event.(type) {
		case StructBeginEvent:
			mp := make(map[string]any)
			ctx.StackPush(mp)

		case StructArriveFieldEvent:
			if !WillWalkInto(e.Value) {
				ctx.StackPush(e.Value.Interface())
			}
		case StructLeaveFieldEvent:
			val := ctx.StackPop()
			mp := ctx.StackPeek().(map[string]any)
			jsonTag := e.Info.Tag.Get("json")
			if jsonTag == "-" {
				break
			}

			opts := strings.Split(jsonTag, ",")
			keyName := opts[0]
			if keyName == "" {
				keyName = e.Info.Name
			}

			if contains(opts, "string", 1) {
				val = fmt.Sprintf("%v", val)
			}

			mp[keyName] = val

		case StructEndEvent:

		case MapBeginEvent:
			ctx.StackPush(make(map[string]any))
		case MapArriveEntryEvent:
			if !WillWalkInto(e.Value) {
				ctx.StackPush(e.Value.Interface())
			}
		case MapLeaveEntryEvent:
			val := ctx.StackPop()
			mp := ctx.StackPeek().(map[string]any)
			mp[fmt.Sprintf("%v", e.Key)] = val
		case MapEndEvent:

		case ArrayBeginEvent:
			ctx.StackPush(make([]any, e.Value.Len()))
		case ArrayArriveElementEvent:
			if !WillWalkInto(e.Value) {
				ctx.StackPush(e.Value.Interface())
			}
		case ArrayLeaveElementEvent:
			val := ctx.StackPop()
			arr := ctx.StackPeek().([]any)
			arr[e.Index] = val
		case ArrayEndEvent:
		}

		return Next

	}, WalkOption{
		StackValues: []any{make(map[string]any)},
	})

	return ctx.StackPop().(map[string]any), nil
}

func contains(arr []string, ele string, startIndex int) bool {
	for i := startIndex; i < len(arr); i++ {
		if arr[i] == ele {
			return true
		}
	}

	return false
}
