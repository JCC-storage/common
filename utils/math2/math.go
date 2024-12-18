package math2

import "golang.org/x/exp/constraints"

func Max[T constraints.Ordered](v1, v2 T) T {
	if v1 < v2 {
		return v2
	}

	return v1
}

func Min[T constraints.Ordered](v1, v2 T) T {
	if v1 < v2 {
		return v1
	}

	return v2
}

func Ceil[T constraints.Integer](v T, div T) T {
	return (v + div - 1) / div * div
}

func Floor[T constraints.Integer](v T, div T) T {
	return v / div * div
}

func CeilDiv[T constraints.Integer](v T, div T) T {
	return (v + div - 1) / div
}

func FloorDiv[T constraints.Integer](v T, div T) T {
	return v / div
}

func Clamp[T constraints.Integer](v, min, max T) T {
	if v < min {
		return min
	}

	if v > max {
		return max
	}

	return v
}

// 将一个整数切分成小于maxValue的整数列表，尽量均匀
func SplitLessThan[T constraints.Integer](v T, maxValue T) []T {
	cnt := int(CeilDiv(v, maxValue))
	result := make([]T, cnt)
	last := int64(0)
	for i := 0; i < cnt; i++ {
		cur := int64(v) * int64(i+1) / int64(cnt)
		result[i] = T(cur - last)
		last = cur
	}

	return result
}

// 将一个整数切分成n个整数，尽量均匀
func SplitN[T constraints.Integer](v T, n int) []T {
	result := make([]T, n)
	last := int64(0)
	for i := 0; i < n; i++ {
		cur := int64(v) * int64(i+1) / int64(n)
		result[i] = T(cur - last)
		last = cur
	}

	return result
}
