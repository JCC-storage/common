package serder

import (
	"encoding/json"
	"fmt"
	"reflect"
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
	var m map[string]any
	err := AnyToAny(obj, &m)
	return m, err
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
