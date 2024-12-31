package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/imdario/mergo"
)

// Load 从本地文件读取配置，加载配置文件
func Load(filePath string, cfg interface{}) error {
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	// json.Unmarshal用于将JSON解码成结构体
	return json.Unmarshal(fileData, cfg)
}

// DefaultLoad 默认的加载配置的方式：
// 从应用程序上上级的conf目录中读取，文件名：<moduleName>.config.json
func DefaultLoad(modeulName string, defCfg interface{}) error {
	// 获取当前进程文件执行路径，并判断是否为空
	execPath, err := os.Executable()
	if err != nil {
		return err
	}

	// TODO 可以考虑根据环境变量读取不同的配置
	// filepath.Join用于将多个路径组合成一个路径
	configFilePath := filepath.Join(filepath.Dir(execPath), "..", "confs", fmt.Sprintf("%s.config.json", modeulName))

	return Load(configFilePath, defCfg)
}

// Merge 合并两个配置结构体。会将src中的非空字段覆盖到dst的同名字段中。两个结构的类型必须相同
func Merge(dst interface{}, src interface{}) error {
	return mergo.Merge(dst, src)
}
