package io2

import "io"

type Counter struct {
	inner io.Reader
	count int64
}

func (c *Counter) Read(buf []byte) (n int, err error) {
	n, err = c.inner.Read(buf)
	c.count += int64(n)
	return
}

func (c *Counter) Count() int64 {
	return c.count
}

func NewCounter(inner io.Reader) *Counter {
	return &Counter{inner: inner, count: 0}
}
