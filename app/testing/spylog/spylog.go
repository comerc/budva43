package spylog

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"reflect"
	"runtime"
	"strconv"
	"sync"
)

// из-за циклической зависимости app/spylog vs app/log - тесты вынесены в test/log_test.go

// ready for t.Parallel() and multiple t.Run()

var handlerInstance *logHandler

func CreateHandler(logger *slog.Logger) {
	handlerInstance = &logHandler{
		handlers: make(map[string]map[string]*ModuleLogHandler),
		handler:  logger.Handler(),
	}
}

type logHandler struct {
	mu       sync.Mutex
	current  sync.Map
	handlers map[string]map[string]*ModuleLogHandler
	handler  slog.Handler
}

func GetModuleLogHandler(moduleName, testName string, init func()) *ModuleLogHandler {
	h := handlerInstance
	h.mu.Lock()
	defer h.mu.Unlock()
	h.current.Store(getGID(), testName) // need for WithAttrs
	handlers, ok := h.handlers[moduleName]
	if !ok {
		handlers = make(map[string]*ModuleLogHandler)
		h.handlers[moduleName] = handlers
	}
	handler, ok := handlers[testName]
	if !ok {
		handler = &ModuleLogHandler{}
		h.handlers[moduleName][testName] = handler
	}
	slog.SetDefault(slog.New(h))
	init() // for slog.With("module", "name")
	return handler
}

func (h *logHandler) Handle(ctx context.Context, r slog.Record) error {
	h.handler.Handle(ctx, r)
	return nil
}

func (h *logHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	var module string
	for _, attr := range attrs {
		if attr.Key == "module" {
			module = attr.Value.String()
			break
		}
	}

	if module == "" {
		return h
	}

	if testName, ok := h.current.Load(getGID()); ok {
		if handler, exists := h.handlers[module][testName.(string)]; exists {
			return handler
		}
	}
	return h
}

func getGID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}

func (h *logHandler) WithGroup(name string) slog.Handler {
	return h
}

func (h *logHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

type ModuleLogHandler struct {
	records []*slog.Record
}

func (h *ModuleLogHandler) GetRecords() []*slog.Record {
	return h.records
}

func (h *ModuleLogHandler) Handle(ctx context.Context, r slog.Record) error {
	handlerInstance.handler.Handle(ctx, r)
	h.records = append(h.records, &r)
	return nil
}

func (h *ModuleLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *ModuleLogHandler) WithGroup(name string) slog.Handler {
	return h
}

func (h *ModuleLogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return handlerInstance.handler.Enabled(ctx, level)
}

func GetAttrValue(record *slog.Record, key string) string {
	var value *slog.Value
	record.Attrs(func(attr slog.Attr) bool {
		if attr.Key == key {
			value = &attr.Value
			return false
		}
		return true
	})

	if value == nil {
		return ""
	}

	// Получаем любое значение из slog.Value
	anyValue := value.Any()
	if anyValue != nil {
		// Проверяем, является ли значение указателем
		rv := reflect.ValueOf(anyValue)
		if rv.Kind() == reflect.Ptr && !rv.IsNil() {
			// Разыменовываем указатель и форматируем значение
			return fmt.Sprintf("%v", rv.Elem().Interface())
		}
		// Если не указатель, используем fmt.Sprintf("%v")
		return fmt.Sprintf("%v", anyValue)
	}

	// Если Any() вернул nil, используем стандартное строковое представление
	return value.String()
}
