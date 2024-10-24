package mq

import (
	"fmt"
	"net"
	"time"

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

// 处理消息。会将第一个返回值作为响应回复给客户端，如果为nil，则不回复。
type MessageHandlerFn func(msg *Message) (*Message, error)

type RabbitMQServer struct {
	queueName  string
	connection *amqp.Connection
	channel    *amqp.Channel
	closed     chan any

	OnMessage MessageHandlerFn
	OnError   func(err error)
}

func NewRabbitMQServer(url string, queueName string, onMessage MessageHandlerFn) (*RabbitMQServer, error) {
	config := amqp.Config{
		Dial: func(network, addr string) (net.Conn, error) {
			return net.DialTimeout(network, addr, 60*time.Second) // 设置连接超时时间为 60 秒
		},
	}
	connection, err := amqp.DialConfig(url, config)

	//connection, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("connecting to %s: %w", url, err)
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
	}

	return srv, nil
}

func (s *RabbitMQServer) Serve() error {
	_, err := s.channel.QueueDeclare(
		s.queueName,
		false,
		true,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("declare queue failed, err: %w", err)
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
		return fmt.Errorf("open consume channel failed, err: %w", err)
	}

	for {
		select {
		case rawReq, ok := <-channel:
			if !ok {
				s.onError(NewDeserializeError(fmt.Errorf("channel is closed")))
				return NewReceiveMessageError(fmt.Errorf("channel is closed"))
			}

			reqMsg, err := Deserialize(rawReq.Body)
			if err != nil {
				s.onError(NewDeserializeError(err))
				continue
			}

			go s.handleMessage(reqMsg, rawReq)

		case <-s.closed:
			return nil
		}
	}
}

func (s *RabbitMQServer) handleMessage(reqMsg *Message, rawReq amqp.Delivery) {
	replyed := make(chan bool)
	defer close(replyed)

	keepAliveTimeoutMs := reqMsg.GetKeepAlive()
	if keepAliveTimeoutMs != 0 {
		go s.keepAlive(keepAliveTimeoutMs, reqMsg, rawReq, replyed)
	}

	reply, err := s.OnMessage(reqMsg)
	if err != nil {
		s.onError(NewDispatchError(err))
		return
	}

	if reply != nil {
		reply.SetRequestID(reqMsg.GetRequestID())
		err := s.replyToClient(*reply, &rawReq)
		if err != nil {
			s.onError(NewReplyError(err))
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
