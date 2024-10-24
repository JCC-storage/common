package io2

import (
	"bytes"
	"io"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"gitlink.org.cn/cloudream/common/utils/sync2"
)

type syncReader struct {
	data         [][]byte
	curDataIndex int
	nextData     int
	counter      *sync2.CounterCond
}

func (r *syncReader) Read(p []byte) (n int, err error) {
	if r.nextData >= len(r.data) {
		return 0, io.EOF
	}

	if r.data[r.nextData] == nil {
		r.counter.Wait()
		r.nextData++
	}

	n = copy(p, r.data[r.nextData][r.curDataIndex:])
	r.curDataIndex += n
	if r.curDataIndex == len(r.data[r.nextData]) {
		r.curDataIndex = 0
		r.nextData++
	}
	return n, nil
}

func (r *syncReader) Close() error {
	return nil
}

func Test_RingBuffer(t *testing.T) {
	Convey("写满读满", t, func() {
		b := Ring(io.NopCloser(bytes.NewBuffer([]byte{1, 2, 3})), 4)

		ret := make([]byte, 3)
		n, err := b.Read(ret)
		So(err, ShouldEqual, nil)
		So(n, ShouldEqual, 3)
		So(ret, ShouldResemble, []byte{1, 2, 3})
	})

	Convey("1+3+1", t, func() {
		sy := sync2.NewCounterCond(0)

		b := Ring(&syncReader{
			data: [][]byte{
				{1},
				nil,
				{2, 3, 4, 5},
			},
			counter: sy,
		}, 4)

		ret := make([]byte, 3)
		n, err := b.Read(ret)
		So(err, ShouldEqual, nil)
		So(n, ShouldEqual, 1)
		So(ret[:n], ShouldResemble, []byte{1})

		sy.Release()

		n, err = b.Read(ret)
		So(err, ShouldEqual, nil)
		So(n, ShouldEqual, 3)
		So(ret[:n], ShouldResemble, []byte{2, 3, 4})

		n, err = b.Read(ret)
		So(err, ShouldEqual, nil)
		So(n, ShouldEqual, 1)
		So(ret[:n], ShouldResemble, []byte{5})
	})

	Convey("3+1+2", t, func() {
		sy := sync2.NewCounterCond(0)

		b := Ring(&syncReader{
			data: [][]byte{
				{1, 2, 3, 4, 5, 6},
			},
			counter: sy,
		}, 4)

		ret := make([]byte, 3)
		n, err := b.Read(ret)
		So(err, ShouldEqual, nil)
		So(n, ShouldEqual, 3)
		So(ret[:n], ShouldResemble, []byte{1, 2, 3})

		n, err = b.Read(ret)
		So(err, ShouldEqual, nil)
		So(n, ShouldEqual, 1)
		So(ret[:n], ShouldResemble, []byte{4})

		n, err = b.Read(ret)
		So(err, ShouldEqual, nil)
		So(n, ShouldEqual, 2)
		So(ret[:n], ShouldResemble, []byte{5, 6})
	})
}
