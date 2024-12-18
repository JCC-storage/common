package io2

import (
	"hash"
	"io"
)

type ReadHasher struct {
	hasher hash.Hash
	inner  io.Reader
}

func NewReadHasher(h hash.Hash, r io.Reader) *ReadHasher {
	return &ReadHasher{
		hasher: h,
		inner:  r,
	}
}

func (h *ReadHasher) Read(p []byte) (n int, err error) {
	n, err = h.inner.Read(p)
	if n > 0 {
		h.hasher.Write(p[:n])
	}
	return
}

func (h *ReadHasher) Sum() []byte {
	return h.hasher.Sum(nil)
}

type WriteHasher struct {
	hasher hash.Hash
	inner  io.Writer
}

func NewWriteHasher(h hash.Hash, w io.Writer) *WriteHasher {
	return &WriteHasher{
		hasher: h,
		inner:  w,
	}
}

func (h *WriteHasher) Write(p []byte) (n int, err error) {
	n, err = h.inner.Write(p)
	if n > 0 {
		h.hasher.Write(p[:n])
	}
	return
}

func (h *WriteHasher) Sum() []byte {
	return h.hasher.Sum(nil)
}
