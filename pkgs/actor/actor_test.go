package actor

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_CommandChannel(t *testing.T) {
	wait := func(ch <-chan CommandFn) (CommandFn, bool) {
		select {
		case <-time.After(time.Second * 5):
			return nil, false
		case cmd := <-ch:
			return cmd, true
		}
	}

	Convey("BeginChanReceive", t, func() {
		cmdChan := NewCommandChannel()

		cmdChan.Send(func() {})

		ch := cmdChan.BeginChanReceive()
		defer cmdChan.CloseChanReceive()

		_, ok := wait(ch)

		So(ok, ShouldBeTrue)
	})
}
