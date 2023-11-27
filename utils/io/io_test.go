package io

import (
	"bytes"
	"io"
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
		str := Length(bytes.NewReader([]byte{1, 2, 3}), 3)
		buf := make([]byte, 9)
		rd, err := io.ReadFull(str, buf)
		So(err, ShouldEqual, io.ErrUnexpectedEOF)
		So(buf[:rd], ShouldResemble, []byte{1, 2, 3})
	})

	Convey("非强制，长度小于设定", t, func() {
		str := Length(bytes.NewReader([]byte{1, 2}), 3)

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
		str := Length(bytes.NewReader([]byte{1, 2, 3, 4}), 3)

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
		str := MustLength(bytes.NewReader([]byte{1, 2}), 3)

		buf := make([]byte, 2)
		_, err := io.ReadFull(str, buf)
		if err == nil {
			_, err = io.ReadFull(str, buf)
		}
		So(err, ShouldEqual, io.ErrUnexpectedEOF)
	})
}
