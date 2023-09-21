package schsdk

import "gitlink.org.cn/cloudream/common/sdks"

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
