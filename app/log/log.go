package log

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/comerc/budva43/app/config"
)

type Logger struct {
	log *slog.Logger
}

func NewLogger(moduleName string) *Logger {
	return &Logger{
		log: slog.With("module", moduleName),
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
		var stack []*CallInfo
		var customError *CustomError
		if errors.As(err, &customError) {
			args = append(args, customError.Args...)
			stack = customError.Stack
		}
		if stack == nil {
			stack = GetCallStack(3, 0)
		}
		for i, item := range stack {
			args = append(args, fmt.Sprintf("stack[%d]", i), item)
		}
	}
	l.log.Log(context.Background(), level, message, args...)
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

func WrapError(err error, args ...any) error {
	if err == nil {
		err = errors.New("err is nil")
	} else {
		var customError *CustomError
		if errors.As(err, &customError) {
			customError.Args = append(customError.Args, args...)
			return customError
		}
	}
	return &CustomError{
		error: err,
		Args:  args,
		Stack: GetCallStack(2, 0),
	}
}

func NewError(text string, args ...any) error {
	return &CustomError{
		error: errors.New(text),
		Args:  args,
		Stack: GetCallStack(2, 0),
	}
}

type CustomError struct {
	error
	Args  []any
	Stack []*CallInfo
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
