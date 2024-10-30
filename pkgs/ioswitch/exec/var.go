package exec

import (
	"io"

	"gitlink.org.cn/cloudream/common/pkgs/types"
	"gitlink.org.cn/cloudream/common/utils/reflect2"
	"gitlink.org.cn/cloudream/common/utils/serder"
)

type VarID int

type Var[T VarValue] struct {
	ID    VarID `json:"id"`
	Value T     `json:"value"`
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

type StreamValue struct {
	Stream io.ReadCloser `json:"-"`
}

// 不应该被调用
func (v *StreamValue) Clone() VarValue {
	panic("StreamValue should not be cloned")
}

type StreamVar = Var[*StreamValue]

func NewStreamVar(id VarID, stream io.ReadCloser) StreamVar {
	return StreamVar{
		ID:    id,
		Value: &StreamValue{Stream: stream},
	}
}

type SignalValue struct{}

func (o *SignalValue) Clone() VarValue {
	return &SignalValue{}
}

type SignalVar = Var[*SignalValue]

func NewSignalVar(id VarID) SignalVar {
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

type StringVar = Var[*StringValue]

func NewStringVar(id VarID, value string) StringVar {
	return StringVar{
		ID:    id,
		Value: &StringValue{Value: value},
	}
}
