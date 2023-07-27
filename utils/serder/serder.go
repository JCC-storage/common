package serder

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

func ObjectToJSON(obj any) ([]byte, error) {
	return json.Marshal(obj)
}

func JSONToObject(data []byte, obj any) error {
	return json.Unmarshal(data, obj)
}

type TypeResolver interface {
	TypeToString(typ reflect.Type) (string, error)
	StringToType(typeStr string) (reflect.Type, error)
}

type TypedSerderOption struct {
	TypeResolver  TypeResolver
	TypeFieldName string
}

func MapToObject(m map[string]any, obj any) error {
	return AnyToAny(m, obj)
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

func TypedMapToObject(m map[string]any, opt TypedSerderOption) (any, error) {

	typeVal, ok := m[opt.TypeFieldName]
	if !ok {
		return nil, fmt.Errorf("no type field in the map")
	}

	typeStr, ok := typeVal.(string)
	if !ok {
		return nil, fmt.Errorf("type is not a string")
	}

	typ, err := opt.TypeResolver.StringToType(typeStr)
	if err != nil {
		return nil, fmt.Errorf("get type from string failed, err: %w", err)
	}

	val := reflect.New(typ)

	valPtr := val.Interface()
	err = AnyToAny(m, valPtr)
	if err != nil {
		return nil, err
	}

	return val.Elem().Interface(), nil
}

func ObjectToTypedMap(obj any, opt TypedSerderOption) (map[string]any, error) {
	var mp map[string]any
	err := AnyToAny(obj, &mp)
	if err != nil {
		return nil, err
	}

	_, ok := mp[opt.TypeFieldName]
	if ok {
		return nil, fmt.Errorf("object has the same field as the type field")
	}

	mp[opt.TypeFieldName], err = opt.TypeResolver.TypeToString(reflect.TypeOf(obj))
	if err != nil {
		return nil, fmt.Errorf("get string from type failed, err: %w", err)
	}

	return mp, nil
}
