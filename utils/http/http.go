package http

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	ul "net/url"
	"strings"

	"gitlink.org.cn/cloudream/common/pkgs/iterator"
	"gitlink.org.cn/cloudream/common/utils/math"
	"gitlink.org.cn/cloudream/common/utils/serder"
)

const (
	ContentTypeJSON        = "application/json"
	ContentTypeForm        = "application/x-www-form-urlencoded"
	ContentTypeMultiPart   = "multipart/form-data"
	ContentTypeOctetStream = "application/octet-stream"
)

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

	return http.DefaultClient.Do(req)
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

	return http.DefaultClient.Do(req)
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

	return http.DefaultClient.Do(req)
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

	return http.DefaultClient.Do(req)
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

	return ret, fmt.Errorf("unknow response content type: %s, status: %d, body(prefix): %s", contType, resp.StatusCode, strCont[:math.Min(len(strCont), 200)])
}

type MultiPartRequestParam struct {
	Header any
	Query  any
	Form   any
	Files  MultiPartFileIterator
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
				mp, err := serder.ObjectToMap(param.Form)
				if err != nil {
					return fmt.Errorf("formValues object to map failed, err: %w", err)
				}

				for k, v := range mp {
					err := muWriter.WriteField(k, fmt.Sprintf("%v", v))
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

				err = func() error {
					defer file.File.Close()

					w, err := muWriter.CreateFormFile(file.FieldName, file.FileName)
					if err != nil {
						return fmt.Errorf("create form file failed, err: %w", err)
					}

					_, err = io.Copy(w, file.File)
					if err != nil {
						return err
					}
					return nil
				}()
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

func prepareQuery(req *http.Request, query any) error {
	if query == nil {
		return nil
	}

	mp, ok := query.(map[string]any)
	if !ok {
		var err error
		if mp, err = serder.ObjectToMap(query); err != nil {
			return fmt.Errorf("query object to map: %w", err)
		}
	}

	values := make(ul.Values)
	for k, v := range mp {
		values.Add(k, fmt.Sprintf("%v", v))
	}

	req.URL.RawQuery = values.Encode()
	return nil
}

func prepareHeader(req *http.Request, header any) error {
	if header == nil {
		return nil
	}

	mp, ok := header.(map[string]any)
	if !ok {
		var err error
		if mp, err = serder.ObjectToMap(header); err != nil {
			return fmt.Errorf("header object to map: %w", err)
		}
	}

	req.Header = make(http.Header)
	for k, v := range mp {
		req.Header.Set(k, fmt.Sprintf("%v", v))
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

	mp, ok := body.(map[string]any)
	if !ok {
		var err error
		if mp, err = serder.ObjectToMap(body); err != nil {
			return fmt.Errorf("body object to map: %w", err)
		}
	}

	values := make(ul.Values)
	for k, v := range mp {
		values.Add(k, fmt.Sprintf("%v", v))
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
