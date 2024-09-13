package xlog

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/liuscraft/spider-network/pkg/utils"
)

var (
	defaultLog = log.New(os.Stdout, "", log.LstdFlags)
)

// xlog is a simple logger.
// It is used to log messages at different levels (debug, info, warn, error, fatal).
// It is based on the standard log package.

// [!]it is only implemented in order to be able to use, and it will be improved later

type Logger interface {
	// Debugf logs a message at the debug level.
	Debugf(format string, v ...interface{})
	// Infof logs a message at the info level.
	Infof(format string, v ...interface{})
	// Warnf logs a message at the warn level.
	Warnf(format string, v ...interface{})
	// Errorf logs a message at the error level.
	Errorf(format string, v ...interface{})
	// Fatalf logs a message at the fatal level and then exits.
	Fatalf(format string, v ...interface{})
	Debug(v ...interface{})
	Info(v ...interface{})
	Warn(v ...interface{})
	Error(v ...interface{})
	Fatal(v ...interface{})
	SetOutput(v ...io.Writer)
}

func (l *logger) Deadline() (deadline time.Time, ok bool) {
	return l.ctx.Deadline()
}

func (l *logger) Done() <-chan struct{} {
	return l.ctx.Done()
}

func (l *logger) Err() error {
	return l.ctx.Err()
}

func (l *logger) Value(key any) any {
	return l.ctx.Value(key)
}

func genLogId() string {
	return utils.RandString(15)
}

type logger struct {
	std   *log.Logger
	ctx   context.Context
	logId string
}

func tiggerLogger(xl Logger) *logger {
	l, ok := xl.(*logger)
	if !ok {
		return NewLogger().(*logger)
	}
	return l
}

func NewLogger() Logger {
	l := &logger{
		std: defaultLog,
		ctx: context.TODO(),
	}
	return WithLogId(l, "")
}

func WithLogId(xl Logger, logId string) Logger {
	if logId == "" {
		logId = genLogId()
	}
	logId = fmt.Sprintf("[%s]", logId)
	l := tiggerLogger(xl)
	if logId != "" {
		l.logId = logId
	}
	return l
}

func WithCtx(xl Logger, ctx context.Context) Logger {
	if ctx == nil {
		ctx = context.Background()
	}
	l := tiggerLogger(xl)
	l.ctx = ctx
	return l
}

func (l *logger) SetOutput(ws ...io.Writer) {
	ws = append([]io.Writer{os.Stdout}, ws...)
	l.std.SetOutput(io.MultiWriter(ws...))
}

func (l *logger) send(level Level, v ...interface{}) {
	switch level {
	case LevelFatal:
		l.std.Fatalf("%s [%s] %s", l.logId, level.String(), fmt.Sprint(v...))
	default:
		l.std.Printf("%s [%s] %s", l.logId, level.String(), fmt.Sprint(v...))
	}
}

// Debugf logs a message at the debug level.
func (l *logger) Debugf(format string, v ...interface{}) {
	l.send(LevelDebug, fmt.Sprintf(format, v...))
}

// Infof logs a message at the info level.
func (l *logger) Infof(format string, v ...interface{}) {
	l.send(LevelInfo, fmt.Sprintf(format, v...))
}

// Warnf logs a message at the warn level.
func (l *logger) Warnf(format string, v ...interface{}) {
	l.send(LevelWarn, fmt.Sprintf(format, v...))
}

// Errorf logs a message at the error level.
func (l *logger) Errorf(format string, v ...interface{}) {
	l.send(LevelError, fmt.Sprintf(format, v...))
}

// Fatalf logs a message at the fatal level and then exits.
func (l *logger) Fatalf(format string, v ...interface{}) {
	l.send(LevelFatal, fmt.Sprintf(format, v...))
}

func (l *logger) Debug(v ...interface{}) {
	l.send(LevelDebug, v...)
}

func (l *logger) Info(v ...interface{}) {
	l.send(LevelInfo, v...)
}

func (l *logger) Warn(v ...interface{}) {
	l.send(LevelWarn, v...)
}

func (l *logger) Error(v ...interface{}) {
	l.send(LevelError, v...)
}

func (l *logger) Fatal(v ...interface{}) {
	l.send(LevelFatal, v...)
}
