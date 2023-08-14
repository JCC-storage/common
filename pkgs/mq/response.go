package mq

import (
	"gitlink.org.cn/cloudream/common/consts/errorcode"
)

type CodeMessage struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (msg *CodeMessage) IsOK() bool {
	return msg.Code == errorcode.OK
}

func (msg *CodeMessage) IsFailed() bool {
	return !msg.IsOK()
}

func OK() *CodeMessage {
	return &CodeMessage{
		Code:    errorcode.OK,
		Message: "",
	}
}

func Failed(errCode string, msg string) *CodeMessage {
	return &CodeMessage{
		Code:    errCode,
		Message: msg,
	}
}

func ReplyFailed[T any](errCode string, msg string) (*T, *CodeMessage) {
	return nil, &CodeMessage{
		Code:    errCode,
		Message: msg,
	}
}

func ReplyOK[T any](val T) (*T, *CodeMessage) {
	return &val, &CodeMessage{
		Code:    errorcode.OK,
		Message: "",
	}
}
