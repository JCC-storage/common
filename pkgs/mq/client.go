package mq

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/go-multierror"
	"github.com/streadway/amqp"
	"gitlink.org.cn/cloudream/common/consts/errorcode"
	"gitlink.org.cn/cloudream/common/pkgs/logger"
	myreflect "gitlink.org.cn/cloudream/common/utils/reflect"
)

const (
	DirectReplyTo = "amq.rabbitmq.reply-to"

	KeepAliveTimeoutMaxTimes = 3
)

var ErrWaitResponseTimeout = fmt.Errorf("wait response timeout")

type CodeMessageError struct {
	code    string
	message string
}

func (e *CodeMessageError) Error() string {
	return fmt.Sprintf("code: %s, message: %s", e.code, e.message)
}

type SendOption struct {
	// 发送消息的超时时间，为0代表不设置超时时间
	Timeout time.Duration
}

type RequestOption struct {
	// 等待响应的超时时间，为0代表不设置超时时间。
	// 如果设置了KeepAlive，那么这个设置代表心跳包发送间隔
	Timeout time.Duration
	// 让服务端定时发送心跳包来表示存活。连续丢失3个心跳包，则认为连接已经断开。
	KeepAlive bool
}

type RoundTripper interface {
	Send(msg Message, opt SendOption) error
	Request(req Message, opt RequestOption) (*Message, error)
	Close() error
}

type requesting struct {
	RequestID      string
	Receiving      chan *Message
	ReceiveStopped chan bool
	TimeoutTimes   int
	Option         RequestOption
}

type RabbitMQTransport struct {
	connection *amqp.Connection
	channel    *amqp.Channel
	exchange   string
	key        string

	requestings     map[string]*requesting
	requestingsLock sync.Mutex

	closed chan any
}

func NewRabbitMQTransport(url string, key string, exchange string) (*RabbitMQTransport, error) {
	connection, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("connecting to %s: %w", url, err)
	}

	channel, err := connection.Channel()
	if err != nil {
		connection.Close()
		return nil, fmt.Errorf("openning channel on connection: %w", err)
	}

	cli := &RabbitMQTransport{
		connection:  connection,
		channel:     channel,
		exchange:    exchange,
		key:         key,
		requestings: make(map[string]*requesting),
		closed:      make(chan any),
	}

	// NOTE! 经测试发现，必须在Publish之前调用Consume进行消息接收，否则Consume会返回错误
	// 因此这段代码不能移动到serve函数中，必须放在这里，保证顺序
	recvChan, err := channel.Consume(
		// 一个特殊队列，服务端的回复消息都会发送到这个队列里
		DirectReplyTo,
		"",
		true,
		true,
		false,
		false,
		nil,
	)
	if err != nil {
		channel.Close()
		connection.Close()
		return nil, fmt.Errorf("openning consume channel: %w", err)
	}

	go func() {
		err := cli.serve(recvChan)
		if err != nil {
			// TODO 错误处理
			logger.Std.Warnf("rabbitmq client serving: %s", err.Error())
		}
	}()

	return cli, nil
}

func (cli *RabbitMQTransport) Request(req Message, opt RequestOption) (*Message, error) {
	// 如果没有设置timeout，却设置了keepalive，那么默认心跳间隔为15秒
	if opt.KeepAlive && opt.Timeout == 0 {
		opt.Timeout = time.Second * 15
	}

	reqID := uuid.NewString()
	req.SetRequestID(reqID)
	if opt.KeepAlive {
		req.SetKeepAlive(int(opt.Timeout / time.Millisecond))
	}

	reqing := &requesting{
		RequestID:      reqID,
		Receiving:      make(chan *Message),
		ReceiveStopped: make(chan bool),
		TimeoutTimes:   0,
		Option:         opt,
	}
	cli.startRequesting(reqing)
	defer cli.cancelRequsting(reqing)

	err := cli.Send(req, SendOption{
		Timeout: opt.Timeout,
	})
	if err != nil {
		return nil, fmt.Errorf("sending message: %w", err)
	}

	// 启动超时定时器
	if opt.Timeout != 0 {
		return cli.receiveWithTimeout(reqing)
	}

	return cli.receiveNoTimeout(reqing)
}

func (cli *RabbitMQTransport) receiveWithTimeout(reqing *requesting) (*Message, error) {
	ticker := time.NewTicker(reqing.Option.Timeout)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			reqing.TimeoutTimes++
			if reqing.Option.KeepAlive && reqing.TimeoutTimes < KeepAliveTimeoutMaxTimes {
				continue
			}

			return nil, ErrWaitResponseTimeout

		case msg := <-reqing.Receiving:
			if msg.Type == MessageTypeHeartbeat && reqing.Option.KeepAlive {
				reqing.TimeoutTimes = 0
				ticker.Reset(reqing.Option.Timeout)
				continue
			}

			if msg.Type == MessageTypeAppData {
				return msg, nil
			}
		}
	}
}

