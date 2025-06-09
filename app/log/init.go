package log

import (
	"log/slog"
	"path/filepath"
	"sync"

	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/comerc/budva43/app/config"
)

var once sync.Once

// init() - это зло https://habr.com/ru/articles/771858/
func Init() {
	once.Do(func() {
		setupDefaultLogger()
	})
}

func setupDefaultLogger() {
	writer := &lumberjack.Logger{
		Filename:   filepath.Join(config.LogOptions.Directory, "app.log"),
		MaxSize:    config.LogOptions.MaxFileSize,
		MaxBackups: 10,
		MaxAge:     2, // days
		Compress:   false,
	}
	logHandler := slog.NewTextHandler(writer, &slog.HandlerOptions{
		Level: config.LogOptions.Level,
	})
	logger := slog.New(logHandler)
	slog.SetDefault(logger)
}
