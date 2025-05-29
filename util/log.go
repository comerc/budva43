package util

import (
	"context"
	"errors"
	"log/slog"
)

type Logger struct {
	slog.Logger
}

func NewLogger(moduleName string) *Logger {
	return &Logger{
		Logger: *slog.With("module", moduleName),
	}
}

func (l *Logger) logOrError(level slog.Level, message string, errPtr *error, argsPtr *[]any) {
	args := []any{}
	err := *errPtr
	if err != nil {
		level = slog.LevelError
		args = append(args, "err", err)
	}
	if argsPtr != nil {
		args = append(args, (*argsPtr)...)
	}
	var errorWithCall *ErrorWithCall
	if errors.As(err, &errorWithCall) {
		args = append(args, "call", errorWithCall.Call)
	}
	l.Log(context.Background(), level, message, args...)
}

func (l *Logger) DebugOrError(message string, errPtr *error, argsPtr *[]any) {
	l.logOrError(slog.LevelDebug, message, errPtr, argsPtr)
}

func (l *Logger) InfoOrError(message string, errPtr *error, argsPtr *[]any) {
	l.logOrError(slog.LevelInfo, message, errPtr, argsPtr)
}

func (l *Logger) WarnOrError(message string, errPtr *error, argsPtr *[]any) {
	l.logOrError(slog.LevelWarn, message, errPtr, argsPtr)
}
