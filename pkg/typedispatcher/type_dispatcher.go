package typedispatcher

import (
	myreflect "gitlink.org.cn/cloudream/common/utils/reflect"
)

type HandlerFn[TRet any] func(val any) TRet

type TypeDispatcher[TRet any] struct {
	handlers map[myreflect.Type]HandlerFn[TRet]
}

func NewTypeDispatcher[TRet any]() TypeDispatcher[TRet] {
	return TypeDispatcher[TRet]{
		handlers: make(map[myreflect.Type]HandlerFn[TRet]),
	}
}

func (t *TypeDispatcher[TRet]) Add(typ myreflect.Type, fn HandlerFn[TRet]) {
	t.handlers[typ] = fn
}

func (t *TypeDispatcher[TRet]) Dispatch(val any) (TRet, bool) {
	var ret TRet
	typ := myreflect.TypeOfValue(val)
	handler, ok := t.handlers[typ]
	if !ok {
		return ret, false
	}

	return handler(val), true
}

func Add[T any, TRet any](dispatcher TypeDispatcher[TRet], handler func(val T) TRet) {
	dispatcher.Add(myreflect.TypeOf[T](), func(val any) TRet {
		return handler(val.(T))
	})
}
