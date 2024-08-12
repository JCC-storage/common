package io2

import (
	"io"

	"gitlink.org.cn/cloudream/common/utils/math2"
)

type rng struct {
	offset int64
	length *int64
	inner  io.ReadCloser
	err    error
}

func (r *rng) Read(p []byte) (n int, err error) {
	if r.err != nil {
		return 0, r.err
	}

	if r.offset > 0 {
		buf := make([]byte, 1024*16)
		for r.offset > 0 {
			need := math2.Min(r.offset, int64(len(buf)))
			rd, err := r.inner.Read(buf[:need])
			if err != nil {
				r.err = err
				return 0, err
			}
			r.offset -= int64(rd)
		}
	}

	if r.length == nil {
		return r.inner.Read(p)
	}

	need := math2.Min(*r.length, int64(len(p)))
	rd, err := r.inner.Read(p[:need])
	if err != nil {
		r.err = err
		return rd, io.EOF
	}

	*r.length -= int64(rd)
	if *r.length == 0 {
		r.err = io.EOF
	}

	return rd, nil
}

func (r *rng) Close() error {
	r.err = io.ErrClosedPipe
	return r.inner.Close()
}

func NewRange(inner io.ReadCloser, offset int64, length *int64) io.ReadCloser {
	return &rng{
		offset: offset,
		length: length,
		inner:  inner,
	}
}

func Ranged(inner io.ReadCloser, offset int64, length int64) io.ReadCloser {
	return &rng{
		offset: offset,
		length: &length,
		inner:  inner,
	}
}

func Offset(inner io.ReadCloser, offset int64) io.ReadCloser {
	return &rng{
		offset: offset,
		inner:  inner,
	}
}
