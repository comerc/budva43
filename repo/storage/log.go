package storage

import (
	"fmt"
	"log/slog"
	"path/filepath"

	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/comerc/budva43/app/config"
)

type Logger struct {
	log *slog.Logger
}

func NewLogger() *Logger {

	writer := &lumberjack.Logger{
		Filename:   filepath.Join(config.Storage.LogDirectory, "badger.log"),
		MaxSize:    config.Storage.LogMaxFileSize,
		MaxBackups: 10,
		MaxAge:     2, // days
		Compress:   false,
	}
	logHandler := slog.NewTextHandler(writer, &slog.HandlerOptions{
		Level: config.Storage.LogLevel,
	})
	logger := slog.New(logHandler)

	return &Logger{
		log: logger,
	}
}

func (l *Logger) Errorf(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	l.log.Error(msg)
}

func (l *Logger) Warningf(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	l.log.Warn(msg)
}

func (l *Logger) Infof(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	l.log.Info(msg)
}

func (l *Logger) Debugf(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	l.log.Debug(msg)
}
