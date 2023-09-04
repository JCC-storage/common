package future

import (
	"context"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_SetVoidFuture(t *testing.T) {
	wait := func(fut *SetVoidFuture) bool {
		futChan := make(chan any)

		go func() {
			fut.Wait(context.Background())
			close(futChan)
		}()

		select {
		case <-time.After(time.Second * 5):
			return false
		case <-futChan:
			return true
		}
	}

	waitTimeout := func(fut *SetVoidFuture, timeoutMs int) bool {
		futChan := make(chan any)

		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*time.Duration(timeoutMs))
			defer cancel()
			fut.Wait(ctx)
			close(futChan)
		}()

		select {
		case <-time.After(time.Second * 5):
			return false
		case <-futChan:
			return true
		}
	}

	Convey("正常返回", t, func() {
		fut := NewSetVoid()
		fut.SetVoid()
		ok := wait(fut)

		So(ok, ShouldBeTrue)
	})

	Convey("超时返回", t, func() {
		fut := NewSetVoid()
		ok := waitTimeout(fut, 2000)

		So(ok, ShouldBeTrue)
	})

	Convey("没有返回", t, func() {
		fut := NewSetVoid()
		ok := wait(fut)

		So(ok, ShouldBeFalse)
	})
}
