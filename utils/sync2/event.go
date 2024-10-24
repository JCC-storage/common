package sync2

import (
	"context"
	"errors"
	"sync"
)

var ErrEventClosed = errors.New("event is closed")
var ErrContextCanceled = errors.New("context canceled")

type Event struct {
	ch        chan any
	closeOnce sync.Once
}

func NewEvent() Event {
	return Event{
		ch: make(chan any, 1),
	}
}

func (e *Event) Set() {
	select {
	case e.ch <- nil:
	default:
	}
}

func (e *Event) Wait(ctx context.Context) error {
	select {
	case _, ok := <-e.ch:
		if ok {
			return nil
		}

		return ErrEventClosed

	case <-ctx.Done():
		return ErrContextCanceled
	}
}

func (e *Event) Close() {
	e.closeOnce.Do(func() {
		close(e.ch)
	})
}
