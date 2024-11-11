package io2

import "io"

type nopWriteCloser struct {
	writer io.Writer
}

func NopWriteCloser(w io.Writer) io.WriteCloser {
	return &nopWriteCloser{writer: w}
}

func (w *nopWriteCloser) Write(p []byte) (n int, err error) {
	return w.writer.Write(p)
}

func (w *nopWriteCloser) Close() error {
	return nil
}

type delegateWriteCloser struct {
	writer io.Writer
	fn     func() error
}

func DelegateWriteCloser(w io.Writer, fn func() error) io.WriteCloser {
	return &delegateWriteCloser{writer: w, fn: fn}
}

func (w *delegateWriteCloser) Write(p []byte) (n int, err error) {
	return w.writer.Write(p)
}

func (w *delegateWriteCloser) Close() error {
	return w.fn()
}

type delegateReadCloser struct {
	reader io.Reader
	fn     func() error
}

func DelegateReadCloser(r io.Reader, fn func() error) io.ReadCloser {
	return &delegateReadCloser{reader: r, fn: fn}
}

func (r *delegateReadCloser) Read(p []byte) (n int, err error) {
	return r.reader.Read(p)
}

func (r *delegateReadCloser) Close() error {
	return r.fn()
}
