package log

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
)

type Logger struct {
	slog.Logger
}

func NewLogger(moduleName string) *Logger {
	return &Logger{
		Logger: *slog.With("module", moduleName),
	}
}

func (l *Logger) DebugOrError(message string, errPtr *error, args ...any) {
	l.logOrError(slog.LevelDebug, message, errPtr, args...)
}

func (l *Logger) InfoOrError(message string, errPtr *error, args ...any) {
	l.logOrError(slog.LevelInfo, message, errPtr, args...)
}

func (l *Logger) WarnOrError(message string, errPtr *error, args ...any) {
	l.logOrError(slog.LevelWarn, message, errPtr, args...)
}

func (l *Logger) logOrError(level slog.Level, message string, errPtr *error, args ...any) {
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
		if stack == nil {
			stack = GetCallStack(3, 1)
		}
		args = append(args, "source", stack[0].String())
		// TODO: вынести в конфиг?
		// group := []any{}
		// for i, item := range stack {
		// 	group = append(group, fmt.Sprintf("%d", i), item)
		// }
		// args = append(args, slog.Group("source", group...))
	}
	l.Logger.Log(context.Background(), level, message, args...)
}
