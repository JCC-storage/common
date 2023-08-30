package serder

import (
	"fmt"
	"reflect"
)

type StringTypeResolver struct {
	strToType map[string]reflect.Type
	typeToStr map[reflect.Type]string
}

func NewStringTypeResolver() *StringTypeResolver {
	return &StringTypeResolver{
		strToType: make(map[string]reflect.Type),
		typeToStr: make(map[reflect.Type]string),
	}
}

func (r *StringTypeResolver) Add(str string, typ reflect.Type) *StringTypeResolver {
	r.strToType[str] = typ
	r.typeToStr[typ] = str
	return r
}

func (r *StringTypeResolver) TypeToString(typ reflect.Type) (string, error) {
	var typeStr string
	var ok bool
	if typeStr, ok = r.typeToStr[typ]; !ok {
		return "", fmt.Errorf("type %s is not registered before", typ)
	}

	return typeStr, nil
}

func (r *StringTypeResolver) StringToType(typeStr string) (reflect.Type, error) {
	typ, ok := r.strToType[typeStr]
	if !ok {
		return nil, fmt.Errorf("unknow type string %s", typeStr)
	}

	return typ, nil
}
