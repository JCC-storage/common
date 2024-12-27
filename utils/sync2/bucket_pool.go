package sync2

import "sync"

type BucketPool[T any] struct {
	empty      []T
	filled     []T
	emptyCond  *sync.Cond
	filledCond *sync.Cond
}

func NewBucketPool[T any]() *BucketPool[T] {
	return &BucketPool[T]{
		emptyCond:  sync.NewCond(&sync.Mutex{}),
		filledCond: sync.NewCond(&sync.Mutex{}),
	}
}

func (p *BucketPool[T]) GetEmpty() (T, bool) {
	p.emptyCond.L.Lock()
	defer p.emptyCond.L.Unlock()

	if len(p.empty) == 0 {
		p.emptyCond.Wait()
	}

	if len(p.empty) == 0 {
		var t T
		return t, false
	}

	t := p.empty[0]
	p.empty = p.empty[1:]
	return t, true
}

func (p *BucketPool[T]) PutEmpty(t T) {
	p.emptyCond.L.Lock()
	defer p.emptyCond.L.Unlock()

	p.empty = append(p.empty, t)
	p.emptyCond.Signal()
}

func (p *BucketPool[T]) GetFilled() (T, bool) {
	p.filledCond.L.Lock()
	defer p.filledCond.L.Unlock()

	if len(p.filled) == 0 {
		p.filledCond.Wait()
	}

	if len(p.filled) == 0 {
		var t T
		return t, false
	}

	t := p.filled[0]
	p.filled = p.filled[1:]
	return t, true
}

func (p *BucketPool[T]) PutFilled(t T) {
	p.filledCond.L.Lock()
	defer p.filledCond.L.Unlock()

	p.filled = append(p.filled, t)
	p.filledCond.Signal()
}

func (p *BucketPool[T]) WakeUpAll() {
	p.emptyCond.Broadcast()
	p.filledCond.Broadcast()
}
