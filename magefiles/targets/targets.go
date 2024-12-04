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

// [配置项]设置编译架构为arm64
func ARM64() {
	magefiles.Global.Arch = "arm64"
}

// [配置项]设置编译的根目录
func BuildRoot(dir string) {
	magefiles.Global.BuildRoot = dir
}

// [配置项]关闭编译优化，用于调试
func Debug() {
	magefiles.Global.Debug = true
}
