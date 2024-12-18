package utils

import (
	"fmt"
	"strings"

	"gitlink.org.cn/cloudream/common/pkgs/ioswitch/exec"
)

func FormatVarIDs(arr []exec.VarID) string {
	sb := strings.Builder{}
	for i, v := range arr {
		sb.WriteString(fmt.Sprintf("%v", v))
		if i < len(arr)-1 {
			sb.WriteString(",")
		}
	}
	return sb.String()
}
