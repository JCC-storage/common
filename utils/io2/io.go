package io2

import "io"

type PromiseWriteCloser[T any] interface {
	io.Writer
	Abort(err error)    // 中断发送文件
	Finish() (T, error) // 发送文件完成，等待返回结果
}

func WriteAll(writer io.Writer, data []byte) error {
	pos := 0
	dataLen := len(data)

	for pos < dataLen {
		writeLen, err := writer.Write(data[pos:])
		if err != nil {
			return err
		}

		pos += writeLen
	}

	return nil
}

const (
	onceDisabled = 0
	onceEnabled  = 1
	onceDone     = 2
)

type readCloserHook struct {
	readCloser io.ReadCloser
	callback   func(closer io.ReadCloser)
	once       int
	isBefore   bool // callback调用时机，true则在closer的Close之前调用
}

func (hook *readCloserHook) Read(buf []byte) (n int, err error) {
	return hook.readCloser.Read(buf)
}

func (hook *readCloserHook) Close() error {
	if hook.once == onceDone {
		return hook.readCloser.Close()
	}

	if hook.isBefore {
		hook.callback(hook.readCloser)
	}

	err := hook.readCloser.Close()

	if !hook.isBefore {
		hook.callback(hook.readCloser)
	}

	if hook.once == onceEnabled {
		hook.once = onceDone
	}

	return err
}

func BeforeReadClosing(closer io.ReadCloser, callback func(closer io.ReadCloser)) io.ReadCloser {
	return &readCloserHook{
		readCloser: closer,
		callback:   callback,
		once:       onceDisabled,
		isBefore:   true,
	}
}

func AfterReadClosed(closer io.ReadCloser, callback func(closer io.ReadCloser)) io.ReadCloser {
	return &readCloserHook{
		readCloser: closer,
		callback:   callback,
		once:       onceDisabled,
		isBefore:   false,
	}
}

func AfterReadClosedOnce(closer io.ReadCloser, callback func(closer io.ReadCloser)) io.ReadCloser {
	return &readCloserHook{
		readCloser: closer,
		callback:   callback,
		once:       onceEnabled,
		isBefore:   false,
	}
}

type afterEOF struct {
	inner    io.ReadCloser
	callback func(str io.ReadCloser, err error)
}

func (hook *afterEOF) Read(buf []byte) (n int, err error) {
	n, err = hook.inner.Read(buf)
	if hook.callback != nil {
		if err == io.EOF {
			hook.callback(hook.inner, nil)
			hook.callback = nil
		} else if err != nil {
			hook.callback(hook.inner, err)
			hook.callback = nil
		}
	}
	return n, err
}

func (hook *afterEOF) Close() error {
	err := hook.inner.Close()
	if hook.callback != nil {
		hook.callback(hook.inner, io.ErrClosedPipe)
		hook.callback = nil
	}
	return err
}

func AfterEOF(str io.ReadCloser, callback func(str io.ReadCloser, err error)) io.ReadCloser {
	return &afterEOF{
		inner:    str,
		callback: callback,
	}
}

type readerWithCloser struct {
	reader io.Reader
	closer func(reader io.Reader) error
}

func (hook *readerWithCloser) Read(buf []byte) (n int, err error) {
	return hook.reader.Read(buf)
}
func (c *readerWithCloser) Close() error {
	return c.closer(c.reader)
}

func WithCloser(reader io.Reader, closer func(reader io.Reader) error) io.ReadCloser {
	return &readerWithCloser{
		reader: reader,
		closer: closer,
	}
}

type LazyReadCloser struct {
	open   func() (io.ReadCloser, error)
	stream io.ReadCloser
}

func (r *LazyReadCloser) Read(buf []byte) (n int, err error) {
	if r.stream == nil {
		var err error
		r.stream, err = r.open()
		if err != nil {
			return 0, err
		}
	}

	return r.stream.Read(buf)
}

func (r *LazyReadCloser) Close() error {
	if r.stream == nil {
		return nil
	}

	return r.stream.Close()
}

func Lazy(open func() (io.ReadCloser, error)) *LazyReadCloser {
	return &LazyReadCloser{
		open: open,
	}
}

func ToReaders(strs []io.ReadCloser) ([]io.Reader, func()) {
	var readers []io.Reader
	for _, s := range strs {
		readers = append(readers, s)
	}

	return readers, func() {
		for _, s := range strs {
			s.Close()
		}
	}
}

func DropWithBuf(str io.Reader, buf []byte) {
	for {
		_, err := str.Read(buf)
		if err != nil {
			break
		}
	}
}
