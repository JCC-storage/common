package types

func Ref[T any](val T) *T {
	return &val
}
