package cmdtrie

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_CommandTrie(t *testing.T) {
	Convey("无参数命令", t, func() {
		trie := NewCommandTrie[int]()

		var ret string

		trie.Add(func() {
			ret = "ok"
		}, "a")

		err := trie.Execute(0, "a")
		So(err, ShouldBeNil)

		So(ret, ShouldEqual, "ok")
	})

	Convey("各种参数", t, func() {
		trie := NewCommandTrie[int]()

		var argI int
		var argStr string
		var argBl bool
		var argFP float32

		trie.Add(func(i int, str string, bl bool, fp float32) {
			argI = i
			argStr = str
			argBl = bl
			argFP = fp

		}, "a", "b")

		err := trie.Execute(0, "a", "b", "1", "2", "true", "3")
		So(err, ShouldBeNil)

		So(argI, ShouldEqual, 1)
		So(argStr, ShouldEqual, "2")
		So(argBl, ShouldBeTrue)
		So(argFP, ShouldEqual, 3)
	})

	Convey("有数组参数", t, func() {
		trie := NewCommandTrie[int]()

		var argI int
		var argArr []int64

		trie.Add(func(i int, arr []int64) {
			argI = i
			argArr = arr

		}, "a", "b")

		err := trie.Execute(0, "a", "b", "1", "2", "3", "4")
		So(err, ShouldBeNil)

		So(argI, ShouldEqual, 1)
		So(argArr, ShouldResemble, []int64{2, 3, 4})
	})

	Convey("有数组参数，但为空", t, func() {
		trie := NewCommandTrie[int]()

		var argI int
		var argArr []int64

		trie.Add(func(i int, arr []int64) {
			argI = i
			argArr = arr

		}, "a", "b")

		err := trie.Execute(0, "a", "b", "1")
		So(err, ShouldBeNil)

		So(argI, ShouldEqual, 1)
		So(argArr, ShouldResemble, []int64{})
	})
}
