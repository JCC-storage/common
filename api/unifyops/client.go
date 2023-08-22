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
