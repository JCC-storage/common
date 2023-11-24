package io

import (
	"io"

	"gitlink.org.cn/cloudream/common/utils/lo"
	"gitlink.org.cn/cloudream/common/utils/math"
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

	bufLen := math.Min(math.Min(s.chunkSize, len(buf)), s.chunkSize-s.currentRead)
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
		s.inputs = lo.RemoveAt(s.inputs, s.currentInput)
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
