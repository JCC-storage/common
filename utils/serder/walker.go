package serder

import (
	"reflect"

	"github.com/zyedidia/generic/stack"
)

type WalkEvent interface{}

type StructBeginEvent struct {
	Value reflect.Value
}

type StructArriveFieldEvent struct {
	Info  reflect.StructField
	Value reflect.Value
}

type StructLeaveFieldEvent struct {
	Info  reflect.StructField
	Value reflect.Value
}

type StructEndEvent struct {
	Value reflect.Value
}

type ArrayBeginEvent struct {
	Value reflect.Value
}

type ArrayArriveElementEvent struct {
	Index int
	Value reflect.Value
}

type ArrayLeaveElementEvent struct {
	Index int
	Value reflect.Value
}

type ArrayEndEvent struct {
	Value reflect.Value
}

type MapBeginEvent struct {
	Value reflect.Value
}

type MapArriveEntryEvent struct {
	Key   reflect.Value
	Value reflect.Value
}

type MapLeaveEntryEvent struct {
	Key   reflect.Value
	Value reflect.Value
}

type MapEndEvent struct {
	Value reflect.Value
}

type WalkingOp int

const (
	Next WalkingOp = iota
	Skip
	Stop
)

type Walker func(ctx *WalkContext, event WalkEvent) WalkingOp

type WalkContext struct {
	stack *stack.Stack[any]
}

func (c *WalkContext) StackPush(val any) {
	c.stack.Push(val)
}

func (c *WalkContext) StackPop() any {
	return c.stack.Pop()
}

func (c *WalkContext) StackPeek() any {
	return c.stack.Peek()
}

type WalkOption struct {
	StackValues []any
}

func WalkValue(value any, walker Walker, opts ...WalkOption) *WalkContext {
	var opt WalkOption
	if len(opts) > 0 {
		opt = opts[0]
	}

	ctx := &WalkContext{
		stack: stack.New[any](),
	}

	for _, v := range opt.StackValues {
		ctx.StackPush(v)
	}

	doWalking(ctx, reflect.ValueOf(value), walker)

	return ctx
}

func doWalking(ctx *WalkContext, val reflect.Value, walker Walker) WalkingOp {
	if !WillWalkInto(val) {
		return Next
	}

	switch val.Kind() {
	case reflect.Array:
		fallthrough
	case reflect.Slice:
		if walker(ctx, ArrayBeginEvent{Value: val}) == Stop {
			return Stop
		}

		for i := 0; i < val.Len(); i++ {
			eleVal := val.Index(i)

			op := walker(ctx, ArrayArriveElementEvent{
				Index: i,
				Value: eleVal,
			})

			if op == Skip {
				if walker(ctx, ArrayLeaveElementEvent{
					Index: i,
					Value: eleVal,
				}) == Stop {
					return Stop
				}
				continue
			}

			if op == Stop {
				return Stop
			}

			if doWalking(ctx, eleVal, walker) == Stop {
				return Stop
			}

			if walker(ctx, ArrayLeaveElementEvent{
				Index: i,
				Value: eleVal,
			}) == Stop {
				return Stop
			}
		}

		if walker(ctx, ArrayEndEvent{Value: val}) == Stop {
			return Stop
		}

	case reflect.Map:
		if walker(ctx, MapBeginEvent{Value: val}) == Stop {
			return Stop
		}

		keys := val.MapKeys()
		for _, key := range keys {
			val := val.MapIndex(key)

			op := walker(ctx, MapArriveEntryEvent{
				Key:   key,
				Value: val,
			})

			if op == Skip {
				if walker(ctx, MapLeaveEntryEvent{
					Key:   key,
					Value: val,
				}) == Stop {
					return Stop
				}
				continue
			}

			if op == Stop {
				return Stop
			}

			if doWalking(ctx, val, walker) == Stop {
				return Stop
			}

			if walker(ctx, MapLeaveEntryEvent{
				Key:   key,
				Value: val,
			}) == Stop {
				return Stop
			}
		}

		if walker(ctx, MapEndEvent{Value: val}) == Stop {
			return Stop
		}

	case reflect.Struct:
		if walker(ctx, StructBeginEvent{Value: val}) == Stop {
			return Stop
		}

		for i := 0; i < val.NumField(); i++ {
			field := val.Field(i)

			op := walker(ctx, StructArriveFieldEvent{
				Info:  val.Type().Field(i),
				Value: field,
			})

			if op == Skip {
				if walker(ctx, StructLeaveFieldEvent{
					Info:  val.Type().Field(i),
					Value: field,
				}) == Stop {
					return Stop
				}
				continue
			}

			if op == Stop {
				return Stop
			}

			if doWalking(ctx, field, walker) == Stop {
				return Stop
			}

			if walker(ctx, StructLeaveFieldEvent{
				Info:  val.Type().Field(i),
				Value: field,
			}) == Stop {
				return Stop
			}
		}

		if walker(ctx, StructEndEvent{Value: val}) == Stop {
			return Stop
		}

	case reflect.Interface:
		fallthrough
	case reflect.Pointer:
		eleVal := val.Elem()
		return doWalking(ctx, eleVal, walker)
	}

	return Next
}

const (
	WillWalkIntoTypeKinds = (1 << reflect.Array) | (1 << reflect.Map) | (1 << reflect.Slice) | (1 << reflect.Struct)
)

func WillWalkInto(val reflect.Value) bool {
	if val.IsZero() {
		return false
	}

	typ := val.Type()
	typeKind := typ.Kind()
	if typeKind == reflect.Interface || typeKind == reflect.Pointer {
		return WillWalkInto(val.Elem())
	}

	return ((1 << typeKind) & WillWalkIntoTypeKinds) != 0
}
