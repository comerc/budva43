package log

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/comerc/budva43/app/config"
	"github.com/comerc/budva43/app/entity"
	"gopkg.in/natefinch/lumberjack.v2"
)

// из-за циклической зависимости app/spylog vs app/log - тесты вынесены в test/log_test.go

type Logger struct {
	slog.Logger
}

func NewLogger(moduleName string) *Logger {
	return &Logger{
		Logger: *slog.With("module", moduleName),
	}
}

func (l *Logger) ErrorOrDebug(errPtr *error, message string, args ...any) {
	l.logWithError(slog.LevelDebug, errPtr, message, args...)
}

func (l *Logger) ErrorOrInfo(errPtr *error, message string, args ...any) {
	l.logWithError(slog.LevelInfo, errPtr, message, args...)
}

func (l *Logger) ErrorOrWarn(errPtr *error, message string, args ...any) {
	l.logWithError(slog.LevelWarn, errPtr, message, args...)
}

func (l *Logger) logWithError(level slog.Level, errPtr *error, message string, args ...any) {
	var err error
	if errPtr != nil && *errPtr != nil {
		err = *errPtr
		level = slog.LevelError
		message = err.Error()
		var stack []*CallInfo
		var customError *CustomError
		if errors.As(err, &customError) {
			args = append(customError.Args, args...)
			stack = customError.Stack
			err = customError.Unwrap()
		}
		typeName := strings.TrimPrefix(fmt.Sprintf("%T", err), "*")
		args = append(args, "type", typeName)
		if config.ErrorSource.Type != entity.TypeErrorSourceNone {
			if stack == nil {
				stack = GetCallStack(3)
			}
			switch config.ErrorSource.Type {
			case entity.TypeErrorSourceMore:
				var groupArgs []any
				for i, item := range stack {
					groupArgs = append(groupArgs, fmt.Sprintf("%d", i), item.String())
				}
				args = append(args, slog.Group("source", groupArgs...))
			case entity.TypeErrorSourceOne:
				args = append(args, "source", stack[0].String())
			}
		}
	}
	l.Log(context.Background(), level, message, args...)
}

func NewWriter(filePath string, maxSize int) io.Writer {
	testing := os.Getenv("GOEXPERIMENT") == "synctest"
	if testing {
		return io.Discard
	}
	return &lumberjack.Logger{
		Filename:   filePath,
		MaxSize:    maxSize,
		MaxBackups: 10,
		MaxAge:     2, // days
		Compress:   false,
	}
}
