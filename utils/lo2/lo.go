package lo2

import "github.com/samber/lo"

func Remove[T comparable](arr []T, item T) []T {
	index := lo.IndexOf(arr, item)
	if index == -1 {
		return arr
	}

	return RemoveAt(arr, index)
}

func RemoveAll[T comparable](arr []T, item T) []T {
	return lo.Filter(arr, func(i T, idx int) bool {
		return i != item
	})
}

func RemoveAt[T any](arr []T, index int) []T {
	if index >= len(arr) {
		return arr
	}

	return append(arr[:index], arr[index+1:]...)
}

func RemoveAllDefault[T comparable](arr []T) []T {
	var def T
	return lo.Filter(arr, func(i T, idx int) bool {
		return i != def
	})
}

func Clear[T comparable](arr []T, item T) {
	var def T
	for i := 0; i < len(arr); i++ {
		if arr[i] == item {
			arr[i] = def
		}
	}
}

func ArrayClone[T any](arr []T) []T {
	return append([]T{}, arr...)
}

func Insert[T any](arr []T, index int, item T) []T {
	arr = append(arr, item)
	copy(arr[index+1:], arr[index:])
	arr[index] = item
	return arr
}

func Deref[T any](arr []*T) []T {
	result := make([]T, len(arr))
	for i := 0; i < len(arr); i++ {
		result[i] = *arr[i]
	}

	return result
}
