package http

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	ul "net/url"
	"strings"

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

type MultiPartRequestParam struct {
	Header   any
	Query    any
	Form     any
	DataName string
	Data     io.Reader
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

			if param.Data != nil {
				w, err := muWriter.CreateFormFile("file", param.DataName)
				if err != nil {
					return fmt.Errorf("create form file failed, err: %w", err)
				}

				_, err = io.Copy(w, param.Data)
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

	req.Body = serder.ObjectToJSONStream(body)
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

	req.Body = io.NopCloser(strings.NewReader(values.Encode()))
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
