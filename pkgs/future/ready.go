package future

import "context"

type Ready struct {
	ch chan error
}

func NewReady(err error) *Ready {
	ch := make(chan error, 1)
	ch <- err
	close(ch)

	return &Ready{
		ch: ch,
	}
}

func (f *Ready) IsComplete() bool {
	return true
}

func (f *Ready) Wait(ctx context.Context) error {
	select {
	case v, ok := <-f.ch:
		if !ok {
			return ErrCompleted
		}
		return v

	case <-ctx.Done():
		return context.Canceled
	}
}

func (f *Ready) Chan() <-chan error {
	return f.ch
}

type Ready1[T any] struct {
	ch chan ChanValue1[T]
}

func NewReady1[T any](val T, err error) *Ready1[T] {
	ch := make(chan ChanValue1[T], 1)
	ch <- ChanValue1[T]{
		Err:   err,
		Value: val,
	}
	close(ch)

	return &Ready1[T]{
		ch: ch,
	}
}

func NewReadyValue1[T any](val T) *Ready1[T] {
	return NewReady1[T](val, nil)
}

func NewReadyError1[T any](err error) *Ready1[T] {
	var ret T
	return NewReady1[T](ret, err)
}

func (f *Ready1[T]) IsComplete() bool {
	return true
}

func (f *Ready1[T]) Wait(ctx context.Context) (T, error) {
	select {
	case cv, ok := <-f.ch:
		if !ok {
			var ret T
			return ret, cv.Err
		}
		return cv.Value, cv.Err

	case <-ctx.Done():
		var ret T
		return ret, context.Canceled
	}
}

func (f *Ready1[T]) Chan() <-chan ChanValue1[T] {
	return f.ch
}
