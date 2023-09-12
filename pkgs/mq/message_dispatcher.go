package mq

import (
	"fmt"

	myreflect "gitlink.org.cn/cloudream/common/utils/reflect"
)

type HandlerFn func(svcBase any, msg *Message) (*Message, error)

type MessageDispatcher struct {
	Handlers map[myreflect.Type]HandlerFn
}

func NewMessageDispatcher() MessageDispatcher {
	return MessageDispatcher{
		Handlers: make(map[myreflect.Type]HandlerFn),
	}
}

func (h *MessageDispatcher) Add(typ myreflect.Type, handler HandlerFn) {
	h.Handlers[typ] = handler
}

func (h *MessageDispatcher) Handle(svcBase any, msg *Message) (*Message, error) {
	typ := myreflect.TypeOfValue(msg.Body)
	fn, ok := h.Handlers[typ]
	if !ok {
		return nil, fmt.Errorf("unsupported message type: %s", typ.Name())
	}

	return fn(svcBase, msg)
}

// 将Service中的一个接口函数作为指定类型消息的处理函数
func AddServiceFn[TSvc any, TReq MessageBody, TResp MessageBody](dispatcher *MessageDispatcher, svcFn func(svc TSvc, msg TReq) (TResp, *CodeMessage)) {
	dispatcher.Add(myreflect.TypeOf[TReq](), func(svcBase any, reqMsg *Message) (*Message, error) {

		reqMsgBody := reqMsg.Body.(TReq)
		ret, codeMsg := svcFn(svcBase.(TSvc), reqMsgBody)
		respMsg := MakeAppDataMessage(ret)
		respMsg.SetCodeMessage(codeMsg.Code, codeMsg.Message)

		return &respMsg, nil
	})
}

// 将Service中的一个*没有返回值的*接口函数作为指定类型消息的处理函数
func AddNoRespServiceFn[TSvc any, TReq MessageBody](dispatcher *MessageDispatcher, svcFn func(svc TSvc, msg TReq)) {
	dispatcher.Add(myreflect.TypeOf[TReq](), func(svcBase any, reqMsg *Message) (*Message, error) {

		reqMsgBody := reqMsg.Body.(TReq)
		svcFn(svcBase.(TSvc), reqMsgBody)

		return nil, nil
	})
}
