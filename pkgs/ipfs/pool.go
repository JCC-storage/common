package ipfs

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
	cli, err := NewClient(p.cfg)
	if err != nil {
		return nil, err
	}

	return &PoolClient{
		Client: cli,
		owner:  p,
	}, nil
}

func (p *Pool) Release(cli *PoolClient) {
}
