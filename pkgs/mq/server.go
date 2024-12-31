package mq

import (
	"fmt"
	"net"
	"time"

	"gitlink.org.cn/cloudream/common/utils/sync2"

	"github.com/streadway/amqp"
)

type ReceiveMessageError struct {
	err error
}

func (err ReceiveMessageError) Error() string {
	return fmt.Sprintf("receive message error: %s", err.err.Error())
}

func NewReceiveMessageError(err error) ReceiveMessageError {
	return ReceiveMessageError{
		err: err,
	}
}

type DeserializeError struct {
	err error
}

func (err DeserializeError) Error() string {
	return fmt.Sprintf("deserialize error: %s", err.err.Error())
}

func NewDeserializeError(err error) DeserializeError {
	return DeserializeError{
		err: err,
	}
}

type DispatchError struct {
	err error
}

func (err DispatchError) Error() string {
	return fmt.Sprintf("dispatch error: %s", err.err.Error())
}

func NewDispatchError(err error) DispatchError {
	return DispatchError{
		err: err,
	}
}

type ReplyError struct {
	err error
}

func (err ReplyError) Error() string {
	return fmt.Sprintf("replay to client : %s", err.err.Error())
}

func NewReplyError(err error) ReplyError {
	return ReplyError{
		err: err,
	}
}

type ServerExit struct {
	Error error
}

type RabbitMQServerEvent interface{}

// 处理消息。会将第一个返回值作为响应回复给客户端，如果为nil，则不回复。
type MessageHandlerFn func(msg *Message) (*Message, error)

type RabbitMQServer struct {
	queueName  string
	connection *amqp.Connection
	channel    *amqp.Channel
	closed     chan any
	config     Config

	OnMessage MessageHandlerFn
	OnError   func(err error)
}

type RabbitMQParam struct {
	RetryNum      int `json:"retryNum"`
	RetryInterval int `json:"retryInterval"`
}

func NewRabbitMQServer(cfg Config, queueName string, onMessage MessageHandlerFn) (*RabbitMQServer, error) {
	config := amqp.Config{
		Vhost: cfg.VHost,
		Dial: func(network, addr string) (net.Conn, error) {
			return net.DialTimeout(network, addr, 60*time.Second) // 设置连接超时时间为 60 秒
		},
	}
	connection, err := amqp.DialConfig(fmt.Sprintf("amqp://%s:%s@%s", cfg.Account, cfg.Password, cfg.Address), config)

	if err != nil {
		return nil, fmt.Errorf("connecting to %s: %w", cfg.Address, err)
	}

	channel, err := connection.Channel()
	if err != nil {
		connection.Close()
		return nil, fmt.Errorf("openning channel on connection: %w", err)
	}

	srv := &RabbitMQServer{
		connection: connection,
		channel:    channel,
		queueName:  queueName,
		closed:     make(chan any),
		OnMessage:  onMessage,
		config:     cfg,
	}

	return srv, nil
}

func (s *RabbitMQServer) Start() *sync2.UnboundChannel[RabbitMQServerEvent] {
	ch := sync2.NewUnboundChannel[RabbitMQServerEvent]()

	channel := s.openChannel(ch)
	if channel == nil {
		ch.Send(1)
		return ch
	}

	retryNum := 0

	for {
		select {
		case rawReq, ok := <-channel:
			if !ok {
				if retryNum > s.config.Param.RetryNum {
					ch.Send(ServerExit{Error: fmt.Errorf("maximum number of retries exceeded")})
					return ch
				}
				retryNum++

				time.Sleep(time.Duration(s.config.Param.RetryInterval) * time.Millisecond)
				channel = s.openChannel(ch)
				continue
			}

			reqMsg, err := Deserialize(rawReq.Body)
			if err != nil {
				ch.Send(NewDeserializeError(err))
				continue
			}

			go s.handleMessage(ch, reqMsg, rawReq)

		case <-s.closed:
			return nil
		}
	}
}

func (s *RabbitMQServer) openChannel(ch *sync2.UnboundChannel[RabbitMQServerEvent]) <-chan amqp.Delivery {
	_, err := s.channel.QueueDeclare(
		s.queueName,
		false,
		true,
		false,
		false,
		nil,
	)
	if err != nil {
		ch.Send(fmt.Errorf("declare queue failed, err: %w", err))
		return nil
	}

	channel, err := s.channel.Consume(
		s.queueName,
		"",
		true,
		false,
		true,
		false,
		nil,
	)

	if err != nil {
		ch.Send(fmt.Errorf("get rabbitmq channel failed, err: %w", err))
		return nil
	}

	return channel
}

func (s *RabbitMQServer) handleMessage(ch *sync2.UnboundChannel[RabbitMQServerEvent], reqMsg *Message, rawReq amqp.Delivery) {
	replyed := make(chan bool)
	defer close(replyed)

	keepAliveTimeoutMs := reqMsg.GetKeepAlive()
	if keepAliveTimeoutMs != 0 {
		go s.keepAlive(keepAliveTimeoutMs, reqMsg, rawReq, replyed)
	}

	reply, err := s.OnMessage(reqMsg)
	if err != nil {
		ch.Send(NewDispatchError(err))
		return
	}

	if reply != nil {
		reply.SetRequestID(reqMsg.GetRequestID())
		err := s.replyToClient(*reply, &rawReq)
		if err != nil {
			ch.Send(NewReplyError(err))
		}
	}
}

func (s *RabbitMQServer) keepAlive(keepAliveTimeoutMs int, reqMsg *Message, rawReq amqp.Delivery, replyed chan bool) {
	ticker := time.NewTicker(time.Millisecond * time.Duration(keepAliveTimeoutMs))
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			hbMsg := MakeHeartbeatMessage()
			hbMsg.SetRequestID(reqMsg.GetRequestID())

			err := s.replyToClient(hbMsg, &rawReq)
			if err != nil {
				s.onError(NewReplyError(err))
			}

		case <-replyed:
			return
		}
	}
}

func (s *RabbitMQServer) Close() {
	close(s.closed)
}

func (s *RabbitMQServer) onError(err error) {
	if s.OnError != nil {
		s.OnError(err)
	}
}

// replyToClient 回复客户端的消息，需要用到客户端发来的消息中的字段来判断回到哪个队列
func (s *RabbitMQServer) replyToClient(reply Message, reqMsg *amqp.Delivery) error {
	msgData, err := Serialize(reply)
	if err != nil {
		return fmt.Errorf("serialize message failed: %w", err)
	}

	return s.channel.Publish(
		"",
		reqMsg.ReplyTo,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        msgData,
			Expiration:  "30000", // 响应消息的超时时间默认30秒
		},
	)
}
