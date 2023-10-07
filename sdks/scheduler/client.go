package schsdk

import (
	"gitlink.org.cn/cloudream/common/sdks"
)

type response[T any] struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

func (r *response[T]) ToError() *sdks.CodeMessageError {
	return &sdks.CodeMessageError{
		Code:    r.Code,
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
