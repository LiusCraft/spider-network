package xlog

import (
	"context"
)

func SetGlobalLogId(id string) {
	WithLogId(defaultLogger, id)
}

func SetGlobalCtx(ctx context.Context) {
	WithCtx(defaultLogger, ctx)
}

// Debugf logs a message at the debug level.
func Debugf(format string, v ...interface{}) {
	defaultLogger.Debugf(format, v...)
}

// Infof logs a message at the info level.
func Infof(format string, v ...interface{}) {
	defaultLogger.Infof(format, v...)
}

// Warnf logs a message at the warn level.
func Warnf(format string, v ...interface{}) {
	defaultLogger.Warnf(format, v...)
}

// Errorf logs a message at the error level.
func Errorf(format string, v ...interface{}) {
	defaultLogger.Errorf(format, v...)
}

// Fatalf logs a message at the fatal level and then exits.
func Fatalf(format string, v ...interface{}) {
	defaultLogger.Fatalf(format, v...)
}
func Debug(v ...interface{}) {
	defaultLogger.Debug(v...)
}
func Info(v ...interface{}) {
	defaultLogger.Info(v...)
}
func Warn(v ...interface{}) {
	defaultLogger.Warn(v...)
}
func Error(v ...interface{}) {
	defaultLogger.Error(v...)
}
func Fatal(v ...interface{}) {
	defaultLogger.Fatal(v...)
}
