package io

import (
	"fmt"
	"io"
)

type ChunkedSplitOption struct {
	// 如果流的长度不是chunkSize * streamCount的整数倍，启用此参数后，会在输出流里填充0直到满足长度
	PaddingZeros bool
}

// 每次读取一个chunkSize大小的数据，然后轮流写入到返回的流中。注：读取不同流的操作必须在不同的goroutine中进行，或者按顺序读取，每次精确读取一个chunkSize大小
func ChunkedSplit(stream io.Reader, chunkSize int, streamCount int, opts ...ChunkedSplitOption) []io.ReadCloser {
	var opt ChunkedSplitOption
	if len(opts) > 0 {
		opt = opts[0]
	}

	buf := make([]byte, chunkSize)
	prs := make([]io.ReadCloser, streamCount)
	pws := make([]*io.PipeWriter, streamCount)
	for i := 0; i < streamCount; i++ {
		pr, pw := io.Pipe()
		prs[i] = pr
		pws[i] = pw
	}

	go func() {
		var closeErr error
		eof := false
	loop:
		for {
			for i := 0; i < streamCount; i++ {
				var rd int = 0
				if !eof {
					var err error
					rd, err = io.ReadFull(stream, buf)
					if err == io.ErrUnexpectedEOF || err == io.EOF {
						eof = true
					} else if err != nil {
						closeErr = err
						break loop
					}
				}

				// 如果rd为0，那么肯定是eof，如果此时正好是在一轮读取的第一次，那么就直接退出整个读取，避免填充不必要的0
				if rd == 0 && i == 0 {
					break
				}

				if opt.PaddingZeros {
					Zero(buf[rd:])
					err := WriteAll(pws[i], buf)
					if err != nil {
						closeErr = fmt.Errorf("writing to one of the output streams: %w", err)
						break loop
					}
				} else {
					err := WriteAll(pws[i], buf[:rd])
					if err != nil {
						closeErr = fmt.Errorf("writing to one of the output streams: %w", err)
						break loop
					}
				}
			}

			if eof {
				break
			}
		}

		for _, pw := range pws {
			pw.CloseWithError(closeErr)
		}
	}()

	return prs
}
