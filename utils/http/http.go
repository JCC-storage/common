package http

import (
	"bytes"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	ul "net/url"
	"reflect"
	"strings"

	"github.com/mitchellh/mapstructure"
	"gitlink.org.cn/cloudream/common/pkgs/iterator"
	"gitlink.org.cn/cloudream/common/utils/io2"
	"gitlink.org.cn/cloudream/common/utils/math2"
	"gitlink.org.cn/cloudream/common/utils/serder"
)

const (
	ContentTypeJSON        = "application/json"
	ContentTypeForm        = "application/x-www-form-urlencoded"
	ContentTypeMultiPart   = "multipart/form-data"
	ContentTypeOctetStream = "application/octet-stream"
	ContentTypeEventStream = "text/event-stream"
)

var defaultClient = http.DefaultClient

type RequestParam struct {
	Header any
	Query  any
	Body   any
}

func GetJSON(url string, param RequestParam) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	if err = prepareQuery(req, param.Query); err != nil {
		return nil, err
	}

	if err = prepareHeader(req, param.Header); err != nil {
		return nil, err
	}

	if err = prepareJSONBody(req, param.Body); err != nil {
		return nil, err
	}

	return defaultClient.Do(req)
}

func GetForm(url string, param RequestParam) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	if err = prepareQuery(req, param.Query); err != nil {
		return nil, err
	}

	if err = prepareHeader(req, param.Header); err != nil {
		return nil, err
	}

	if err = prepareFormBody(req, param.Body); err != nil {
		return nil, err
	}

	return defaultClient.Do(req)
}

func PostJSON(url string, param RequestParam) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return nil, err
	}

	if err = prepareQuery(req, param.Query); err != nil {
		return nil, err
	}

	if err = prepareHeader(req, param.Header); err != nil {
		return nil, err
	}

	if err = prepareJSONBody(req, param.Body); err != nil {
		return nil, err
	}

	return defaultClient.Do(req)
}

func PostForm(url string, param RequestParam) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return nil, err
	}

	if err = prepareQuery(req, param.Query); err != nil {
		return nil, err
	}

	if err = prepareHeader(req, param.Header); err != nil {
		return nil, err
	}

	if err = prepareFormBody(req, param.Body); err != nil {
		return nil, err
	}

	return defaultClient.Do(req)
}

