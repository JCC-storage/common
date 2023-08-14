package logger

import (
	"reflect"

	"github.com/sirupsen/logrus"
)

type logrusLogger struct {
	entry *logrus.Entry
}

func (l *logrusLogger) Debug(args ...interface{}) {
	l.entry.Debug(args...)
}
func (l *logrusLogger) Debugf(format string, args ...interface{}) {
	l.entry.Debugf(format, args...)
}

func (l *logrusLogger) Info(args ...interface{}) {
	l.entry.Info(args...)
}
func (l *logrusLogger) Infof(format string, args ...interface{}) {
	l.entry.Infof(format, args...)
}

func (l *logrusLogger) Warn(args ...interface{}) {
	l.entry.Warn(args...)
}
func (l *logrusLogger) Warnf(format string, args ...interface{}) {
	l.entry.Warnf(format, args...)
}

func (l *logrusLogger) Error(args ...interface{}) {
	l.entry.Error(args...)
}
func (l *logrusLogger) Errorf(format string, args ...interface{}) {
	l.entry.Errorf(format, args...)
}

func (l *logrusLogger) Fatal(args ...interface{}) {
	l.entry.Fatal(args...)
}
func (l *logrusLogger) Fatalf(format string, args ...interface{}) {
	l.entry.Fatalf(format, args...)
}

func (l *logrusLogger) Panic(args ...interface{}) {
	l.entry.Panic(args...)
}
func (l *logrusLogger) Panicf(format string, args ...interface{}) {
	l.entry.Panicf(format, args...)
}

func (l *logrusLogger) WithField(key string, val any) Logger {
	return &logrusLogger{
		entry: l.entry.WithField(key, val),
	}
}

func (l *logrusLogger) WithType(key string, typ reflect.Type) Logger {
	return &logrusLogger{
		entry: l.entry.WithField(key, typ.Name()),
	}
}
