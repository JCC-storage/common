package logger

type Config struct {
	Output          string `json:"output"`          // 输出日志的方式。file：输出到文件，stdout：输出到标准输出
	OutputFileName  string `json:"outputFileName"`  // 输出日志的文件名，只在Output字段为file时有意义
	OutputDirectory string `json:"outputDirectory"` // 输出日志的目录，只在Output字段为file时有意义
	Level           string `json:"level"`
}
