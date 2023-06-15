package lo

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_Remove(t *testing.T) {
	Convey("删除数组元素", t, func() {
		arr := []string{"a", "b", "c"}
		arr = Remove(arr, "b")
		So(arr, ShouldResemble, []string{"a", "c"})
	})

	Convey("删除最后一个元素", t, func() {
		arr := []string{"a", "b", "c"}
		arr = Remove(arr, "c")
		So(arr, ShouldResemble, []string{"a", "b"})
	})

	Convey("删除第一个元素", t, func() {
		arr := []string{"a", "b", "c"}
		arr = Remove(arr, "a")
		So(arr, ShouldResemble, []string{"b", "c"})
	})

	Convey("删除不存在的元素", t, func() {
		arr := []string{"a", "b", "c"}
		arr = Remove(arr, "d")
		So(arr, ShouldResemble, []string{"a", "b", "c"})
	})
}

func Test_ArrayClone(t *testing.T) {
	Convey("复制数组", t, func() {
		arr := []string{"a", "b", "c"}
		arr2 := ArrayClone(arr)

		arr2[1] = "a"

		So(arr, ShouldResemble, []string{"a", "b", "c"})
		So(arr2, ShouldResemble, []string{"a", "a", "c"})
	})

	Convey("复制出来的数组进行append", t, func() {
		arr := []string{"a", "b", "c"}
		arr2 := ArrayClone(arr)

		arr = append(arr, "d")
		arr2 = append(arr2, "c")

		So(arr, ShouldResemble, []string{"a", "b", "c", "d"})
		So(arr2, ShouldResemble, []string{"a", "b", "c", "c"})
	})
}
