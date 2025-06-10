package spylog

// Это обёртка над библиотекой github.com/comerc/spylog.
//
// Решает вопрос циклической зависимости.
// Инициализирует comerc/spylog.
// Используется для тестирования.

import (
	"log/slog"
	"sync"

	"github.com/comerc/spylog"

	_ "github.com/comerc/budva43/app/log" // init() app/log before slog.Default()
)

var once sync.Once

// init() - это зло https://habr.com/ru/articles/771858/
// но подходит для реализации синглтона
func init() {
	once.Do(func() {
		spylog.Init(slog.Default())
	})
}

func GetModuleLogHandler(moduleName, testName string, init func()) *spylog.ModuleLogHandler {
	return spylog.GetModuleLogHandler(moduleName, testName, init)
}

func GetAttrValue(record *slog.Record, key string) string {
	return spylog.GetAttrValue(record, key)
}
