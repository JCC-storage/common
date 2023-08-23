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

func NewClient(cfg *Config) *Client {
	return &Client{
		baseURL: cfg.URL,
	}
}

type PoolClient struct {
	*Client
	owner *Pool
}

func (c *PoolClient) Close() {
	c.owner.Release(c)
}

type Pool struct {
	cfg *Config
}

func NewPool(cfg *Config) *Pool {
	return &Pool{
		cfg: cfg,
	}
}
func (p *Pool) Acquire() (*PoolClient, error) {
	cli := NewClient(p.cfg)
	return &PoolClient{
		Client: cli,
		owner:  p,
	}, nil
}

func (p *Pool) Release(cli *PoolClient) {

}
