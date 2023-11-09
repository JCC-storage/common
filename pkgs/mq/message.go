package mq

import (
	"gitlink.org.cn/cloudream/common/pkgs/types"
	myreflect "gitlink.org.cn/cloudream/common/utils/reflect"
	"gitlink.org.cn/cloudream/common/utils/serder"
)

const (
	MessageTypeAppData   = "AppData"
	MessageTypeHeartbeat = "Heartbeat"
)

type MessageBody interface {
	// 此方法无任何作用，仅用于避免MessageBody是一个空interface，从而导致任何类型的值都可以赋值给它
	// 与下方的MessageBodyBase配合使用：
	// Noop只让实现了此接口的类型能赋值给它，内嵌MessageBodyBase让类型必须是个指针类型，
	// 这就确保了Message.Body必是某个类型的指针类型，避免序列化、反序列化过程出错
	Noop()
}

// 这个结构体无任何字段，但实现了Noop，每种MessageBody都要内嵌这个结构体
type MessageBodyBase struct{}

// 此处的receiver是指针
func (b *MessageBodyBase) Noop() {}

type Message struct {
	Type    string         `json:"type"`
	Headers map[string]any `json:"headers"`
	Body    MessageBody    `json:"body"`
}

func (m *Message) GetRequestID() string {
	reqID, _ := m.Headers["requestID"].(string)
	return reqID
}

func (m *Message) SetRequestID(id string) {
	m.Headers["requestID"] = id
}

func (m *Message) GetKeepAlive() int {
	timeoutMs, _ := m.Headers["keepAliveTimeout"].(float64)
	return int(timeoutMs)
}

func (m *Message) SetKeepAlive(timeoutMs int) {
	m.Headers["keepAliveTimeout"] = timeoutMs
}

func (m *Message) SetCodeMessage(code string, msg string) {
	m.Headers["responseCode"] = code
	m.Headers["responseMessage"] = msg
}

func (m *Message) GetCodeMessage() (string, string) {
	code, _ := m.Headers["responseCode"].(string)
	msg, _ := m.Headers["responseMessage"].(string)
	return code, msg
}

func MakeAppDataMessage(body MessageBody) Message {
	msg := Message{
		Type:    MessageTypeAppData,
		Headers: make(map[string]any),
		Body:    body,
	}

	return msg
}

func MakeHeartbeatMessage() Message {
	msg := Message{
		Type:    MessageTypeHeartbeat,
		Headers: make(map[string]any),
	}

	return msg
}

var msgBodyTypeUnion = serder.UseTypeUnionExternallyTagged(types.Ref(types.NewTypeUnion[MessageBody]()))

// 所有新定义的Message都需要在init中调用此函数
func RegisterMessage[T MessageBody]() {
	err := msgBodyTypeUnion.Add(myreflect.TypeOf[T]())
	if err != nil {
		panic(err)
	}
}

func Serialize(msg Message) ([]byte, error) {
	return serder.ObjectToJSONEx(msg)
}

func Deserialize(data []byte) (*Message, error) {
	return serder.JSONToObjectEx[*Message](data)
}
