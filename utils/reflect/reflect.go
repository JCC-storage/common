package reflect

import "reflect"

type Type = reflect.Type

// TypeOfValue 获得实际值的类型
func TypeOfValue(val any) reflect.Type {
	return reflect.TypeOf(val)
}

// TypeOf 获得泛型的类型
func TypeOf[T any]() reflect.Type {
	return reflect.TypeOf([0]T{}).Elem()
}

// ElemTypeOf 获得泛型的类型。适用于数组、指针类型
func ElemTypeOf[T any]() reflect.Type {
	return reflect.TypeOf([0]T{}).Elem().Elem()
}

func TypeNameOf[T any]() string {
	return TypeOf[T]().Name()
}
