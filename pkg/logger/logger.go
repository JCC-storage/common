package logger

import (
	"reflect"

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

type Logger interface {
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})

	Info(args ...interface{})
	Infof(format string, args ...interface{})

	Warn(args ...interface{})
	Warnf(format string, args ...interface{})

	Error(args ...interface{})
	Errorf(format string, args ...interface{})

	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})

	Panic(args ...interface{})
	Panicf(format string, args ...interface{})

	WithField(key string, val any) Logger

	WithType(key string, typ reflect.Type) Logger
}
