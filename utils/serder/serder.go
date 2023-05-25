package serder

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	mp "github.com/mitchellh/mapstructure"
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

func parseTimeHook(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if t != reflect.TypeOf(time.Time{}) {
		return data, nil
	}

	switch f.Kind() {
	case reflect.String:
		return time.Parse(time.RFC3339, data.(string))
	case reflect.Float64:
		return time.Unix(0, int64(data.(float64))*int64(time.Millisecond)), nil
	case reflect.Int64:
		return time.Unix(0, data.(int64)*int64(time.Millisecond)), nil
	default:
		return data, nil
	}
}

func MapToObject(m map[string]any, obj any) error {
	config := &mp.DecoderConfig{
		TagName:          "json",
		Squash:           true,
		WeaklyTypedInput: true,
		Result:           obj,
		DecodeHook:       mp.ComposeDecodeHookFunc(parseTimeHook),
	}

	decoder, err := mp.NewDecoder(config)
	if err != nil {
		return err
	}

	return decoder.Decode(m)
}

func ObjectToMap(obj any) (map[string]any, error) {
	var retMap map[string]any
	config := &mp.DecoderConfig{
		TagName:          "json",
		Squash:           true,
		WeaklyTypedInput: true,
		Result:           &retMap,
	}

	decoder, err := mp.NewDecoder(config)
	if err != nil {
		return nil, err
	}

	err = decoder.Decode(obj)
	return retMap, err
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
	err = MapToObject(m, valPtr)
	if err != nil {
		return nil, err
	}

	return val.Elem().Interface(), nil
}

func ObjectToTypedMap(obj any, opt TypedSerderOption) (map[string]any, error) {
	mp, err := ObjectToMap(obj)
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
