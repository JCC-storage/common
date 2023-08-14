package trie

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_CommandTrie(t *testing.T) {
	Convey("全是Word节点", t, func() {
		trie := Trie[int]{}

		{
			n := trie.Create([]any{"a", "b"})
			So(n, ShouldNotBeNil)
			n.Value = 123
		}

		{
			n, ok := trie.WalkEnd([]string{"a", "b"})
			So(n, ShouldNotBeNil)
			So(ok, ShouldBeTrue)
			So(n.Value, ShouldEqual, 123)
		}
	})

	Convey("包含Any节点", t, func() {
		trie := Trie[int]{}

		{
			n := trie.Create([]any{"a", WORD_ANY, "b"})
			So(n, ShouldNotBeNil)
			n.Value = 123
		}

		{
			n, ok := trie.WalkEnd([]string{"a", "11", "b"})
			So(n, ShouldNotBeNil)
			So(ok, ShouldBeTrue)
			So(n.Value, ShouldEqual, 123)
		}

		{
			n, ok := trie.WalkEnd([]string{"a", "22", "b"})
			So(n, ShouldNotBeNil)
			So(ok, ShouldBeTrue)
			So(n.Value, ShouldEqual, 123)
		}
	})

	Convey("优先经过Word节点", t, func() {
		trie := Trie[int]{}

		{
			n := trie.Create([]any{"a", "b", "c"})
			So(n, ShouldNotBeNil)
			n.Value = 123
		}

		{
			n := trie.Create([]any{"a", WORD_ANY, "c"})
			So(n, ShouldNotBeNil)
			n.Value = 456
		}

		{
			n, ok := trie.WalkEnd([]string{"a", "b", "c"})
			So(n, ShouldNotBeNil)
			So(ok, ShouldBeTrue)
			So(n.Value, ShouldEqual, 123)
		}

		{
			n, ok := trie.WalkEnd([]string{"a", "d", "c"})
			So(n, ShouldNotBeNil)
			So(ok, ShouldBeTrue)
			So(n.Value, ShouldEqual, 456)
		}
	})
}
