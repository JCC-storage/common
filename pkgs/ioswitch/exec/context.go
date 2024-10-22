package exec

import (
	"context"
	"fmt"

	"gitlink.org.cn/cloudream/common/utils/reflect2"
)

var ErrValueNotFound = fmt.Errorf("value not found")

type ExecContext struct {
	Context context.Context
	Values  map[any]any
}

func NewExecContext() *ExecContext {
	return NewWithContext(context.Background())
}

func NewWithContext(ctx context.Context) *ExecContext {
	return &ExecContext{Context: ctx, Values: make(map[any]any)}
}

// error只会是ErrValueNotFound
func (c *ExecContext) Value(key any) (any, error) {
	value, ok := c.Values[key]
	if !ok {
		return nil, ErrValueNotFound
	}
	return value, nil
}

func (c *ExecContext) SetValue(key any, value any) {
	c.Values[key] = value
}

func ValueByType[T any](ctx *ExecContext) (T, error) {
	var ret T

	value, err := ctx.Value(reflect2.TypeOf[T]())
	if err != nil {
		return ret, err
	}

	ret, ok := value.(T)
	if !ok {
		return ret, fmt.Errorf("value is %T, not %T", value, ret)
	}

	return ret, nil
}
