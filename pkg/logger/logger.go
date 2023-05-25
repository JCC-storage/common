package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
)

const (
	TRACE_LEVEL = "TRACE"
	DEBUG_LEVEL = "DEBUG"
	INFO_LEVEL  = "INFO"
	WARN_LEVEL  = "WARN"
	ERROR_LEVEL = "ERROR"
	FATAL_LEVEL = "FATAL"
	PANIC_LEVEL = "PANIC"

	OUTPUT_FILE   = "FILE"
	OUTPUT_STDOUT = "STDOUT"
)

var loggerLevels = map[string]logrus.Level{
	TRACE_LEVEL: logrus.TraceLevel,
	DEBUG_LEVEL: logrus.DebugLevel,
	INFO_LEVEL:  logrus.InfoLevel,
	WARN_LEVEL:  logrus.WarnLevel,
	ERROR_LEVEL: logrus.ErrorLevel,
	FATAL_LEVEL: logrus.FatalLevel,
	PANIC_LEVEL: logrus.PanicLevel,
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

		if err := os.MkdirAll(cfg.OutputDirectory, 0644); err != nil {
			return err
		}

		file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE, 0644)
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

func WithField(key string, val any) *logrus.Entry {
	return logrus.WithField(key, val)
}
