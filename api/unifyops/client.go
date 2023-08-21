package unifyops

import (
	"fmt"

	"gitlink.org.cn/cloudream/common/api"
)

type response[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

func (r *response[T]) ToError() *api.CodeMessageError {
	return &api.CodeMessageError{
		Code:    fmt.Sprintf("%d", r.Code),
		Message: r.Message,
	}
}

type Client struct {
	baseURL string
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
	}
}

// func (c *Client) SendRequest(path string, method string, body []byte) (*http.Response, error) {
// 	url := c.BaseURL + path
// 	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
// 	if err != nil {
// 		return nil, err
// 	}

// 	client := &http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return resp, nil
// }
