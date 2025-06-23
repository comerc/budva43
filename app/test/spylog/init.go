package spylog

import (
	"log/slog"
	"sync"

	_ "github.com/comerc/budva43/app/log" // init() app/log before slog.Default()
)

var once sync.Once

// init() - это зло https://habr.com/ru/articles/771858/
// но подходит для реализации синглтона
func init() {
	once.Do(func() {
		CreateHandler(slog.Default())
	})
}
