package exec

import (
	"gitlink.org.cn/cloudream/common/pkgs/types"
	"gitlink.org.cn/cloudream/common/utils/reflect2"
	"gitlink.org.cn/cloudream/common/utils/serder/json"
)

type ConfigBuilder struct {
	unions         []*types.AnyTypeUnion
	opUnion        types.TypeUnion[Op]
	workerInfoType reflect2.Type
}

func (c *ConfigBuilder) UseUnion(u *types.AnyTypeUnion) *ConfigBuilder {
	c.unions = append(c.unions, u)
	return c
}

func (c *ConfigBuilder) UseOpType(nilValue Op) *ConfigBuilder {
	c.opUnion.Add(reflect2.TypeOfValue(nilValue))
	return c
}

func (c *ConfigBuilder) UseWorkerInfoType(nilValue WorkerInfo) *ConfigBuilder {
	c.workerInfoType = reflect2.TypeOfValue(nilValue)
	return c
}

func (c *ConfigBuilder) Build() Config {
	b := json.New().UseUnionExternallyTagged(c.opUnion.ToAny())
	for _, u := range c.unions {
		b.UseUnionExternallyTagged(u)
	}

	// b.UseExtension(&workerInfoJSONExt{workerInfoType: c.workerInfoType})

	ser := b.Build()
	return Config{
		Serder: ser,
	}
}

type Config struct {
	Serder json.Serder
}
