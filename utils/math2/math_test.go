package math2

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_SplitLessThan(t *testing.T) {
	checker := func(t *testing.T, arr []int, total int, maxValue int) {
		t.Logf("arr: %v, total: %d, maxValue: %d", arr, total, maxValue)

		sum := 0
		for _, v := range arr {
			sum += v

			if v > maxValue {
				t.Errorf("value should be less than %d", maxValue)
			}
		}

		if sum != total {
			t.Errorf("sum should be %d", total)
		}
	}

	Convey("测试", t, func() {
		checker(t, SplitLessThan(9, 9), 9, 9)
		checker(t, SplitLessThan(9, 3), 9, 3)
		checker(t, SplitLessThan(10, 3), 10, 3)
		checker(t, SplitLessThan(11, 3), 11, 3)
		checker(t, SplitLessThan(12, 3), 12, 3)
	})
}

func Test_SplitN(t *testing.T) {
	checker := func(t *testing.T, arr []int, total int) {
		t.Logf("arr: %v, total: %d", arr, total)

		sum := 0
		for _, v := range arr {
			sum += v
		}

		if sum != total {
			t.Errorf("sum should be %d", total)
		}
	}

	Convey("测试", t, func() {
		checker(t, SplitN(9, 9), 9)
		checker(t, SplitN(9, 3), 9)
		checker(t, SplitN(10, 3), 10)
		checker(t, SplitN(11, 3), 11)
		checker(t, SplitN(12, 3), 12)
	})
}
