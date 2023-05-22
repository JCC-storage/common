package io

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

type readCloserHook struct {
	readCloser io.ReadCloser
	callback   func(closer io.ReadCloser)
	isBefore   bool // callback调用时机，true则在closer的Close之前调用
}

func (hook *readCloserHook) Read(buf []byte) (n int, err error) {
	return hook.readCloser.Read(buf)
}

func (hook *readCloserHook) Close() error {
	if hook.isBefore {
		hook.callback(hook.readCloser)
	}

	err := hook.readCloser.Close()

	if !hook.isBefore {
		hook.callback(hook.readCloser)
	}
	return err
}

func BeforeReadClosing(closer io.ReadCloser, callback func(closer io.ReadCloser)) io.ReadCloser {
	return &readCloserHook{
		readCloser: closer,
		callback:   callback,
		isBefore:   true,
	}
}

func AfterReadClosed(closer io.ReadCloser, callback func(closer io.ReadCloser)) io.ReadCloser {
	return &readCloserHook{
		readCloser: closer,
		callback:   callback,
		isBefore:   false,
	}
}
