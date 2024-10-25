package exec

import (
	"io"

	"gitlink.org.cn/cloudream/common/pkgs/types"
	"gitlink.org.cn/cloudream/common/utils/reflect2"
	"gitlink.org.cn/cloudream/common/utils/serder"
)

type VarID int

type Var struct {
	ID    VarID    `json:"id"`
	Value VarValue `json:"value"`
}

type VarPack[T VarValue] struct {
	ID    VarID `json:"id"`
	Value T     `json:"value"`
}

func (v *VarPack[T]) ToAny() AnyVar {
	return AnyVar{
		ID:    v.ID,
		Value: v.Value,
	}
}

// 变量的值
type VarValue interface {
	Clone() VarValue
}

var valueUnion = serder.UseTypeUnionExternallyTagged(types.Ref(types.NewTypeUnion[VarValue](
	(*StreamValue)(nil),
	(*SignalValue)(nil),
	(*StringValue)(nil),
)))

func UseVarValue[T VarValue]() {
	valueUnion.Add(reflect2.TypeOf[T]())
}

type AnyVar = VarPack[VarValue]

func V(id VarID, value VarValue) AnyVar {
	return AnyVar{
		ID:    id,
		Value: value,
	}
}

type StreamValue struct {
	Stream io.ReadCloser `json:"-"`
}

// 不应该被调用
func (v *StreamValue) Clone() VarValue {
	panic("StreamValue should not be cloned")
}

type SignalValue struct{}

func (o *SignalValue) Clone() VarValue {
	return &SignalValue{}
}

type SignalVar = VarPack[*SignalValue]

func NewSignal(id VarID) SignalVar {
	return SignalVar{
		ID:    id,
		Value: &SignalValue{},
	}
}

type StringValue struct {
	Value string `json:"value"`
}

func (o *StringValue) Clone() VarValue {
	return &StringValue{Value: o.Value}
}
