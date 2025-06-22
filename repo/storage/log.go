package storage

import (
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/comerc/budva43/app/config"
	"github.com/comerc/budva43/app/log"
)

type Logger struct {
	log *slog.Logger
}

func NewLogger() *Logger {
	writer := log.NewWriter(
		filepath.Join(config.Storage.LogDirectory, "badger.log"),
		config.Storage.LogMaxFileSize,
	)
	logHandler := slog.NewJSONHandler(writer, &slog.HandlerOptions{
		Level: config.Storage.LogLevel,
	})
	return &Logger{
		log: slog.New(logHandler),
	}
}

func (l *Logger) Errorf(format string, v ...any) {
	msg := fmt.Sprintf(format, v...)
	l.log.Error(msg)
}

func (l *Logger) Warningf(format string, v ...any) {
	msg := fmt.Sprintf(format, v...)
	l.log.Warn(msg)
}

func (l *Logger) Infof(format string, v ...any) {
	msg := fmt.Sprintf(format, v...)
	l.log.Info(msg)
}

func (l *Logger) Debugf(format string, v ...any) {
	msg := fmt.Sprintf(format, v...)
	l.log.Debug(msg)
}
