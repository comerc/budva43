package log

import (
	"log/slog"
	"path/filepath"
	"sync"

	"github.com/comerc/budva43/app/config" // init() config
)

var once sync.Once

// init() - это зло https://habr.com/ru/articles/771858/
// но подходит для реализации синглтона
func init() {
	once.Do(func() {
		setupDefaultLogger()
		createLoggerNames()
	})
}

func setupDefaultLogger() {
	writer := NewWriter(
		filepath.Join(config.General.Log.Directory, "app.log"),
		config.General.Log.MaxFileSize,
	)
	logHandler := slog.NewJSONHandler(writer, &slog.HandlerOptions{
		Level: config.General.Log.Level,
	})
	slog.SetDefault(slog.New(logHandler))
}
