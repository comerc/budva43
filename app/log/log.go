package log

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"

	"github.com/comerc/budva43/app/config"
)

type Logger struct {
	slog.Logger
}

func NewLogger(moduleName string) *Logger {
	return &Logger{
		Logger: *slog.With("module", moduleName),
	}
}

func (l *Logger) logOrError(level slog.Level, message string, errPtr *error, args ...any) {
	var err error
	if errPtr != nil && *errPtr != nil {
		err = *errPtr
		message = err.Error()
		level = slog.LevelError
	}
	if err != nil {
		var errorWithCall *ErrorWithCall
		if !errors.As(err, &errorWithCall) {
			errorWithCall = &ErrorWithCall{
				error: err,
				Stack: GetCallStack(3, 0),
			}
		}
		for i, item := range errorWithCall.Stack {
			args = append(args, fmt.Sprintf("stack[%d]", i), item)
		}
	}
	if strings.Contains(message, " ") {
		message = fmt.Sprintf("\"%s\"", message)
	}
	l.Log(context.Background(), level, message, args...)
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

type ErrorWithCall struct {
	error
	Stack []*CallInfo
}

func AddCall(errPtr *error) {
	if errPtr == nil || *errPtr == nil {
		return
	}
	*errPtr = withCall(*errPtr)
}

func WithCall(err error) error {
	if err == nil {
		return nil
	}
	return withCall(err)
}

func withCall(err error) error {
	var errWithCall *ErrorWithCall
	if errors.As(err, &errWithCall) {
		return err // повторная обёртка не нужна
	}
	return &ErrorWithCall{
		error: err,
		Stack: GetCallStack(3, 0),
	}
}

func setupLogger() {
	logHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level:     config.LogOptions.Level,
		AddSource: config.LogOptions.AddSource,
	})
	logger := slog.New(logHandler)
	slog.SetDefault(logger)
}

var once sync.Once

// init - это зло https://habr.com/ru/articles/771858/
// но подходит для реализации синглтона
func init() {
	once.Do(func() {
		setupLogger()
	})
}
