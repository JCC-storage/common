package json

import (
	"reflect"

	jsoniter "github.com/json-iterator/go"
	"gitlink.org.cn/cloudream/common/pkgs/types"
)

type Config struct {
	unionHandler *UnionHandler
	exts         []jsoniter.Extension
}

func New() *Config {
	return &Config{
		unionHandler: &UnionHandler{
			internallyTagged: make(map[reflect.Type]*anyTypeUnionInternallyTagged),
			externallyTagged: make(map[reflect.Type]*anyTypeUnionExternallyTagged),
		},
	}
}

func (c *Config) UseUnionInternallyTagged(u *types.AnyTypeUnion, tagField string) *Config {
	iu := &anyTypeUnionInternallyTagged{
		Union:     u,
		TagField:  tagField,
		TagToType: make(map[string]reflect.Type),
	}

	for _, eleType := range u.ElementTypes {
		iu.Add(eleType)
	}

	c.unionHandler.internallyTagged[u.UnionType] = iu
	return c
}

func (c *Config) UseUnionExternallyTagged(u *types.AnyTypeUnion) *Config {
	eu := &anyTypeUnionExternallyTagged{
		Union:          u,
		TypeNameToType: make(map[string]reflect.Type),
	}

	for _, eleType := range u.ElementTypes {
		eu.Add(eleType)
	}

	c.unionHandler.externallyTagged[u.UnionType] = eu
	return c
}

func (c *Config) UseExtension(ext jsoniter.Extension) *Config {
	c.exts = append(c.exts, ext)
	return c
}

func (c *Config) Build() Serder {
	cfg := jsoniter.Config{}
	api := cfg.Froze()

	api.RegisterExtension(c.unionHandler)

	for _, ext := range c.exts {
		api.RegisterExtension(ext)
	}

	return Serder{
		cfg: *c,
		api: api,
	}
}
