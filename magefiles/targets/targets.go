package targets

import (
	"gitlink.org.cn/cloudream/common/magefiles"
)

// [配置项]设置编译平台为windows
func Win() {
	magefiles.Global.OS = "win"
}

// [配置项]设置编译平台为linux
func Linux() {
	magefiles.Global.OS = "linux"
}

// [配置项]设置编译架构为amd64
func AMD64() {
	magefiles.Global.Arch = "amd64"
}
