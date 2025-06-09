package log

import (
	"log/slog"
	"path/filepath"
	"sync"

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
	writer := NewWriter(
		filepath.Join(config.LogOptions.Directory, "app.log"),
		config.LogOptions.MaxFileSize,
	)
	logHandler := slog.NewTextHandler(writer, &slog.HandlerOptions{
		Level: config.LogOptions.Level,
	})
	slog.SetDefault(slog.New(logHandler))
}
