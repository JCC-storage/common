package mq

import (
	"testing"
	"time"

	"github.com/google/uuid"
	. "github.com/smartystreets/goconvey/convey"
)

func Test_ServerClient(t *testing.T) {
	Convey("心跳", t, func() {
		type Msg struct {
			MessageBodyBase
			Data int64
		}
		RegisterMessage[Msg]()

		rabbitURL := "amqp://cloudream:123456@127.0.0.1:5672/"
		testQueue := "Test" + uuid.NewString()

		svr, err := NewRabbitMQServer(rabbitURL, testQueue,
			func(msg *Message) (*Message, error) {
				<-time.After(time.Second * 10)
				reply := MakeAppDataMessage(&Msg{Data: 1})
				return &reply, nil
			})
		So(err, ShouldBeNil)

		go svr.Serve()

		cli, err := NewRabbitMQClient(rabbitURL, testQueue, "")
		So(err, ShouldBeNil)

		_, err = cli.Request(MakeAppDataMessage(&Msg{}), RequestOption{
			Timeout: time.Second * 2,
		})
		So(err, ShouldEqual, ErrWaitResponseTimeout)

		reply, err := cli.Request(MakeAppDataMessage(&Msg{}), RequestOption{
			Timeout:   time.Second * 2,
			KeepAlive: true,
		})
		So(err, ShouldBeNil)

		msgReply, ok := reply.Body.(*Msg)
		So(ok, ShouldBeTrue)
		So(msgReply.Data, ShouldEqual, 1)

		svr.Close()
		cli.Close()
	})
}
