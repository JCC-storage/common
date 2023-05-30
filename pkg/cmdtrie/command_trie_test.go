package cmdtrie

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_CommandTrie(t *testing.T) {
	Convey("无参数命令", t, func() {
		trie := NewVoidCommandTrie[int]()

		var ret string

		err := trie.Add(func(int) {
			ret = "ok"
		}, "a")
		So(err, ShouldBeNil)

		err = trie.Execute(0, []string{"a"})
		So(err, ShouldBeNil)

		So(ret, ShouldEqual, "ok")
	})

	Convey("各种参数", t, func() {
		trie := NewVoidCommandTrie[int]()

		var argI int
		var argStr string
		var argBl bool
		var argFP float32

		err := trie.Add(func(int, i int, str string, bl bool, fp float32) {
			argI = i
			argStr = str
			argBl = bl
			argFP = fp

		}, "a", "b")
		So(err, ShouldBeNil)

		err = trie.Execute(0, []string{"a", "b", "1", "2", "true", "3"})
		So(err, ShouldBeNil)

		So(argI, ShouldEqual, 1)
		So(argStr, ShouldEqual, "2")
		So(argBl, ShouldBeTrue)
		So(argFP, ShouldEqual, 3)
	})

	Convey("有数组参数", t, func() {
		trie := NewVoidCommandTrie[int]()

		var argI int
		var argArr []int64

		err := trie.Add(func(int, i int, arr []int64) {
			argI = i
			argArr = arr

		}, "a", "b")
		So(err, ShouldBeNil)

		err = trie.Execute(0, []string{"a", "b", "1", "2", "3", "4"})
		So(err, ShouldBeNil)

		So(argI, ShouldEqual, 1)
		So(argArr, ShouldResemble, []int64{2, 3, 4})
	})

	Convey("有数组参数，但为空", t, func() {
		trie := NewVoidCommandTrie[int]()

		var argI int
		var argArr []int64

		err := trie.Add(func(int, i int, arr []int64) {
			argI = i
			argArr = arr

		}, "a", "b")
		So(err, ShouldBeNil)

		err = trie.Execute(0, []string{"a", "b", "1"})
		So(err, ShouldBeNil)

		So(argI, ShouldEqual, 1)
		So(argArr, ShouldResemble, []int64{})
	})

	Convey("带返回值", t, func() {
		trie := NewCommandTrie[int, int]()

		var argI int
		var argArr []int64

		err := trie.Add(func(int, i int, arr []int64) int {
			argI = i
			argArr = arr
			return 123
		}, "a", "b")
		So(err, ShouldBeNil)

		ret, err := trie.Execute(0, []string{"a", "b", "1"})
		So(err, ShouldBeNil)

		So(argI, ShouldEqual, 1)
		So(argArr, ShouldResemble, []int64{})
		So(ret, ShouldEqual, 123)
	})

	Convey("返回值是接口类型", t, func() {
		trie := NewCommandTrie[int, any]()

		var argI int
		var argArr []int64

		err := trie.Add(func(int, i int, arr []int64) int {
			argI = i
			argArr = arr
			return 123
		}, "a", "b")
		So(err, ShouldBeNil)

		err = trie.Add(func(int, i int, arr []int64) string {
			return "123"
		}, "a", "c")
		So(err, ShouldBeNil)

		ret, err := trie.Execute(0, []string{"a", "b", "1"})
		So(err, ShouldBeNil)
		So(argI, ShouldEqual, 1)
		So(argArr, ShouldResemble, []int64{})
		So(ret, ShouldEqual, 123)

		ret2, err := trie.Execute(0, []string{"a", "c", "1"})
		So(err, ShouldBeNil)
		So(ret2, ShouldEqual, "123")
	})

	Convey("无Ctx参数", t, func() {
		trie := NewStaticCommandTrie[int]()

		var argI int
		var argArr []int64

		err := trie.Add(func(i int, arr []int64) int {
			argI = i
			argArr = arr
			return 123
		}, "a", "b")
		So(err, ShouldBeNil)

		ret, err := trie.Execute([]string{"a", "b", "1"})
		So(err, ShouldBeNil)

		So(argI, ShouldEqual, 1)
		So(argArr, ShouldResemble, []int64{})
		So(ret, ShouldEqual, 123)
	})

	Convey("完全无参数", t, func() {
		trie := NewStaticCommandTrie[int]()

		var argI int
		var argArr []int64

		err := trie.Add(func() int {
			argI = 1
			argArr = []int64{}
			return 123
		}, "a", "b")
		So(err, ShouldBeNil)

		ret, err := trie.Execute([]string{"a", "b"})
		So(err, ShouldBeNil)

		So(argI, ShouldEqual, 1)
		So(argArr, ShouldResemble, []int64{})
		So(ret, ShouldEqual, 123)
	})

	Convey("空数组参数变成nil", t, func() {
		trie := NewStaticCommandTrie[int]()

		var argStrs []string

		err := trie.Add(func(strs []string) int {
			argStrs = strs
			return 123
		}, "a", "b")
		So(err, ShouldBeNil)

		ret, err := trie.Execute([]string{"a", "b"}, ExecuteOption{ReplaceEmptyArrayWithNil: true})
		So(err, ShouldBeNil)

		So(argStrs, ShouldBeNil)
		So(ret, ShouldEqual, 123)
	})
}
