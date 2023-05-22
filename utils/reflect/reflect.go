package reflect

import "reflect"

// GetGenericType 获得泛型的类型
func GetGenericType[T any]() reflect.Type {
	return reflect.TypeOf([0]T{}).Elem()
}

// GetGenericElemType 获得泛型的类型。适用于数组、指针类型
func GetGenericElemType[T any]() reflect.Type {
	return reflect.TypeOf([0]T{}).Elem().Elem()
}
