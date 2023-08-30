package serder

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"

	myreflect "gitlink.org.cn/cloudream/common/utils/reflect"
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

type UnionTypeInfo struct {
	UnionType     reflect.Type
	TypeFieldName string
	ElementTypes  TypeResolver
}

func NewTypeUnion[TU any](typeField string, eleTypes TypeResolver) UnionTypeInfo {
	return UnionTypeInfo{
		UnionType:     myreflect.TypeOf[TU](),
		TypeFieldName: typeField,
		ElementTypes:  eleTypes,
	}
}

type MapToObjectOption struct {
	UnionTypes []UnionTypeInfo // 转换过程中遇到这些类型时，会依据指定的字段的值，来决定转换后的实际类型
}

func MapToObject(m map[string]any, obj any, opt ...MapToObjectOption) error {
	var op MapToObjectOption
	if len(opt) > 0 {
		op = opt[0]
	}

	unionTypeMapping := make(map[reflect.Type]*UnionTypeInfo)

	for _, u := range op.UnionTypes {
		unionTypeMapping[u.UnionType] = &u
	}

	convs := []Converter{
		func(from reflect.Value, to reflect.Value) (interface{}, error) {
			info, ok := unionTypeMapping[to.Type()]
			if !ok {
				return from.Interface(), nil
			}

			mp := from.Interface().(map[string]any)
			tag, ok := mp[info.TypeFieldName]
			if !ok {
				return nil, fmt.Errorf("converting to %v: no tag field %s in map", to.Type(), info.TypeFieldName)
			}

			tagStr, ok := tag.(string)
			if !ok {
				return nil, fmt.Errorf("converting to %v: tag field %s value is %v, which is not a string", to.Type(), info.TypeFieldName, tag)
			}

			eleType, err := info.ElementTypes.StringToType(tagStr)
			if err != nil {
				return nil, fmt.Errorf("converting to %v: %w", to.Type(), err)
			}

			to.Set(reflect.Indirect(reflect.New(eleType)))

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
