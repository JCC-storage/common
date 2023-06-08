package distlock

/*
import (
	. "github.com/smartystreets/goconvey/convey"
)

func Test_parseLockData_lockDataToString(t *testing.T) {
	cases := []struct {
		title string
		data  lockData
	}{
		{
			title: "多段路径",
			data: lockData{
				Path:   []string{"a", "b", "c"},
				Name:   "d",
				Target: "e",
			},
		},

		{
			title: "包含分隔符",
			data: lockData{
				Path:   []string{"a/", "b", "c/c"},
				Name:   "/d",
				Target: "///e//d/",
			},
		},

		{
			title: "包含转义符",
			data: lockData{
				Path:   []string{"a\\/", "b", "\\c/c"},
				Name:   "/d",
				Target: "///e\\//d/\\",
			},
		},

		{
			title: "包含换行符",
			data: lockData{
				Path:   []string{"a\n", "\nb", "c\nc"},
				Name:   "/d",
				Target: "e\nd\n",
			},
		},
	}

	for _, ca := range cases {
		Convey(ca.title, t, func() {
			str := lockDataToString(ca.data)

			data, err := parseLockData(str)

			So(err, ShouldBeNil)
			So(data, ShouldResemble, ca.data)
		})
	}
}
*/
