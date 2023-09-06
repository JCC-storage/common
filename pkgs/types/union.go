package types

import (
	myreflect "gitlink.org.cn/cloudream/common/utils/reflect"
)

type TypeUnion struct {
	UnionType    myreflect.Type
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
