package union

import (
	"fmt"
	"reflect"

	"gitlink.org.cn/cloudream/common/utils/serder"
)

type Value[T any] struct {
	Value T
}

func (o *Value[T]) Scan(src interface{}) error {
	data, ok := src.([]uint8)
	if !ok {
		return fmt.Errorf("unknow src type: %v", reflect.TypeOf(data))
	}

	val, err := serder.JSONToObjectEx[T](data)
	if err != nil {
		return err
	}

	o.Value = val
	return nil
}
