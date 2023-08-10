package mq

import (
	"fmt"
	"sync"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/streadway/amqp"
	"gitlink.org.cn/cloudream/common/consts/errorcode"
	"gitlink.org.cn/cloudream/common/pkg/future"
	"gitlink.org.cn/cloudream/common/pkg/logger"
	myreflect "gitlink.org.cn/cloudream/common/utils/reflect"
)

const (
	DIRECT_REPLY_TO = "amq.rabbitmq.reply-to"
)

type CodeMessageError struct {
	code    string
	message string
}

func (e *CodeMessageError) Error() string {
	return fmt.Sprintf("code: %s, message: %s", e.code, e.message)
}

type SendOption struct {
	// 等待响应的超时时间，为0代表不设置超时时间
	Timeout time.Duration
}

type RequestOption struct {
	// 等待响应的超时时间，为0代表不设置超时时间
	Timeout time.Duration
}

type RabbitMQClient struct {
	connection *amqp.Connection
	channel    *amqp.Channel
	exchange   string
	key        string

	requests     map[string]*future.SetValueFuture[*Message]
	requestsLock sync.Mutex

	closed chan any
}

func NewRabbitMQClient(url string, key string, exchange string) (*RabbitMQClient, error) {
	connection, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("connecting to %s: %w", url, err)
	}

	channel, err := connection.Channel()
	if err != nil {
		connection.Close()
		return nil, fmt.Errorf("openning channel on connection: %w", err)
	}

	cli := &RabbitMQClient{
		connection: connection,
		channel:    channel,
		exchange:   exchange,
		key:        key,
		requests:   make(map[string]*future.SetValueFuture[*Message]),
		closed:     make(chan any),
	}

	// NOTE! 经测试发现，必须在Publish之前调用Consume进行消息接收，否则Consume会返回错误
	// 因此这段代码不能移动到serve函数中，必须放在这里，保证顺序
	recvChan, err := channel.Consume(
		// 一个特殊队列，服务端的回复消息都会发送到这个队列里
		DIRECT_REPLY_TO,
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

func (cli *RabbitMQClient) Request(req Message, opts ...RequestOption) (*Message, error) {
	opt := RequestOption{Timeout: time.Second * 15}
	if len(opts) > 0 {
		opt = opts[0]
	}

	reqID := req.MakeRequestID()
	fut := future.NewSetValue[*Message]()

	cli.requestsLock.Lock()
	cli.requests[reqID] = fut
	cli.requestsLock.Unlock()

	// 启动超时定时器
	if opt.Timeout != 0 {
		go func() {
			<-time.After(opt.Timeout)
			cli.requestsLock.Lock()
			// 由于只会在requestsLock.Lock()之后修改fut的状态，所以Complete的判断是可信的
			if !fut.IsComplete() {
				fut.SetError(fmt.Errorf("wait response timeout"))
			}
			delete(cli.requests, reqID)
			cli.requestsLock.Unlock()
		}()
	}

	err := cli.Send(req, SendOption{
		Timeout: opt.Timeout,
	})
	if err != nil {
		cli.requestsLock.Lock()
		delete(cli.requests, reqID)
		cli.requestsLock.Unlock()

		return nil, fmt.Errorf("sending message: %w", err)
	}

	resp, err := fut.WaitValue()
	if err != nil {
		return nil, fmt.Errorf("requesting: %w", err)
	}

	return resp, nil
}

func (c *RabbitMQClient) Send(msg Message, opts ...SendOption) error {
	opt := SendOption{}
	if len(opts) > 0 {
		opt = opts[0]
	}

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
		ReplyTo:    DIRECT_REPLY_TO,
		Expiration: expiration,
	})

	if err != nil {
		return fmt.Errorf("publishing data: %w", err)
	}

	return nil
}

func (c *RabbitMQClient) serve(recvChan <-chan amqp.Delivery) error {
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
				c.requestsLock.Lock()
				if req, ok := c.requests[reqID]; ok {
					req.SetValue(msg)
					delete(c.requests, reqID)
				}
				c.requestsLock.Unlock()
			}

		case <-c.closed:
			return nil
		}
	}
}

func (c *RabbitMQClient) Close() error {
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
func Request[TResp any, TReq any](cli *RabbitMQClient, req TReq, opts ...RequestOption) (*TResp, error) {
	resp, err := cli.Request(MakeMessage(req), opts...)
	if err != nil {
		return nil, fmt.Errorf("requesting: %w", err)
	}

	errCode, errMsg := resp.GetCodeMessage()
	if errCode != errorcode.OK {
		return nil, &CodeMessageError{
			code:    errCode,
			message: errMsg,
		}
	}

	respBody, ok := resp.Body.(TResp)
	if !ok {
		return nil, fmt.Errorf("expect a %s body, but got %s",
			myreflect.ElemTypeOf[TResp]().Name(),
			myreflect.TypeOfValue(resp.Body).Name())
	}

	return &respBody, nil
}

// 发送消息，不等待回应
func Send[TReq any](cli *RabbitMQClient, msg TReq, opts ...SendOption) error {
	req := MakeMessage(msg)

	err := cli.Send(req, opts...)
	if err != nil {
		return fmt.Errorf("sending: %w", err)
	}

	return nil
}
