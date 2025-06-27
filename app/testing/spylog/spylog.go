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

var root *rootHandler

func createRootHandler(logger *slog.Logger) {
	root = &rootHandler{
		handlers: make(map[string]map[string]*Handler),
		handler:  logger.Handler(),
	}
}

type rootHandler struct {
	mu       sync.Mutex
	current  sync.Map
	handlers map[string]map[string]*Handler
	handler  slog.Handler
}

func getHandler(loggerName, testName string, init func()) *Handler {
	root.mu.Lock()
	defer root.mu.Unlock()
	root.current.Store(getGID(), testName) // need for WithAttrs
	handlers, ok := root.handlers[loggerName]
	if !ok {
		handlers = make(map[string]*Handler)
		root.handlers[loggerName] = handlers
	}
	handler, ok := handlers[testName]
	if !ok {
		handler = &Handler{}
		root.handlers[loggerName][testName] = handler
	}
	slog.SetDefault(slog.New(root))
	init() // for slog.With("logger", "name")
	return handler
}

func (h *rootHandler) Handle(ctx context.Context, r slog.Record) error {
	h.handler.Handle(ctx, r)
	return nil
}

func (h *rootHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	var module string
	for _, attr := range attrs {
		if attr.Key == "logger" {
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

func (h *rootHandler) WithGroup(name string) slog.Handler {
	return h
}

func (h *rootHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

type Handler struct {
	records []*slog.Record
}

func (h *Handler) GetRecords() []*slog.Record {
	return h.records
}

func (h *Handler) Handle(ctx context.Context, r slog.Record) error {
	root.handler.Handle(ctx, r)
	h.records = append(h.records, &r)
	return nil
}

func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *Handler) WithGroup(name string) slog.Handler {
	return h
}

func (h *Handler) Enabled(ctx context.Context, level slog.Level) bool {
	return root.handler.Enabled(ctx, level)
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
