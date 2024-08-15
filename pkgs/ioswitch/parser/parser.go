package parser

import (
	"gitlink.org.cn/cloudream/common/pkgs/ioswitch/exec"
)

type FromToParser interface {
	Parse(ft FromTo, blder *exec.PlanBuilder) error
}
