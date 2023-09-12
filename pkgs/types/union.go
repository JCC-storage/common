package types

import (
	myreflect "gitlink.org.cn/cloudream/common/utils/reflect"
)

// 描述一个类型集合
type TypeUnion struct {
	// 这个集合的类型
	UnionType myreflect.Type
	// 集合中包含的类型，即遇到UnionType类型的值时，它内部的实际类型的范围
	ElementTypes []myreflect.Type
}

func NewTypeUnion[TU any](eleTypes ...myreflect.Type) TypeUnion {
	return TypeUnion{
		UnionType:    myreflect.TypeOf[TU](),
		ElementTypes: eleTypes,
	}
}

func (u *TypeUnion) Include(typ myreflect.Type) bool {
	for _, t := range u.ElementTypes {
		if t == typ {
			return true
		}
	}

	return false
}

func (u *TypeUnion) Add(typ myreflect.Type) {
	u.ElementTypes = append(u.ElementTypes, typ)
}
