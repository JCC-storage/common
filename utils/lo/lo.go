package lo

import "github.com/samber/lo"

func Remove[T comparable](arr []T, item T) []T {
	index := lo.IndexOf(arr, item)
	if index == -1 {
		return arr
	}

	return RemoveAt(arr, index)
}

func RemoveAt[T any](arr []T, index int) []T {
	if index >= len(arr) {
		return arr
	}

	return append(arr[:index], arr[:index+1]...)
}

func ArrayClone[T any](arr []T) []T {
	return append([]T{}, arr...)
}
