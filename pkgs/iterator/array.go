package iterator

type ArrayIterator[T any] struct {
	arr   []T
	index int
}

func (i *ArrayIterator[T]) MoveNext() (T, error) {
	if i.index >= len(i.arr) {
		var ret T
		return ret, ErrNoMoreItem
	}

	item := i.arr[i.index]
	i.index++

	return item, nil
}

func (i *ArrayIterator[T]) Close() {
}

func Array[T any](eles ...T) *ArrayIterator[T] {
	return &ArrayIterator[T]{
		arr: eles,
	}
}
