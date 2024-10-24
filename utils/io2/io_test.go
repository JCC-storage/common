package io2

import (
	"bytes"
	"io"
	"sync"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_Join(t *testing.T) {
	Convey("连接多个流", t, func() {
		str := Join([]io.Reader{
			bytes.NewReader([]byte{1, 2, 3}),
			bytes.NewReader([]byte{4}),
			bytes.NewReader([]byte{5, 6, 7, 8}),
		})

		buf := make([]byte, 9)
		rd, err := io.ReadFull(str, buf)

		So(err, ShouldEqual, io.ErrUnexpectedEOF)
		So(buf[:rd], ShouldResemble, []byte{1, 2, 3, 4, 5, 6, 7, 8})
	})

	Convey("分块式连接多个流，每个流长度相等", t, func() {
		str := ChunkedJoin([]io.Reader{
			bytes.NewReader([]byte{1, 2, 3}),
			bytes.NewReader([]byte{4, 5, 6}),
			bytes.NewReader([]byte{7, 8, 9}),
		}, 3)

		buf := make([]byte, 10)
		rd, err := io.ReadFull(str, buf)

		So(err, ShouldEqual, io.ErrUnexpectedEOF)
		So(buf[:rd], ShouldResemble, []byte{1, 2, 3, 4, 5, 6, 7, 8, 9})
	})

	Convey("分块式连接多个流，流长度不相等，但都是chunkSize的整数倍", t, func() {
		str := ChunkedJoin([]io.Reader{
			bytes.NewReader([]byte{1, 2, 3}),
			bytes.NewReader([]byte{4, 5, 6, 7, 8, 9, 10, 11, 12}),
			bytes.NewReader([]byte{}),
			bytes.NewReader([]byte{13, 14, 15}),
		}, 3)

		buf := make([]byte, 100)
		rd, err := io.ReadFull(str, buf)

		So(err, ShouldEqual, io.ErrUnexpectedEOF)
		So(buf[:rd], ShouldResemble, []byte{1, 2, 3, 4, 5, 6, 13, 14, 15, 7, 8, 9, 10, 11, 12})
	})

	Convey("分块式连接多个流，流长度不相等，且不一定是chunkSize的整数倍", t, func() {
		str := ChunkedJoin([]io.Reader{
			bytes.NewReader([]byte{1, 2, 3}),
			bytes.NewReader([]byte{4, 5, 6, 7, 8}),
			bytes.NewReader([]byte{9}),
		}, 3)

		buf := make([]byte, 10)
		rd, err := io.ReadFull(str, buf)

		So(err, ShouldEqual, io.ErrUnexpectedEOF)
		So(buf[:rd], ShouldResemble, []byte{1, 2, 3, 4, 5, 6, 9, 7, 8})
	})
}

func Test_Length(t *testing.T) {
	Convey("非强制，长度刚好", t, func() {
		str := Length(io.NopCloser(bytes.NewReader([]byte{1, 2, 3})), 3)
		buf := make([]byte, 9)
		rd, err := io.ReadFull(str, buf)
		So(err, ShouldEqual, io.ErrUnexpectedEOF)
		So(buf[:rd], ShouldResemble, []byte{1, 2, 3})
	})

	Convey("非强制，长度小于设定", t, func() {
		str := Length(io.NopCloser(bytes.NewBuffer([]byte{1, 2})), 3)

		buf := make([]byte, 2)
		rd, err := io.ReadFull(str, buf)
		if err == nil {
			var rd2 int
			rd2, err = io.ReadFull(str, buf)
			So(rd2, ShouldEqual, 0)
		}
		So(err, ShouldEqual, io.EOF)
		So(buf[:rd], ShouldResemble, []byte{1, 2})
	})

	Convey("非强制，长度大于设定", t, func() {
		str := Length(io.NopCloser(bytes.NewReader([]byte{1, 2, 3, 4})), 3)

		buf := make([]byte, 3)
		rd, err := io.ReadFull(str, buf)
		if err == nil {
			var rd2 int
			rd2, err = io.ReadFull(str, buf)
			So(rd2, ShouldEqual, 0)
		}
		So(err, ShouldEqual, io.EOF)
		So(buf[:rd], ShouldResemble, []byte{1, 2, 3})
	})

	Convey("强制，长度小于设定", t, func() {
		str := MustLength(io.NopCloser(bytes.NewReader([]byte{1, 2})), 3)

		buf := make([]byte, 2)
		_, err := io.ReadFull(str, buf)
		if err == nil {
			_, err = io.ReadFull(str, buf)
		}
		So(err, ShouldEqual, io.ErrUnexpectedEOF)
	})
}

func Test_Clone(t *testing.T) {
	Convey("所有输出流都会被读取完", t, func() {
		data := []byte{1, 2, 3, 4, 5}
		str := bytes.NewReader(data)

		cloneds := Clone(str, 3)
		reads := make([][]byte, 3)
		errs := make([]error, 3)

		wg := sync.WaitGroup{}
		wg.Add(3)

		go func() {
			reads[0], errs[0] = io.ReadAll(cloneds[0])
			wg.Done()
		}()
		go func() {
			reads[1], errs[1] = io.ReadAll(cloneds[1])
			wg.Done()
		}()
		go func() {
			reads[2], errs[2] = io.ReadAll(cloneds[2])
			wg.Done()
		}()

		wg.Wait()

		So(reads, ShouldResemble, [][]byte{data, data, data})
		So(errs, ShouldResemble, []error{nil, nil, nil})
	})

	Convey("其中一个流读到一半就停止读取", t, func() {
		data := []byte{1, 2, 3, 4, 5}
		str := bytes.NewReader(data)

		cloneds := Clone(str, 3)
		reads := make([][]byte, 3)
		errs := make([]error, 3)

		wg := sync.WaitGroup{}
		wg.Add(3)

		go func() {
			reads[0], errs[0] = io.ReadAll(cloneds[0])
			wg.Done()
		}()
		go func() {
			buf := make([]byte, 3)
			_, errs[1] = io.ReadFull(cloneds[1], buf)
			reads[1] = buf
			cloneds[1].Close()
			wg.Done()
		}()
		go func() {
			reads[2], errs[2] = io.ReadAll(cloneds[2])
			wg.Done()
		}()

		wg.Wait()

		So(reads, ShouldResemble, [][]byte{data, {1, 2, 3}, data})
		So(errs, ShouldResemble, []error{nil, nil, nil})
	})
}
