package prescheduler

import (
	"testing"

	"github.com/samber/lo"
	. "github.com/smartystreets/goconvey/convey"

	schsdk "gitlink.org.cn/cloudream/common/sdks/scheduler"
)

func TestOrderByAfters(t *testing.T) {
	cases := []struct {
		title string
		jobs  []*schedulingJob
		wants []string
	}{
		{
			title: "所有Job都有直接或间接的依赖关系",
			jobs: []*schedulingJob{
				{
					Job:    &schsdk.NormalJobInfo{JobInfoBase: schsdk.JobInfoBase{LocalJobID: "1"}},
					Afters: []string{"2"},
				},

				{
					Job:    &schsdk.NormalJobInfo{JobInfoBase: schsdk.JobInfoBase{LocalJobID: "2"}},
					Afters: []string{},
				},

				{
					Job:    &schsdk.NormalJobInfo{JobInfoBase: schsdk.JobInfoBase{LocalJobID: "3"}},
					Afters: []string{"1"},
				},
			},
			wants: []string{"2", "1", "3"},
		},

		{
			title: "部分Job之间无依赖关系",
			jobs: []*schedulingJob{
				{
					Job:    &schsdk.NormalJobInfo{JobInfoBase: schsdk.JobInfoBase{LocalJobID: "1"}},
					Afters: []string{"2"},
				},

				{
					Job:    &schsdk.NormalJobInfo{JobInfoBase: schsdk.JobInfoBase{LocalJobID: "2"}},
					Afters: []string{},
				},

				{
					Job:    &schsdk.NormalJobInfo{JobInfoBase: schsdk.JobInfoBase{LocalJobID: "3"}},
					Afters: []string{"1"},
				},

				{
					Job:    &schsdk.NormalJobInfo{JobInfoBase: schsdk.JobInfoBase{LocalJobID: "4"}},
					Afters: []string{"5"},
				},

				{
					Job:    &schsdk.NormalJobInfo{JobInfoBase: schsdk.JobInfoBase{LocalJobID: "5"}},
					Afters: []string{},
				},
			},
			wants: []string{"2", "5", "1", "3", "4"},
		},

		{
			title: "存在循环依赖",
			jobs: []*schedulingJob{
				{
					Job:    &schsdk.NormalJobInfo{JobInfoBase: schsdk.JobInfoBase{LocalJobID: "1"}},
					Afters: []string{"2"},
				},

				{
					Job:    &schsdk.NormalJobInfo{JobInfoBase: schsdk.JobInfoBase{LocalJobID: "2"}},
					Afters: []string{"1"},
				},
			},
			wants: nil,
		},

		{
			title: "完全不依赖",
			jobs: []*schedulingJob{
				{
					Job:    &schsdk.NormalJobInfo{JobInfoBase: schsdk.JobInfoBase{LocalJobID: "1"}},
					Afters: []string{},
				},

				{
					Job:    &schsdk.NormalJobInfo{JobInfoBase: schsdk.JobInfoBase{LocalJobID: "2"}},
					Afters: []string{},
				},
			},
			wants: []string{"1", "2"},
		},
	}

	sch := NewDefaultPreScheduler()
	for _, c := range cases {
		Convey(c.title, t, func() {
			ordered, ok := sch.orderByAfters(c.jobs)
			if c.wants == nil {
				So(ok, ShouldBeFalse)
			} else {
				So(ok, ShouldBeTrue)

				ids := lo.Map(ordered, func(item *schedulingJob, idx int) string { return item.Job.GetLocalJobID() })
				So(ids, ShouldResemble, c.wants)
			}
		})
	}
}