func (cli *RabbitMQTransport) receiveNoTimeout(reqing *requesting) (*Message, error) {
	for {
		msg := <-reqing.Receiving
		if msg.Type != MessageTypeAppData {
			continue
		}

		return msg, nil
	}
}

func (cli *RabbitMQTransport) startRequesting(reqing *requesting) {
	cli.requestingsLock.Lock()
	cli.requestings[reqing.RequestID] = reqing
	cli.requestingsLock.Unlock()
}

func (cli *RabbitMQTransport) cancelRequsting(reqing *requesting) {
	cli.requestingsLock.Lock()
	delete(cli.requestings, reqing.RequestID)
	cli.requestingsLock.Unlock()

	// 告诉发送端，接收端已经停止接收
	close(reqing.ReceiveStopped)
}

func (c *RabbitMQTransport) findReuqesting(reqID string) *requesting {
	c.requestingsLock.Lock()
	reqing := c.requestings[reqID]
	c.requestingsLock.Unlock()
	return reqing
}

func (c *RabbitMQTransport) Send(msg Message, opt SendOption) error {
	data, err := Serialize(msg)
	if err != nil {
		return fmt.Errorf("serialize message failed: %w", err)
	}

	expiration := ""
	if opt.Timeout > 0 {
		if opt.Timeout < time.Millisecond {
			expiration = "1"
		} else {
			expiration = fmt.Sprintf("%d", opt.Timeout.Milliseconds()+1)
		}
	}

	err = c.channel.Publish(c.exchange, c.key, false, false, amqp.Publishing{
		ContentType: "text/plain",
		Body:        data,
		// 设置了此字段后rabbitmq会建立一个临时且私有的队列，服务端的回复消息都是送到此队列中
		ReplyTo:    DirectReplyTo,
		Expiration: expiration,
	})

	if err != nil {
		return fmt.Errorf("publishing data: %w", err)
	}

	return nil
}

func (c *RabbitMQTransport) serve(recvChan <-chan amqp.Delivery) error {
	for {
		select {
		case rawMsg, ok := <-recvChan:
			if !ok {
				return fmt.Errorf("receive channel closed")
			}

			msg, err := Deserialize(rawMsg.Body)
			if err != nil {
				// TODO 记录日志
				logger.Std.Warnf("deserializing message body: %s", err.Error())
				continue
			}

			reqID := msg.GetRequestID()
			if reqID != "" {
				reqing := c.findReuqesting(reqID)
				if reqing != nil {
					select {
					case reqing.Receiving <- msg:
					case <-reqing.ReceiveStopped:
						// 防止发送端在接收端停止消费时，发送端还在发送导致的阻塞
					}
				}
			}

		case <-c.closed:
			return nil
		}
	}
}

func (c *RabbitMQTransport) Close() error {
	var retErr error

	close(c.closed)

	err := c.channel.Close()
	if err != nil {
		multierror.Append(retErr, fmt.Errorf("closing channel: %w", err))
	}

	err = c.connection.Close()
	if err != nil {
		multierror.Append(retErr, fmt.Errorf("closing connection: %w", err))
	}

	return retErr
}

// 发送消息并等待回应。因为无法自动推断出TResp的类型，所以将其放在第一个手工填写，之后的TBody可以自动推断出来
func Request[TSvc any, TReq MessageBody, TResp MessageBody](_ func(svc TSvc, msg TReq) (TResp, *CodeMessage), cli RoundTripper, req TReq, opts ...RequestOption) (TResp, error) {
	opt := RequestOption{Timeout: time.Second * 15}
	if len(opts) > 0 {
		opt = opts[0]
	}

	var defRet TResp

	resp, err := cli.Request(MakeAppDataMessage(req), opt)
	if err != nil {
		return defRet, fmt.Errorf("requesting: %w", err)
	}

	errCode, errMsg := resp.GetCodeMessage()
	if errCode != errorcode.OK {
		return defRet, &CodeMessageError{
			code:    errCode,
			message: errMsg,
		}
	}

	respBody, ok := resp.Body.(TResp)
	if !ok {
		return defRet, fmt.Errorf("expect a %s body, but got %s",
			myreflect.ElemTypeOf[TResp]().Name(),
			myreflect.TypeOfValue(resp.Body).Name())
	}

	return respBody, nil
}

// 发送消息，不等待回应
func Send[TSvc any, TReq MessageBody](_ func(svc TSvc, msg TReq), cli RoundTripper, msg TReq, opts ...SendOption) error {
	opt := SendOption{Timeout: time.Second * 15}
	if len(opts) > 0 {
		opt = opts[0]
	}

	req := MakeAppDataMessage(msg)

	err := cli.Send(req, opt)
	if err != nil {
		return fmt.Errorf("sending: %w", err)
	}

	return nil
}
