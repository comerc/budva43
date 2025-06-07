package log

import (
	"os"
	"sync"

	"log/slog"

	"github.com/comerc/budva43/app/config"
)

var once sync.Once

// init() - это зло https://habr.com/ru/articles/771858/
func Init() {
	once.Do(func() {
		setupLogger()
	})
}

func setupLogger() {
	logHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: config.LogOptions.Level,
	})
	logger := slog.New(logHandler)
	slog.SetDefault(logger)
}
