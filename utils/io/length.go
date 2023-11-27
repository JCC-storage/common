package io

import (
	"io"

	"gitlink.org.cn/cloudream/common/utils/math"
)

type lengthStream struct {
	src        io.Reader
	length     int64
	readLength int64
	must       bool
	err        error
}

func (s *lengthStream) Read(buf []byte) (int, error) {
	if s.err != nil {
		return 0, s.err
	}

	bufLen := math.Min(s.length-s.readLength, int64(len(buf)))
	rd, err := s.src.Read(buf[:bufLen])
	if err == nil {
		s.readLength += int64(rd)
		if s.readLength == s.length {
			s.err = io.EOF
		}
		return rd, nil
	}

	if err == io.EOF {
		s.readLength += int64(rd)
		if s.readLength < s.length && s.must {
			s.err = io.ErrUnexpectedEOF
			return rd, io.ErrUnexpectedEOF
		}

		s.err = io.EOF
		return rd, io.EOF
	}

	s.err = err
	return 0, err
}

func (s *lengthStream) Close() error {
	s.err = io.ErrClosedPipe
	return nil
}

func Length(str io.Reader, length int64) io.ReadCloser {
	return &lengthStream{
		src:    str,
		length: length,
	}
}

func MustLength(str io.Reader, length int64) io.ReadCloser {
	return &lengthStream{
		src:    str,
		length: length,
		must:   true,
	}
}
