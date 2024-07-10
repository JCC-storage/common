package sync2

import (
	"sync"
	"sync/atomic"
)

func ParallelDo[T any](args []T, fn func(val T, index int) error) error {
	err := atomic.Value{}
	err.Store((error)(nil))

	var wg sync.WaitGroup
	wg.Add(len(args))
	for i, arg := range args {
		go func(arg T, index int) {
			defer wg.Done()

			if e := fn(arg, index); e != nil {
				err.CompareAndSwap((error)(nil), e)
			}
		}(arg, i)
	}
	wg.Wait()
	return err.Load().(error)
}
