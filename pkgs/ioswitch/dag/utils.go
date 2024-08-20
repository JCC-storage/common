package dag

func NProps[T any](n *Node) T {
	return n.Props.(T)
}

func SProps[T any](str *StreamVar) T {
	return str.Props.(T)
}

func VProps[T any](v *ValueVar) T {
	return v.Props.(T)
}
