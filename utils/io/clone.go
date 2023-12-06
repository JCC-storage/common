package io

import (
	"io"
)

// 复制一个流。注：返回的多个流的读取不能在同一个线程，且如果不再需要读取返回的某个流，那么必须关闭这个流，否则会阻塞其他流的读取。
func Clone(str io.Reader, count int) []io.ReadCloser {
	prs := make([]io.ReadCloser, count)
	pws := make([]*io.PipeWriter, count)

	for i := 0; i < count; i++ {
		prs[i], pws[i] = io.Pipe()
	}

	go func() {
		pwCount := count
		buf := make([]byte, 4096)
		var closeErr error
		for {
			if pwCount == 0 {
				return
			}

			rd, err := str.Read(buf)
			for i := 0; i < count; i++ {
				if pws[i] == nil {
					continue
				}

				err := WriteAll(pws[i], buf[:rd])
				if err != nil {
					pws[i] = nil
					pwCount--
				}
			}

			if err == nil {
				continue
			}

			closeErr = err
			break
		}

		for i := 0; i < count; i++ {
			if pws[i] == nil {
				continue
			}
			pws[i].CloseWithError(closeErr)
		}
	}()

	return prs
}

/*
func BufferedClone(str io.Reader, count int, bufSize int) []io.ReadCloser {

}
*/
