package io2

import "io"

var zeros zeroStream

type zeroStream struct{}

func (s *zeroStream) Read(buf []byte) (int, error) {
	for i := range buf {
		buf[i] = 0
	}

	return len(buf), nil
}

func Zeros() io.Reader {
	return &zeros
}

func Zero(buf []byte) {
	for i := range buf {
		buf[i] = 0
	}
}
