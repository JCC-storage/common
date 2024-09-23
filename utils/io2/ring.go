package io2

import (
	"io"
	"sync"
	"time"

	"gitlink.org.cn/cloudream/common/utils/math2"
)

type RingBufferStats struct {
	MaxWaitDataTime        time.Duration // 外部读取数据时的最长等待时间
	MaxWaitFreeSpaceTime   time.Duration // 从数据源读取数据之前，等待空闲空间的最长时间
	TotalWaitDataTime      time.Duration // 总等待读取数据的时间
	TotalWaitFreeSpaceTime time.Duration // 总等待空闲空间的时间
}

type RingBuffer struct {
	buf           []byte
	src           io.ReadCloser
	maxPerRead    int // 后台读取线程每次读取的最大字节数，太小会导致IO次数增多，太大会导致读、写并行性下降
	err           error
	isReading     bool
	writePos      int // 指向下一次写入的位置，应该是一个空位
	readPos       int // 执行下一次读取的位置，应该是有效数据
	waitReading   *sync.Cond
	waitComsuming *sync.Cond
	stats         RingBufferStats
}

func Ring(src io.ReadCloser, size int) *RingBuffer {
	lk := &sync.Mutex{}
	return &RingBuffer{
		buf:           make([]byte, size),
		src:           src,
		maxPerRead:    size / 4,
		waitReading:   sync.NewCond(lk),
		waitComsuming: sync.NewCond(lk),
	}
}

func (r *RingBuffer) Stats() RingBufferStats {
	return r.stats
}

func (r *RingBuffer) Read(p []byte) (n int, err error) {
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

		startTime := time.Now()
		r.waitReading.Wait()
		dt := time.Since(startTime)
		r.stats.MaxWaitDataTime = math2.Max(r.stats.MaxWaitDataTime, dt)
		r.stats.TotalWaitDataTime += dt
	}
	writePos := r.writePos
	readPos := r.readPos
	r.waitReading.L.Unlock()

	if readPos < writePos {
		maxRead := math2.Min(r.maxPerRead, writePos-readPos)
		n = copy(p, r.buf[readPos:readPos+maxRead])
	} else {
		maxRead := math2.Min(r.maxPerRead, len(r.buf)-readPos)
		n = copy(p, r.buf[readPos:readPos+maxRead])
	}

	r.waitComsuming.L.Lock()
	r.readPos = (r.readPos + n) % len(r.buf)
	r.waitComsuming.L.Unlock()
	r.waitComsuming.Broadcast()

	err = nil
	return
}

func (r *RingBuffer) Close() error {
	r.src.Close()
	r.waitComsuming.L.Lock()
	r.err = io.ErrClosedPipe
	r.waitComsuming.L.Unlock()
	r.waitComsuming.Broadcast()
	r.waitReading.Broadcast()
	return nil
}

func (r *RingBuffer) reading() {
	defer r.src.Close()

	for {
		r.waitComsuming.L.Lock()
		// writePos不能和readPos重合，因为无法区分缓冲区是已经满了，还是完全是空的
		// 所以writePos最多能到readPos的前一格
		for (r.writePos+1)%len(r.buf) == r.readPos {
			startTime := time.Now()
			r.waitComsuming.Wait()
			dt := time.Since(startTime)
			r.stats.MaxWaitFreeSpaceTime = math2.Max(r.stats.MaxWaitFreeSpaceTime, dt)
			r.stats.TotalWaitFreeSpaceTime += dt

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
				maxWrite := math2.Min(r.maxPerRead, len(r.buf)-1-writePos)
				n, err = r.src.Read(r.buf[writePos : writePos+maxWrite])
			} else {
				maxWrite := math2.Min(r.maxPerRead, len(r.buf)-writePos)
				n, err = r.src.Read(r.buf[writePos : writePos+maxWrite])
			}
		} else {
			maxWrite := math2.Min(r.maxPerRead, readPos-1-writePos)
			n, err = r.src.Read(r.buf[writePos : writePos+maxWrite])
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
