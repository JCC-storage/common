package math

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
