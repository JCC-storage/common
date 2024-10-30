package cdsapi

import (
	"encoding/binary"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"gitlink.org.cn/cloudream/common/utils/http2"
	"gitlink.org.cn/cloudream/common/utils/io2"
	"gitlink.org.cn/cloudream/common/utils/math2"
	"gitlink.org.cn/cloudream/common/utils/serder"
)

func MakeIPFSFilePath(fileHash string) string {
	return filepath.Join("ipfs", fileHash)
}

func ParseJSONResponse[TBody any](resp *http.Response) (TBody, error) {
	var ret TBody
	contType := resp.Header.Get("Content-Type")
	if strings.Contains(contType, http2.ContentTypeJSON) {
		var err error
		if ret, err = serder.JSONToObjectStreamEx[TBody](resp.Body); err != nil {
			return ret, fmt.Errorf("parsing response: %w", err)
		}

		return ret, nil
	}

	cont, err := io.ReadAll(resp.Body)
	if err != nil {
		return ret, fmt.Errorf("unknow response content type: %s, status: %d", contType, resp.StatusCode)
	}
	strCont := string(cont)

	return ret, fmt.Errorf("unknow response content type: %s, status: %d, body(prefix): %s", contType, resp.StatusCode, strCont[:math2.Min(len(strCont), 200)])
}

func WriteStream(dst io.Writer, src io.Reader) (int64, error) {
	sent := int64(0)

	buf := make([]byte, 1024*4)
	header := make([]byte, 4)

	for {
		n, err := src.Read(buf)
		if n > 0 {
			binary.LittleEndian.PutUint32(header, uint32(n))
			err := io2.WriteAll(dst, header)
			if err != nil {
				return sent, err
			}

			sent += int64(n)
		}

		if err == io.EOF {
			binary.LittleEndian.PutUint32(header, uint32(0))
			err := io2.WriteAll(dst, header)
			if err != nil {
				return sent, err
			}
			return sent, nil
		}

		if err != nil {
			errData := []byte(err.Error())
			header := make([]byte, 4)
			binary.LittleEndian.PutUint32(header, uint32(-len(errData)))
			// 不管有没有成功
			io2.WriteAll(dst, header)
			io2.WriteAll(dst, errData)
			return sent, err
		}
	}
}

func ReadStream(src io.Reader) io.ReadCloser {
	pr, pw := io.Pipe()

	buf := make([]byte, 1024*4)
	go func() {
		for {
			_, err := io.ReadFull(src, buf[:4])
			if err != nil {
				pw.CloseWithError(err)
				break
			}

			h := int32(binary.LittleEndian.Uint32(buf[:4]))
			if h == 0 {
				pw.Close()
				break
			}

			if h < 0 {
				_, err := io.ReadFull(src, buf[:-h])
				if err != nil {
					pw.CloseWithError(err)
					break
				}

				pw.CloseWithError(fmt.Errorf(string(buf[:-h])))
				break
			}

			_, err = io.ReadFull(src, buf[:h])
			if err != nil {
				pw.CloseWithError(err)
				break
			}

			_, err = pw.Write(buf[:h])
			if err != nil {
				pw.Close()
				break
			}
		}
	}()

	return pr
}
