package io2

import (
	"bytes"
	"io"
	"sync"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_RoundRobin(t *testing.T) {
	Convey("数据长度为chunkSize的整数倍", t, func() {
		input := []byte{
			1, 2, 3, 4, 5, 6, 7, 8, 9,
			1, 2, 3, 4, 5, 6, 7, 8, 9,
		}

		outputs := ChunkedSplit(bytes.NewReader(input), 3, 3)

		wg := sync.WaitGroup{}
		wg.Add(3)

		o1 := make([]byte, 10)
		var e1 error
		var rd1 int
		go func() {
			rd1, e1 = io.ReadFull(outputs[0], o1)
			wg.Done()
		}()

		o2 := make([]byte, 10)
		var e2 error
		var rd2 int
		go func() {
			rd2, e2 = io.ReadFull(outputs[1], o2)
			wg.Done()
		}()

		o3 := make([]byte, 10)
		var e3 error
		var rd3 int
		go func() {
			rd3, e3 = io.ReadFull(outputs[2], o3)
			wg.Done()
		}()

		wg.Wait()

		So(e1, ShouldEqual, io.ErrUnexpectedEOF)
		So(o1[:rd1], ShouldResemble, []byte{1, 2, 3, 1, 2, 3})

		So(e2, ShouldEqual, io.ErrUnexpectedEOF)
		So(o2[:rd2], ShouldResemble, []byte{4, 5, 6, 4, 5, 6})

		So(e3, ShouldEqual, io.ErrUnexpectedEOF)
		So(o3[:rd3], ShouldResemble, []byte{7, 8, 9, 7, 8, 9})
	})

	Convey("数据长度比chunkSize的整数倍少小于chunkSize的数量，且不填充0", t, func() {
		input := []byte{
			1, 2, 3, 4, 5, 6, 7, 8, 9,
			1, 2, 3, 4, 5, 6, 7,
		}

		outputs := ChunkedSplit(bytes.NewReader(input), 3, 3)

		wg := sync.WaitGroup{}
		wg.Add(3)

		o1 := make([]byte, 10)
		var e1 error
		var rd1 int
		go func() {
			rd1, e1 = io.ReadFull(outputs[0], o1)
			wg.Done()
		}()

		o2 := make([]byte, 10)
		var e2 error
		var rd2 int
		go func() {
			rd2, e2 = io.ReadFull(outputs[1], o2)
			wg.Done()
		}()

		o3 := make([]byte, 10)
		var e3 error
		var rd3 int
		go func() {
			rd3, e3 = io.ReadFull(outputs[2], o3)
			wg.Done()
		}()

		wg.Wait()

		So(e1, ShouldEqual, io.ErrUnexpectedEOF)
		So(o1[:rd1], ShouldResemble, []byte{1, 2, 3, 1, 2, 3})

		So(e2, ShouldEqual, io.ErrUnexpectedEOF)
		So(o2[:rd2], ShouldResemble, []byte{4, 5, 6, 4, 5, 6})

		So(e3, ShouldEqual, io.ErrUnexpectedEOF)
		So(o3[:rd3], ShouldResemble, []byte{7, 8, 9, 7})
	})

	Convey("数据长度比chunkSize的整数倍少多于chunkSize的数量，且不填充0", t, func() {
		input := []byte{
			1, 2, 3, 4, 5, 6, 7, 8, 9,
			1, 2, 3, 4, 5,
		}

		outputs := ChunkedSplit(bytes.NewReader(input), 3, 3)

		wg := sync.WaitGroup{}
		wg.Add(3)

		o1 := make([]byte, 10)
		var e1 error
		var rd1 int
		go func() {
			rd1, e1 = io.ReadFull(outputs[0], o1)
			wg.Done()
		}()

		o2 := make([]byte, 10)
		var e2 error
		var rd2 int
		go func() {
			rd2, e2 = io.ReadFull(outputs[1], o2)
			wg.Done()
		}()

		o3 := make([]byte, 10)
		var e3 error
		var rd3 int
		go func() {
			rd3, e3 = io.ReadFull(outputs[2], o3)
			wg.Done()
		}()

		wg.Wait()

		So(e1, ShouldEqual, io.ErrUnexpectedEOF)
		So(o1[:rd1], ShouldResemble, []byte{1, 2, 3, 1, 2, 3})

		So(e2, ShouldEqual, io.ErrUnexpectedEOF)
		So(o2[:rd2], ShouldResemble, []byte{4, 5, 6, 4, 5})

		So(e3, ShouldEqual, io.ErrUnexpectedEOF)
		So(o3[:rd3], ShouldResemble, []byte{7, 8, 9})
	})

	Convey("数据长度是chunkSize的整数倍，且填充0", t, func() {
		input := []byte{
			1, 2, 3, 4, 5, 6, 7, 8, 9,
		}

		outputs := ChunkedSplit(bytes.NewReader(input), 3, 3, ChunkedSplitOption{
			PaddingZeros: true,
		})

		wg := sync.WaitGroup{}
		wg.Add(3)

		o1 := make([]byte, 10)
		var e1 error
		var rd1 int
		go func() {
			rd1, e1 = io.ReadFull(outputs[0], o1)
			wg.Done()
		}()

		o2 := make([]byte, 10)
		var e2 error
		var rd2 int
		go func() {
			rd2, e2 = io.ReadFull(outputs[1], o2)
			wg.Done()
		}()

		o3 := make([]byte, 10)
		var e3 error
		var rd3 int
		go func() {
			rd3, e3 = io.ReadFull(outputs[2], o3)
			wg.Done()
		}()

		wg.Wait()

		So(e1, ShouldEqual, io.ErrUnexpectedEOF)
		So(o1[:rd1], ShouldResemble, []byte{1, 2, 3})

		So(e2, ShouldEqual, io.ErrUnexpectedEOF)
		So(o2[:rd2], ShouldResemble, []byte{4, 5, 6})

		So(e3, ShouldEqual, io.ErrUnexpectedEOF)
		So(o3[:rd3], ShouldResemble, []byte{7, 8, 9})
	})

	Convey("数据长度比chunkSize的整数倍少小于chunkSize的数量，但是填充0", t, func() {
		input := []byte{
			1, 2, 3, 4, 5, 6, 7, 8, 9,
			1, 2, 3, 4, 5, 6, 7,
		}

		outputs := ChunkedSplit(bytes.NewReader(input), 3, 3, ChunkedSplitOption{
			PaddingZeros: true,
		})
		wg := sync.WaitGroup{}
		wg.Add(3)

		o1 := make([]byte, 10)
		var e1 error
		var rd1 int
		go func() {
			rd1, e1 = io.ReadFull(outputs[0], o1)
			wg.Done()
		}()

		o2 := make([]byte, 10)
		var e2 error
		var rd2 int
		go func() {
			rd2, e2 = io.ReadFull(outputs[1], o2)
			wg.Done()
		}()

		o3 := make([]byte, 10)
		var e3 error
		var rd3 int
		go func() {
			rd3, e3 = io.ReadFull(outputs[2], o3)
			wg.Done()
		}()

		wg.Wait()

		So(e1, ShouldEqual, io.ErrUnexpectedEOF)
		So(o1[:rd1], ShouldResemble, []byte{1, 2, 3, 1, 2, 3})

		So(e2, ShouldEqual, io.ErrUnexpectedEOF)
		So(o2[:rd2], ShouldResemble, []byte{4, 5, 6, 4, 5, 6})

		So(e3, ShouldEqual, io.ErrUnexpectedEOF)
		So(o3[:rd3], ShouldResemble, []byte{7, 8, 9, 7, 0, 0})
	})

	Convey("数据长度比chunkSize的整数倍少多于chunkSize的数量，但是填充0", t, func() {
		input := []byte{
			1, 2, 3, 4, 5, 6, 7, 8, 9,
			1, 2,
		}

		outputs := ChunkedSplit(bytes.NewReader(input), 3, 3, ChunkedSplitOption{
			PaddingZeros: true,
		})

		wg := sync.WaitGroup{}
		wg.Add(3)

		o1 := make([]byte, 10)
		var e1 error
		var rd1 int
		go func() {
			rd1, e1 = io.ReadFull(outputs[0], o1)
			wg.Done()
		}()

		o2 := make([]byte, 10)
		var e2 error
		var rd2 int
		go func() {
			rd2, e2 = io.ReadFull(outputs[1], o2)
			wg.Done()
		}()

		o3 := make([]byte, 10)
		var e3 error
		var rd3 int
		go func() {
			rd3, e3 = io.ReadFull(outputs[2], o3)
			wg.Done()
		}()

		wg.Wait()

		So(e1, ShouldEqual, io.ErrUnexpectedEOF)
		So(o1[:rd1], ShouldResemble, []byte{1, 2, 3, 1, 2, 0})

		So(e2, ShouldEqual, io.ErrUnexpectedEOF)
		So(o2[:rd2], ShouldResemble, []byte{4, 5, 6, 0, 0, 0})

		So(e3, ShouldEqual, io.ErrUnexpectedEOF)
		So(o3[:rd3], ShouldResemble, []byte{7, 8, 9, 0, 0, 0})
	})
}
