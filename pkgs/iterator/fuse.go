package iterator

type fuseError[T any] struct {
	err error
}

func (i *fuseError[T]) MoveNext() (T, error) {
	var ret T
	return ret, i.err
}
func (i *fuseError[T]) Close() {

}
func FuseError[T any](err error) Iterator[T] {
	return &fuseError[T]{
		err: err,
	}
}
