package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
	myreflect "gitlink.org.cn/cloudream/common/utils/reflect"
)

// Std 是一个输出日志到标准输出的Logger，适用于没有设计好日志输出方案时的临时使用。
var Std Logger

// init 初始化包，设置日志格式为不带颜色的Nested格式，日志级别为Debug，输出到标准输出。
func init() {
	logger := logrus.New()
	logger.SetFormatter(&nested.Formatter{
		TimestampFormat: "2006-01-02 15:04:05",
		NoColors:        true,
		NoFieldsColors:  true,
	})

	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetOutput(os.Stdout)
	Std = &logrusLogger{entry: logger.WithField("TODO", "")}
}

// Init 初始化全局默认的日志器，根据配置设置日志级别和输出位置。
//
// 参数:
//
//	cfg *Config: 日志配置项，包括日志级别和输出位置等。
//
// 返回值:
//
//	error: 初始化过程中的任何错误。
func Init(cfg *Config) error {
	logrus.SetFormatter(&nested.Formatter{
		TimestampFormat: "2006-01-02 15:04:05",
		NoColors:        true,
		NoFieldsColors:  true,
	})

	// 设置日志级别
	level, ok := loggerLevels[strings.ToUpper(cfg.Level)]
	if !ok {
		return fmt.Errorf("invalid log level: %s", cfg.Level)
	}

	logrus.SetLevel(level)

	// 设置日志输出位置
	output := strings.ToUpper(cfg.Output)

	if output == OUTPUT_FILE {
		logFilePath := filepath.Join(cfg.OutputDirectory, cfg.OutputFileName+".log")

		// 创建日志文件所在的目录
		if err := os.MkdirAll(cfg.OutputDirectory, 0755); err != nil {
			return err
		}

		// 打开或创建日志文件
		file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0755)
		if err != nil {
			return err
		}
		logrus.SetOutput(file)

	} else if output == OUTPUT_STDOUT {
		logrus.SetOutput(os.Stdout)
	} else {
		logrus.SetOutput(os.Stdout)
		logrus.Warnf("unsupported output: %s, will output to stdout", output)
	}

	return nil
}

// 下面是日志记录的方法，它们分别对应不同的日志级别和格式。
// 这些方法最终都会调用logrus对应的方法来记录日志。

func Debug(args ...interface{}) {
	logrus.Debug(args...)
}

func Debugf(format string, args ...interface{}) {
	logrus.Debugf(format, args...)
}

func Info(args ...interface{}) {
	logrus.Info(args...)
}

func Infof(format string, args ...interface{}) {
	logrus.Infof(format, args...)
}

func Warn(args ...interface{}) {
	logrus.Warn(args...)
}

func Warnf(format string, args ...interface{}) {
	logrus.Warnf(format, args...)
}

func Error(args ...interface{}) {
	logrus.Error(args...)
}

func Errorf(format string, args ...interface{}) {
	logrus.Errorf(format, args...)
}

func Fatal(args ...interface{}) {
	logrus.Fatal(args...)
}

func Fatalf(format string, args ...interface{}) {
	logrus.Fatalf(format, args...)
}

func Panic(args ...interface{}) {
	logrus.Panic(args...)
}

func Panicf(format string, args ...interface{}) {
	logrus.Panicf(format, args...)
}

// WithField 创建并返回一个新的Logger，该Logger在记录日志时会包含额外的字段。
//
// 参数:
//
//	key string: 字段键。
//	val any: 字段值。
//
// 返回值:
//
//	Logger: 包含指定字段的Logger。
func WithField(key string, val any) Logger {
	return &logrusLogger{
		entry: logrus.WithField(key, val),
	}
}

// WithType 创建并返回一个新的Logger，该Logger在记录日志时会包含类型的字段。
//
// 参数:
//
//	key string: 字段键。
//
// 返回值:
//
//	Logger: 包含指定类型字段的Logger。
func WithType[T any](key string) Logger {
	return &logrusLogger{
		entry: logrus.WithField(key, myreflect.TypeOf[T]().Name()),
	}
}
