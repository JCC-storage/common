package sort2

import (
	"sort"

	"golang.org/x/exp/constraints"
)

type Comparer[T any] func(left T, right T) int

type sorter[T any] struct {
	arr []T
	cmp Comparer[T]
}

func (s sorter[T]) Len() int {
	return len(s.arr)
}

func (s sorter[T]) Less(i int, j int) bool {
	ret := s.cmp(s.arr[i], s.arr[j])
	return ret < 0
}

func (s sorter[T]) Swap(i int, j int) {
	s.arr[i], s.arr[j] = s.arr[j], s.arr[i]
}

func Sort[T any](arr []T, cmp Comparer[T]) []T {
	st := sorter[T]{
		arr: arr,
		cmp: cmp,
	}

	sort.Sort(st)
	return arr
}

func SortAsc[T constraints.Ordered](arr []T) []T {
	return Sort(arr, Cmp[T])
}

func SortDesc[T constraints.Ordered](arr []T) []T {
	return Sort(arr, func(left, right T) int { return Cmp(right, left) })
}

// false < true
func CmpBool(left, right bool) int {
	leftVal := 0
	if left {
		leftVal = 1
	}

	rightVal := 0
	if right {
		rightVal = 1
	}

	return leftVal - rightVal
}

func Cmp[T constraints.Ordered](left, right T) int {
	if left == right {
		return 0
	}

	if left < right {
		return -1
	}

	return 1
}
