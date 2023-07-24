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

// 输出日志到标准输出。适用于没有设计好日志输出方案时临时使用。
var Std Logger

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

// Init 初始化全局默认的日志器
func Init(cfg *Config) error {
	logrus.SetFormatter(&nested.Formatter{
		TimestampFormat: "2006-01-02 15:04:05",
		NoColors:        true,
		NoFieldsColors:  true,
	})

	level, ok := loggerLevels[strings.ToUpper(cfg.Level)]
	if !ok {
		return fmt.Errorf("invalid log level: %s", cfg.Level)
	}

	logrus.SetLevel(level)

	output := strings.ToUpper(cfg.Output)

	if output == OUTPUT_FILE {
		logFilePath := filepath.Join(cfg.OutputDirectory, cfg.OutputFileName+".log")

		if err := os.MkdirAll(cfg.OutputDirectory, 0755); err != nil {
			return err
		}

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

func WithField(key string, val any) Logger {
	return &logrusLogger{
		entry: logrus.WithField(key, val),
	}
}

func WithType[T any](key string) Logger {
	return &logrusLogger{
		entry: logrus.WithField(key, myreflect.TypeOf[T]().Name()),
	}
}