func ParseJSONResponse[TBody any](resp *http.Response) (TBody, error) {
	var ret TBody
	contType := resp.Header.Get("Content-Type")
	if strings.Contains(contType, ContentTypeJSON) {
		if err := serder.JSONToObjectStream(resp.Body, &ret); err != nil {
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

type MultiPartFile struct {
	FieldName string
	FileName  string
	File      io.ReadCloser
	Header    textproto.MIMEHeader
}

type multiPartFileIterator struct {
	iterErr      error
	mr           *multipart.Reader
	nextPart     *multipart.Part
	nextFileName string
}

// 不可异步调用
func (m *multiPartFileIterator) MoveNext() (*MultiPartFile, error) {
	if m.iterErr != nil {
		return nil, m.iterErr
	}

	return &MultiPartFile{
		FieldName: m.nextPart.FormName(),
		FileName:  m.nextFileName,
		File: &multiPartFileReader{
			iter:     m,
			curPart:  m.nextPart,
			fileName: m.nextFileName,
		},
		Header: m.nextPart.Header,
	}, nil
}

func (m *multiPartFileIterator) Close() {}

type multiPartFileReader struct {
	iter     *multiPartFileIterator
	curPart  *multipart.Part
	fileName string
}

func (r *multiPartFileReader) Read(p []byte) (int, error) {
	if r.curPart != nil {
		return r.readFromCurPart(p)
	}

	part, err := r.iter.mr.NextPart()
	if err == io.EOF {
		r.iter.iterErr = iterator.ErrNoMoreItem
		return 0, io.EOF
	}
	if err != nil {
		r.iter.iterErr = err
		return 0, err
	}
	r.curPart = part

	fileName, err := ul.PathUnescape(part.FileName())
	if err != nil {
		r.iter.iterErr = fmt.Errorf("unescape file name: %w", err)
		return 0, r.iter.iterErr
	}

	if fileName == r.fileName {
		return r.readFromCurPart(p)
	}

	// 如果新一个part的文件名不同，则开始一个新文件
	r.iter.nextPart = part
	r.iter.nextFileName = fileName
	return 0, io.EOF
}

func (r *multiPartFileReader) readFromCurPart(p []byte) (int, error) {
	rd, err := r.curPart.Read(p)
	if err == io.EOF {
		r.curPart = nil
		return rd, nil
	}
	if err != nil {
		return 0, err
	}

	return rd, nil
}

func (m *multiPartFileReader) Close() error {
	// 如果在文件还未读取完毕的时候关闭文件，则后续所有文件都无法再读取（暂不支持跳文件读取）
	if m.curPart != nil {
		m.curPart.Close()
		m.iter.iterErr = fmt.Errorf("stream closed")
	}

	return nil
}

// 解析multipart/form-data响应。
//   - 只支持form参数在头部的情况。
//   - 支持文件分片（拆分成连续的同文件名的part）
func ParseMultiPartResponse(resp *http.Response) (map[string]string, iterator.Iterator[*MultiPartFile], error) {
	mtype, params, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if err != nil {
		return nil, nil, fmt.Errorf("parse content type: %w", err)
	}

	if mtype != ContentTypeMultiPart {
		return nil, nil, fmt.Errorf("unknow content type: %s", mtype)
	}

	boundary := params["boundary"]
	if boundary == "" {
		return nil, nil, fmt.Errorf("no boundary in content type: %s", mtype)
	}

	formValues := make(map[string]string)
	rd := multipart.NewReader(resp.Body, boundary)

	var firstFilePart *multipart.Part
	var firstFileName string
	for {
		part, err := rd.NextPart()
		if err == io.EOF {
			return formValues, iterator.Empty[*MultiPartFile](), nil
		}
		if err != nil {
			return nil, nil, err
		}

		formName := part.FormName()
		fileName := part.FileName()

		if formName == "" {
			continue
		}

		if fileName != "" {
			firstFilePart = part
			firstFileName, err = ul.PathUnescape(part.FileName())
			if err != nil {
				return nil, nil, fmt.Errorf("unescape file name: %w", err)
			}

			break
		}

		data, err := io.ReadAll(part)
		if err != nil {
			return nil, nil, err
		}

		formValues[formName] = string(data)
	}

	return formValues, &multiPartFileIterator{
		mr:           rd,
		nextPart:     firstFilePart,
		nextFileName: firstFileName,
	}, nil
}

type MultiPartRequestParam struct {
	Header   any
	Query    any
	Form     any
	Files    MultiPartFileIterator
	PartSize int64 // 文件分片大小，如果为0则不分片
}

type MultiPartFileIterator = iterator.Iterator[*IterMultiPartFile]
type IterMultiPartFile struct {
	FieldName string // 这个文件所属的form字段
	FileName  string // 文件名
	File      io.ReadCloser
}

func PostMultiPart(url string, param MultiPartRequestParam) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return nil, err
	}

	if err = prepareQuery(req, param.Query); err != nil {
		return nil, err
	}

	if err = prepareHeader(req, param.Header); err != nil {
		return nil, err
	}

	pr, pw := io.Pipe()
	muWriter := multipart.NewWriter(pw)

	setHeader(req.Header, "Content-Type", fmt.Sprintf("%s;boundary=%s", ContentTypeMultiPart, muWriter.Boundary()))

	writeResult := make(chan error, 1)
	go func() {
		writeResult <- func() error {
			defer pw.Close()
			defer muWriter.Close()

			if param.Form != nil {
				mp, err := objectToStringMap(param.Form)
				if err != nil {
					return fmt.Errorf("formValues object to map failed, err: %w", err)
				}

				for k, v := range mp {
					err := muWriter.WriteField(k, v)
					if err != nil {
						return fmt.Errorf("write form field failed, err: %w", err)
					}
				}
			}

			for {
				file, err := param.Files.MoveNext()
				if err == iterator.ErrNoMoreItem {
					break
				}
				if err != nil {
					return fmt.Errorf("opening file: %w", err)
				}

				if param.PartSize == 0 {
					err = sendFileOnePart(muWriter, file.FieldName, file.FileName, file.File)
				} else {
					err = sendFileMultiPart(muWriter, file.FieldName, file.FileName, file.File, param.PartSize)
				}
				file.File.Close()
				if err != nil {
					return err
				}
			}

			return nil
		}()
	}()

	req.Body = pr

	cli := http.Client{}
	resp, err := cli.Do(req)
	if err != nil {
		return nil, err
	}

	writeErr := <-writeResult
	if writeErr != nil {
		return nil, writeErr
	}

	return resp, nil
}

func sendFileMultiPart(muWriter *multipart.Writer, fieldName, fileName string, file io.ReadCloser, partSize int64) error {
	for {
		w, err := muWriter.CreateFormFile(fieldName, ul.PathEscape(fileName))
		if err != nil {
			return fmt.Errorf("create form file failed, err: %w", err)
		}

		_, err = io.Copy(w, io2.MustLength(file, partSize))
		if err == io.EOF {
			continue
		}
		if err == io.ErrUnexpectedEOF {
			break
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func sendFileOnePart(muWriter *multipart.Writer, fieldName, fileName string, file io.ReadCloser) error {
	w, err := muWriter.CreateFormFile(fieldName, ul.PathEscape(fileName))
	if err != nil {
		return fmt.Errorf("create form file failed, err: %w", err)
	}

	_, err = io.Copy(w, file)
	return err
}

func prepareQuery(req *http.Request, query any) error {
	if query == nil {
		return nil
	}

	mp, ok := query.(map[string]string)
	if !ok {
		var err error
		if mp, err = objectToStringMap(query); err != nil {
			return fmt.Errorf("query object to map: %w", err)
		}
	}

	values := make(ul.Values)
	for k, v := range mp {
		values.Add(k, v)
	}

	req.URL.RawQuery = values.Encode()
	return nil
}

func prepareHeader(req *http.Request, header any) error {
	if header == nil {
		return nil
	}

	mp, ok := header.(map[string]string)
	if !ok {
		var err error
		if mp, err = objectToStringMap(header); err != nil {
			return fmt.Errorf("header object to map: %w", err)
		}
	}

	req.Header = make(http.Header)
	for k, v := range mp {
		req.Header.Set(k, v)
	}
	return nil
}

func prepareJSONBody(req *http.Request, body any) error {
	setHeader(req.Header, "Content-Type", ContentTypeJSON)

	if body == nil {
		return nil
	}

	data, err := serder.ObjectToJSON(body)
	if err != nil {
		return err
	}

	req.ContentLength = int64(len(data))
	req.Body = io.NopCloser(bytes.NewReader(data))
	return nil
}

func prepareFormBody(req *http.Request, body any) error {
	setHeader(req.Header, "Content-Type", ContentTypeForm)

	if body == nil {
		return nil
	}

	mp, ok := body.(map[string]string)
	if !ok {
		var err error
		if mp, err = objectToStringMap(body); err != nil {
			return fmt.Errorf("body object to map: %w", err)
		}
	}

	values := make(ul.Values)
	for k, v := range mp {
		values.Add(k, v)
	}

	data := values.Encode()
	req.Body = io.NopCloser(strings.NewReader(data))
	req.ContentLength = int64(len(data))
	return nil
}

func setHeader(mp http.Header, key, value string) http.Header {
	if mp == nil {
		mp = make(http.Header)
	}

	mp.Set(key, value)
	return mp
}

func setValue(values ul.Values, key, value string) ul.Values {
	if values == nil {
		values = make(ul.Values)
	}

	values.Add(key, value)
	return values
}

func objectToStringMap(obj any) (map[string]string, error) {
	anyMap := make(map[string]any)
	dec, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName:          "json",
		Result:           &anyMap,
		WeaklyTypedInput: true,
	})
	if err != nil {
		return nil, err
	}

	err = dec.Decode(obj)
	if err != nil {
		return nil, err
	}

	ret := make(map[string]string)
	for k, v := range anyMap {
		val := reflect.ValueOf(v)
		for val.Kind() == reflect.Ptr {
			if val.IsNil() {
				break
			} else {
				val = val.Elem()
			}
		}

		if val.Kind() == reflect.Pointer {
			ret[k] = ""
		} else {
			ret[k] = fmt.Sprintf("%v", val)
		}
	}

	return ret, nil
}
