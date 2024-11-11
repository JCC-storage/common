package http2

import (
	"encoding/binary"
	"fmt"
	"io"

	"gitlink.org.cn/cloudream/common/utils/io2"
	"gitlink.org.cn/cloudream/common/utils/math2"
)

const (
	PartTypeError   = 0xff
	PartTypeEOF     = 0x00
	PartTypeNewPart = 0x01
	PartTypeData    = 0x02
)

type ChunkedWriter struct {
	stream io.WriteCloser
}

func NewChunkedWriter(stream io.WriteCloser) *ChunkedWriter {
	return &ChunkedWriter{stream: stream}
}

// 开始写入一个新Part。每次只能有一个Part在写入。
func (w *ChunkedWriter) BeginPart(name string) io.Writer {
	header := []byte{PartTypeNewPart, 0, 0}
	binary.LittleEndian.PutUint16(header[1:], uint16(len(name)))

	err := io2.WriteAll(w.stream, header)
	if err != nil {
		return io2.ErrorWriter(fmt.Errorf("write header: %w", err))
	}

	err = io2.WriteAll(w.stream, []byte(name))
	if err != nil {
		return io2.ErrorWriter(fmt.Errorf("write part name: %w", err))
	}

	return &PartWriter{stream: w.stream}
}

func (w *ChunkedWriter) WriteDataPart(name string, data []byte) error {
	pw := w.BeginPart(name)
	return io2.WriteAll(pw, data)
}

func (w *ChunkedWriter) WriteStreamPart(name string, stream io.Reader) (int64, error) {
	pw := w.BeginPart(name)
	n, err := io.Copy(pw, stream)
	return n, err
}

// 发送ErrorPart并关闭连接。无论是否返回错误，连接都会关闭
func (w *ChunkedWriter) Abort(msg string) error {
	defer w.stream.Close()

	header := []byte{PartTypeError, 0, 0}
	binary.LittleEndian.PutUint16(header[1:], uint16(len(msg)))

	err := io2.WriteAll(w.stream, header)
	if err != nil {
		return fmt.Errorf("write header: %w", err)
	}

	err = io2.WriteAll(w.stream, []byte(msg))
	if err != nil {
		return fmt.Errorf("write error message: %w", err)
	}

	return nil
}

// 发送EOFPart并关闭连接。无论是否返回错误，连接都会关闭
func (w *ChunkedWriter) Finish() error {
	defer w.stream.Close()

	header := []byte{PartTypeEOF, 0, 0}
	err := io2.WriteAll(w.stream, header)
	if err != nil {
		return fmt.Errorf("write header: %w", err)
	}

	return nil
}

// 直接关闭连接，不发送EOFPart也不发送ErrorPart。接收端会产生一个io.UnexpectedEOF错误
func (w *ChunkedWriter) Close() {
	w.stream.Close()
}

type PartWriter struct {
	stream io.WriteCloser
}

func (w *PartWriter) Write(data []byte) (int, error) {
	sendLen := math2.Min(len(data), 0xffff)

	header := []byte{PartTypeData, 0, 0}
	binary.LittleEndian.PutUint16(header[1:], uint16(sendLen))

	err := io2.WriteAll(w.stream, header)
	if err != nil {
		return 0, fmt.Errorf("write header: %w", err)
	}

	err = io2.WriteAll(w.stream, data[:sendLen])
	if err != nil {
		return 0, fmt.Errorf("write data: %w", err)
	}

	return sendLen, nil
}

type ChunkedAbortError struct {
	Message string
}

func (e *ChunkedAbortError) Error() string {
	return e.Message
}

type ChunkedReader struct {
	stream     io.ReadCloser
	partHeader []byte
	err        error
}

func NewChunkedReader(stream io.ReadCloser) *ChunkedReader {
	return &ChunkedReader{stream: stream}
}

// 读取下一个Part。每次只能读取一个Part，且必须将其全部读取完毕才能读取下一个
func (r *ChunkedReader) NextPart() (string, io.Reader, error) {
	if r.err != nil {
		return "", nil, r.err
	}

	if r.partHeader == nil {
		r.partHeader = make([]byte, 3)
		_, err := io.ReadFull(r.stream, r.partHeader)
		if err != nil {
			r.err = fmt.Errorf("read header: %w", err)
			return "", nil, r.err
		}
	}

	partType := r.partHeader[0]
	switch partType {
	case PartTypeNewPart:
		partNameLen := int(binary.LittleEndian.Uint16(r.partHeader[1:]))
		partName := make([]byte, partNameLen)

		_, err := io.ReadFull(r.stream, partName)
		if err != nil {
			r.err = fmt.Errorf("read part name: %w", err)
			return "", nil, r.err
		}

		return string(partName), &PartReader{creader: r}, nil

	case PartTypeData:
		r.err = fmt.Errorf("unexpected data part")
		return "", nil, r.err

	case PartTypeEOF:
		r.err = io.EOF
		return "", nil, r.err

	case PartTypeError:
		msgLen := int(binary.LittleEndian.Uint16(r.partHeader[1:]))
		msg := make([]byte, msgLen)

		_, err := io.ReadFull(r.stream, msg)
		if err != nil {
			r.err = fmt.Errorf("read error message: %w", err)
			return "", nil, r.err
		}

		r.err = &ChunkedAbortError{Message: string(msg)}
		return "", nil, r.err

	default:
		r.err = fmt.Errorf("unknown part type: %d", partType)
		return "", nil, r.err
	}
}

func (r *ChunkedReader) NextDataPart() (string, []byte, error) {
	partName, partReader, err := r.NextPart()
	if err != nil {
		return "", nil, err
	}

	data, err := io.ReadAll(partReader)
	if err != nil {
		return "", nil, err
	}

	return partName, data, nil
}

func (r *ChunkedReader) Close() {
	r.stream.Close()
}

type PartReader struct {
	creader     *ChunkedReader
	partLen     int
	partReadLen int
}

func (r *PartReader) Read(p []byte) (int, error) {
	// 允许有空的DataPart，因此用循环来跳过空的Part
	for r.partLen-r.partReadLen == 0 {
		header := make([]byte, 3)
		_, err := io.ReadFull(r.creader.stream, header)
		if err != nil {
			r.creader.err = err
			return 0, err
		}

		partType := header[0]
		switch partType {
		case PartTypeNewPart:
			r.creader.partHeader = header
			return 0, io.EOF

		case PartTypeData:
			r.partLen = int(binary.LittleEndian.Uint16(header[1:]))
			r.partReadLen = 0

		case PartTypeEOF:
			r.creader.err = io.EOF
			return 0, io.EOF

		case PartTypeError:
			msgLen := int(binary.LittleEndian.Uint16(header[1:]))
			msg := make([]byte, msgLen)

			_, err := io.ReadFull(r.creader.stream, msg)
			if err != nil {
				r.creader.err = fmt.Errorf("read error message: %w", err)
				return 0, fmt.Errorf("read error message: %w", err)
			}

			r.creader.err = &ChunkedAbortError{Message: string(msg)}
			return 0, r.creader.err
		}
	}

	readLen := math2.Min(len(p), r.partLen-r.partReadLen)
	n, err := r.creader.stream.Read(p[:readLen])
	if err == io.EOF {
		r.creader.err = io.ErrUnexpectedEOF
		return 0, io.ErrUnexpectedEOF
	}
	if err != nil {
		r.creader.err = err
		return 0, err
	}

	r.partReadLen += n
	return n, nil
}
