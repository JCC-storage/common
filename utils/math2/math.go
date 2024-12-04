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
