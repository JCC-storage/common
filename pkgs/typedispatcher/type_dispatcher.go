package typedispatcher

import (
	"gitlink.org.cn/cloudream/common/utils/reflect2"
)

type HandlerFn[TRet any] func(val any) TRet

type TypeDispatcher[TRet any] struct {
	handlers map[reflect2.Type]HandlerFn[TRet]
}

func NewTypeDispatcher[TRet any]() TypeDispatcher[TRet] {
	return TypeDispatcher[TRet]{
		handlers: make(map[reflect2.Type]HandlerFn[TRet]),
	}
}

func (t *TypeDispatcher[TRet]) Add(typ reflect2.Type, fn HandlerFn[TRet]) {
	t.handlers[typ] = fn
}

func (t *TypeDispatcher[TRet]) Dispatch(val any) (TRet, bool) {
	var ret TRet
	typ := reflect2.TypeOfValue(val)
	handler, ok := t.handlers[typ]
	if !ok {
		return ret, false
	}

	return handler(val), true
}

func Add[T any, TRet any](dispatcher TypeDispatcher[TRet], handler func(val T) TRet) {
	dispatcher.Add(reflect2.TypeOf[T](), func(val any) TRet {
		return handler(val.(T))
	})
}
