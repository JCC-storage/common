package blockchain

import (
	"fmt"

	"gitlink.org.cn/cloudream/common/sdks"
)

type response[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

const (
	ResponseCodeOK int = 200
)

func (r *response[T]) ToError() *sdks.CodeMessageError {
	return &sdks.CodeMessageError{
		Code:    fmt.Sprintf("%d", r.Code),
		Message: r.Message,
	}
}

type Client struct {
	baseURL string
}

func NewClient(cfg *Config) *Client {
	return &Client{
		baseURL: cfg.URL,
	}
}

type Pool interface {
	Acquire() (*Client, error)
	Release(cli *Client)
}

type pool struct {
	cfg *Config
}

func NewPool(cfg *Config) Pool {
	return &pool{
		cfg: cfg,
	}
}
func (p *pool) Acquire() (*Client, error) {
	cli := NewClient(p.cfg)
	return cli, nil
}

func (p *pool) Release(cli *Client) {

}
