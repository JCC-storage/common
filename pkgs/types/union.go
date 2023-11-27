package types

import (
	"fmt"
	"reflect"

	myreflect "gitlink.org.cn/cloudream/common/utils/reflect"
)

type AnyTypeUnion struct {
	// 这个集合的类型
	UnionType myreflect.Type
	// 集合中包含的类型，即遇到UnionType类型的值时，它内部的实际类型的范围
	ElementTypes []myreflect.Type
}

func (u *AnyTypeUnion) Include(typ myreflect.Type) bool {
	for _, t := range u.ElementTypes {
		if t == typ {
			return true
		}
	}

	return false
}

func (u *AnyTypeUnion) Add(typ myreflect.Type) error {
	if !typ.AssignableTo(u.UnionType) {
		return fmt.Errorf("type is not assignable to union type")
	}

	u.ElementTypes = append(u.ElementTypes, typ)
	return nil
}

// 描述一个类型集合
type TypeUnion[T any] struct {
	AnyTypeUnion
}

func (u *TypeUnion[T]) AddT(nilValue T) {
	u.ElementTypes = append(u.ElementTypes, reflect.TypeOf(nilValue))
}

func (u *TypeUnion[T]) ToAny() *AnyTypeUnion {
	return &u.AnyTypeUnion
}

// 创建一个TypeUnion。泛型参数为Union的类型，形参为Union中包含的类型的一个实例，无实际用途，仅用于获取类型。
func NewTypeUnion[TU any](eleValues ...TU) TypeUnion[TU] {
	var eleTypes []reflect.Type
	for _, v := range eleValues {
		eleTypes = append(eleTypes, reflect.TypeOf(v))
	}

	return TypeUnion[TU]{
		AnyTypeUnion{
			UnionType:    myreflect.TypeOf[TU](),
			ElementTypes: eleTypes,
		},
	}
}
