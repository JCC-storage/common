package http

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_objectToStringMap(t *testing.T) {
	Convey("包含指针", t, func() {
		type A struct {
			Val  *int `json:"Val,omitempty"`
			Nil  *int `json:"Nil,omitempty"`
			Omit *int `json:"Omit"`
		}

		v := 10
		a := A{
			Val:  &v,
			Nil:  nil,
			Omit: nil,
		}

		mp, err := objectToStringMap(a)
		So(err, ShouldBeNil)

		So(mp, ShouldResemble, map[string]string{
			"Val":  "10",
			"Omit": "",
		})
	})
}
