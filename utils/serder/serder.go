package serder

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"

	jsoniter "github.com/json-iterator/go"
)

var unionHandler = UnionHandler{
	internallyTagged: make(map[reflect.Type]*TypeUnionInternallyTagged),
	externallyTagged: make(map[reflect.Type]*TypeUnionExternallyTagged),
}

var defaultAPI = func() jsoniter.API {
	api := jsoniter.Config{
		EscapeHTML: true,
	}.Froze()

	api.RegisterExtension(&unionHandler)
	return api
}()

// 将对象转为JSON字符串。支持TypeUnion。
func ObjectToJSONEx[T any](obj T) ([]byte, error) {
	buf := new(bytes.Buffer)
	enc := defaultAPI.NewEncoder(buf)
	// 这里使用&obj而直接不使用obj的原因是，Encode的形参类型为any，
	// 如果T是一个interface类型，将obj传递进去后，内部拿到的类型将会是obj的实际类型，
	// 使用&obj，那么内部拿到的将会是*T类型，通过一层一层解引用查找Encoder时，能找到T对应的TypeUnion
	err := enc.Encode(&obj)

	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// 将JSON字符串转为对象。支持TypeUnion。
func JSONToObjectEx[T any](data []byte) (T, error) {
	var ret T
	dec := defaultAPI.NewDecoder(bytes.NewReader(data))
	err := dec.Decode(&ret)
	if err != nil {
		return ret, err
	}

	return ret, nil
}

// 将对象转为JSON字符串。如果需要支持解析TypeUnion类型，则使用"Ex"结尾的同名函数。
func ObjectToJSON(obj any) ([]byte, error) {
	return json.Marshal(obj)
}

// 将对象转为JSON字符串。如果需要支持解析TypeUnion类型，则使用"Ex"结尾的同名函数。
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

// 将JSON字符串转为对象。如果需要支持解析TypeUnion类型，则使用"Ex"结尾的同名函数。
func JSONToObject(data []byte, obj any) error {
	return json.Unmarshal(data, obj)
}

// 将JSON字符串转为对象。如果需要支持解析TypeUnion类型，则使用"Ex"结尾的同名函数。
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

type MapToObjectOption struct {
	NoRegisteredUnionTypes bool // 是否不使用全局注册的UnionType
}

// TODO 使用这个函数来处理TypeUnion的地方都可以直接使用Ex系列的函数
func MapToObject(m map[string]any, obj any, opt ...MapToObjectOption) error {
	var op MapToObjectOption
	if len(opt) > 0 {
		op = opt[0]
	}

	unionTypeMapping := make(map[reflect.Type]*TypeUnionInternallyTagged)

	if !op.NoRegisteredUnionTypes {
		for _, u := range unionHandler.internallyTagged {
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
			tag, ok := mp[info.TagField]
			if !ok {
				return nil, fmt.Errorf("converting to %v: no tag field %s in map", toType, info.TagField)
			}

			tagStr, ok := tag.(string)
			if !ok {
				return nil, fmt.Errorf("converting to %v: tag field %s value is %v, which is not a string", toType, info.TagField, tag)
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
