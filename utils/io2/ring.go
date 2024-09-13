package io2

import (
	"io"
	"sync"
)

type RingBuffer2 struct {
	buf            []byte
	src            io.ReadCloser
	err            error
	isReading      bool
	writePos       int // 指向下一次写入的位置，应该是一个空位
	readPos        int // 执行下一次读取的位置，应该是有效数据
	waitReading    *sync.Cond
	waitComsuming  *sync.Cond
	UpstreamName   string
	DownstreamName string
}

func RingBuffer(src io.ReadCloser, size int) io.ReadCloser {
	lk := &sync.Mutex{}
	return &RingBuffer2{
		buf:           make([]byte, size),
		src:           src,
		waitReading:   sync.NewCond(lk),
		waitComsuming: sync.NewCond(lk),
	}
}

func (r *RingBuffer2) Read(p []byte) (n int, err error) {
	r.waitReading.L.Lock()
	if !r.isReading {
		go r.reading()
		r.isReading = true
	}

	for r.writePos == r.readPos {
		if r.err != nil {
			r.waitReading.L.Unlock()
			return 0, r.err
		}

		// startTime := time.Now()
		r.waitReading.Wait()
		// fmt.Printf("%s wait data for %v\n", r.DownstreamName, time.Since(startTime))
	}
	writePos := r.writePos
	readPos := r.readPos
	r.waitReading.L.Unlock()

	if readPos < writePos {
		n = copy(p, r.buf[readPos:writePos])
	} else {
		n = copy(p, r.buf[readPos:])
	}

	r.waitComsuming.L.Lock()
	r.readPos = (r.readPos + n) % len(r.buf)
	r.waitComsuming.L.Unlock()
	r.waitComsuming.Broadcast()

	err = nil
	return
}

func (r *RingBuffer2) Close() error {
	r.src.Close()
	r.waitComsuming.Broadcast()
	r.waitReading.Broadcast()
	return nil
}

func (r *RingBuffer2) reading() {
	defer r.src.Close()

	for {
		r.waitComsuming.L.Lock()
		// writePos不能和readPos重合，因为无法区分缓冲区是已经满了，还是完全是空的
		// 所以writePos最多能到readPos的前一格
		for r.writePos+1 == r.readPos {
			r.waitComsuming.Wait()

			if r.err != nil {
				return
			}
		}
		writePos := r.writePos
		readPos := r.readPos
		r.waitComsuming.L.Unlock()

		var n int
		var err error
		if readPos <= writePos {
			// 同上理，写入数据的时候如果readPos为0，则它的前一格是底层缓冲区的最后一格
			// 那就不能写入到这一格
			if readPos == 0 {
				n, err = r.src.Read(r.buf[writePos : len(r.buf)-1])
			} else {
				n, err = r.src.Read(r.buf[writePos:])
			}
		} else if readPos > writePos {
			n, err = r.src.Read(r.buf[writePos:readPos])
		}

		// 无论成功还是失败，都发送一下信号通知读取端
		r.waitReading.L.Lock()
		r.err = err
		r.writePos = (r.writePos + n) % len(r.buf)
		r.waitReading.L.Unlock()
		r.waitReading.Broadcast()

		if err != nil {
			break
		}
	}
}
