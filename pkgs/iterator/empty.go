package iterator

type emptyIterator[T any] struct{}

func (i *emptyIterator[T]) MoveNext() (T, error) {
	var ret T
	return ret, ErrNoMoreItem
}
func (i *emptyIterator[T]) Close() {

}

func Empty[T any]() Iterator[T] {
	return &emptyIterator[T]{}
}
