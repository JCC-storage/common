package io2

import (
	"io"

	"gitlink.org.cn/cloudream/common/utils/lo2"
	"gitlink.org.cn/cloudream/common/utils/math2"
	"gitlink.org.cn/cloudream/common/utils/sync2"
)

func Join(strs []io.Reader) io.ReadCloser {

	pr, pw := io.Pipe()

	go func() {
		var closeErr error

		buf := make([]byte, 4096)
	outer:
		for _, str := range strs {
			for {
				bufLen := len(buf)
				if bufLen == 0 {
					break outer
				}

				rd, err := str.Read(buf[:bufLen])
				if err != nil {
					if err != io.EOF {
						closeErr = err
						break outer
					}

					err = WriteAll(pw, buf[:rd])
					if err != nil {
						closeErr = err
						break outer
					}

					break
				}

				err = WriteAll(pw, buf[:rd])
				if err != nil {
					closeErr = err
					break outer
				}
			}
		}

		pw.CloseWithError(closeErr)
	}()

	return pr
}

type chunkedJoin struct {
	inputs       []io.Reader
	chunkSize    int
	currentInput int
	currentRead  int
	err          error
}

func (s *chunkedJoin) Read(buf []byte) (int, error) {
	if s.err != nil {
		return 0, s.err
	}
	if len(s.inputs) == 0 {
		return 0, io.EOF
	}

	bufLen := math2.Min(math2.Min(s.chunkSize, len(buf)), s.chunkSize-s.currentRead)
	rd, err := s.inputs[s.currentInput].Read(buf[:bufLen])
	if err == nil {
		s.currentRead += rd
		if s.currentRead == s.chunkSize {
			s.currentInput = (s.currentInput + 1) % len(s.inputs)
			s.currentRead = 0
		}
		return rd, nil
	}

	if err == io.EOF {
		s.inputs = lo2.RemoveAt(s.inputs, s.currentInput)
		// 此处不需要+1
		if len(s.inputs) > 0 {
			s.currentInput = s.currentInput % len(s.inputs)
			s.currentRead = 0
		}
		return rd, nil
	}

	s.err = err
	return rd, err
}

func (s *chunkedJoin) Close() error {
	s.err = io.ErrClosedPipe
	return nil
}

func ChunkedJoin(inputs []io.Reader, chunkSize int) io.ReadCloser {
	return &chunkedJoin{
		inputs:    inputs,
		chunkSize: chunkSize,
	}
}

type bufferedChunkedJoin struct {
	inputs      []io.Reader
	buffer      []byte
	chunkSize   int
	currentRead int
	err         error
}

func (s *bufferedChunkedJoin) Read(buf []byte) (int, error) {
	if s.err != nil {
		return 0, s.err
	}

	if s.currentRead == len(s.buffer) {
		err := sync2.ParallelDo(s.inputs, func(input io.Reader, i int) error {
			bufStart := i * s.chunkSize
			_, err := io.ReadFull(input, s.buffer[bufStart:bufStart+s.chunkSize])
			return err
		})
		if err == io.EOF {
			return 0, io.EOF
		}
		if err != nil {
			return 0, err
		}
		s.currentRead = 0
	}

	n := copy(buf, s.buffer[s.currentRead:])
	s.currentRead += n
	return n, nil
}

func (s *bufferedChunkedJoin) Close() error {
	s.err = io.ErrClosedPipe
	return nil
}

func BufferedChunkedJoin(inputs []io.Reader, chunkSize int) io.ReadCloser {
	buffer := make([]byte, len(inputs)*chunkSize)
	return &bufferedChunkedJoin{
		inputs:      inputs,
		buffer:      buffer,
		chunkSize:   chunkSize,
		currentRead: len(buffer),
	}
}
