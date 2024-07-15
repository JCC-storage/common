package sync2

import (
	"sync"
)

func ParallelDo[T any](args []T, fn func(val T, index int) error) error {
	lock := sync.Mutex{}
	var err error

	var wg sync.WaitGroup
	wg.Add(len(args))
	for i, arg := range args {
		go func(arg T, index int) {
			defer wg.Done()

			if e := fn(arg, index); e != nil {
				lock.Lock()
				if err == nil {
					err = e
				}
				lock.Unlock()
			}
		}(arg, i)
	}
	wg.Wait()

	return err
}
