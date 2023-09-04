package serder

import (
	"fmt"
	"reflect"
)

type TypeNameResolver struct {
	includePackagePath bool
	types              map[string]reflect.Type
}

func NewTypeNameResolver(includePackagePath bool) *TypeNameResolver {
	return &TypeNameResolver{
		includePackagePath: includePackagePath,
		types:              make(map[string]reflect.Type),
	}
}

func (r *TypeNameResolver) Register(typ reflect.Type) {
	r.types[makeTypeString(typ, r.includePackagePath)] = typ
}

func (r *TypeNameResolver) TypeToString(typ reflect.Type) (string, error) {
	typeStr := makeTypeString(typ, r.includePackagePath)
	if _, ok := r.types[typeStr]; !ok {
		return "", fmt.Errorf("type %s is not registered before", typeStr)
	}

	return typeStr, nil
}

func (r *TypeNameResolver) StringToType(typeStr string) (reflect.Type, error) {
	typ, ok := r.types[typeStr]
	if !ok {
		return nil, fmt.Errorf("unknow type name %s", typeStr)
	}

	return typ, nil
}

func makeTypeString(typ reflect.Type, includePkgPath bool) string {
	if includePkgPath {
		return fmt.Sprintf("%s.%s", typ.PkgPath(), typ.Name())
	}

	return typ.Name()
}
