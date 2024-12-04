package xlog

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/liuscraft/spider-network/pkg/utils"
)

var (
	defaultLog    = log.New(os.Stdout, "", log.LstdFlags)
	defaultLogger = New()
	projectRoot   = findProjectRoot()
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

func genLogId() string {
	return utils.RandString(15)
}

// findProjectRoot 获取项目根目录
func findProjectRoot() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return ""
	}
	// 获取 pkg/xlog/xlog.go 的父目录的父目录
	return filepath.Dir(filepath.Dir(filepath.Dir(file)))
}

// getRelativePath 获取源文件的相对路径
func getRelativePath(file string) string {
	if projectRoot == "" {
		return file
	}
	rel, err := filepath.Rel(projectRoot, file)
	if err != nil {
		return file
	}
	return rel
}

type logger struct {
	std   *log.Logger
	ctx   context.Context
	logId string
}

func convertLogger(xl Logger) *logger {
	l, ok := xl.(*logger)
	if !ok {
		return New().(*logger)
	}
	return l
}

func New() Logger {
	l := &logger{
		std: defaultLog,
		ctx: context.TODO(),
	}
	return WithLogId(l, "")
}

func NewWithLogId(logId string) Logger {
	return WithLogId(New(), logId)
}

func WithLogId(xl Logger, logId string) Logger {
	if logId == "" {
		logId = genLogId()
	}
	logId = fmt.Sprintf("[%s]", logId)
	l := convertLogger(xl)
	if logId != "" {
		l.logId = logId
	}
	return l
}

func (l *logger) SetOutput(ws ...io.Writer) {
	ws = append([]io.Writer{os.Stdout}, ws...)
	l.std.SetOutput(io.MultiWriter(ws...))
}

func (l *logger) send(level Level, v ...interface{}) {
	// 获取调用者信息
	_, file, line, ok := runtime.Caller(2)
	var location string
	if ok {
		location = fmt.Sprintf("%s:%d", getRelativePath(file), line)
	} else {
		location = "unknown:0"
	}

	switch level {
	case LevelFatal:
		l.std.Fatalf("%s [%s] {%s} %s", l.logId, level.String(), location, fmt.Sprint(v...))
	default:
		l.std.Printf("%s [%s] {%s} %s", l.logId, level.String(), location, fmt.Sprint(v...))
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
