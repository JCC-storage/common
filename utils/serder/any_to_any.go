package serder

import (
	"reflect"

	mp "github.com/mitchellh/mapstructure"
	myreflect "gitlink.org.cn/cloudream/common/utils/reflect"
)

type Converter func(srcType reflect.Type, dstType reflect.Type, data interface{}) (interface{}, error)

type AnyToAnyOption struct {
	NoFromAny  bool        // 不判断目的字段是否实现了FromAny接口
	NoToAny    bool        // 不判断源字段是否实现了ToAny接口
	Converters []Converter // 字段类型转换函数
}

type FromAny interface {
	FromAny(val any) (ok bool, err error)
}

type ToAny interface {
	ToAny(typ reflect.Type) (val any, ok bool, err error)
}

// AnyToAny 相同结构的任意类型对象之间的转换
func AnyToAny(src any, dst any, opts ...AnyToAnyOption) error {
	var opt AnyToAnyOption
	if len(opts) > 0 {
		opt = opts[0]
	}

	var hooks []mp.DecodeHookFunc
	if !opt.NoToAny {
		hooks = append(hooks, toAny)
	}

	if !opt.NoFromAny {
		hooks = append(hooks, fromAny)
	}

	for _, c := range opt.Converters {
		hooks = append(hooks, c)
	}

	config := &mp.DecoderConfig{
		TagName:          "json",
		Squash:           true,
		WeaklyTypedInput: true,
		Result:           dst,
		DecodeHook:       mp.ComposeDecodeHookFunc(hooks...),
	}

	decoder, err := mp.NewDecoder(config)
	if err != nil {
		return err
	}

	return decoder.Decode(src)
}

// fromAny 如果目的字段实现的FromAny接口，那么通过此接口实现字段类型转换
func fromAny(srcType reflect.Type, targetType reflect.Type, data interface{}) (interface{}, error) {
	if myreflect.TypeOfValue(data) == targetType {
		return data, nil
	}

	if targetType.Implements(myreflect.TypeOf[FromAny]()) {
		// 非pointer receiver的FromAny没有意义，因为修改不了receiver的内容，所以这里只支持指针类型
		if targetType.Kind() == reflect.Pointer {
			val := reflect.New(targetType.Elem())
			anyIf := val.Interface().(FromAny)
			ok, err := anyIf.FromAny(data)
			if err != nil {
				return nil, err
			}
			if !ok {
				return data, nil
			}

			return val.Interface(), nil
		}

	} else if reflect.PointerTo(targetType).Implements(myreflect.TypeOf[FromAny]()) {
		val := reflect.New(targetType)
		anyIf := val.Interface().(FromAny)
		ok, err := anyIf.FromAny(data)
		if err != nil {
			return nil, err
		}
		if !ok {
			return data, nil
		}

		return val.Interface(), nil
	}

	return data, nil
}

// 如果源字段实现了ToAny接口，那么通过此接口实现字段类型转换
func toAny(srcType reflect.Type, targetType reflect.Type, data interface{}) (interface{}, error) {
	dataType := myreflect.TypeOfValue(data)
	if dataType == targetType {
		return data, nil
	}

	if dataType.Implements(myreflect.TypeOf[ToAny]()) {
		anyIf := data.(ToAny)
		dstVal, ok, err := anyIf.ToAny(targetType)
		if err != nil {
			return nil, err
		}
		if !ok {
			return data, nil
		}

		return dstVal, nil
	} else if reflect.PointerTo(dataType).Implements(myreflect.TypeOf[ToAny]()) {
		dataVal := reflect.ValueOf(data)

		dataPtrVal := reflect.New(dataType)
		dataPtrVal.Elem().Set(dataVal)

		anyIf := dataPtrVal.Interface().(ToAny)
		dstVal, ok, err := anyIf.ToAny(targetType)
		if err != nil {
			return nil, err
		}
		if !ok {
			return data, nil
		}

		return dstVal, nil
	}

	return data, nil
}
