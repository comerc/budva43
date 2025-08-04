# Ограничения defer в Go при panic

## Проблема

При использовании `defer` с функциями логирования в контексте `panic` стек вызовов указывает на место `panic`, а не на место объявления `defer`. Это происходит потому, что Go автоматически создает анонимную функцию для прямых вызовов `defer`, и при `panic` стек перестраивается. Это фундаментальное ограничение языка Go.

### Прямой вызов defer

```go
func (t *Try) fn1() {
    defer t.log.ErrorOrDebug(nil, "fn1") // ❌ Неправильно
    panic("test")
}
```

**Результат**: `source="main.go:23 main.(*Try).fn1"` (место panic)

### Явная анонимная функция

```go
func (t *Try) fn1() {
    defer func() {
        t.log.ErrorOrDebug(nil, "fn1") // ✅ Правильно
    }()
    panic("test")
}
```

**Результат**: `source="main.go:23 main.(*Try).fn1"` (место объявления defer)

## Решение

Для получения правильного стека вызовов в defer-ах при panic **всегда** используйте явные анонимные функции:

```go
// ❌ Неправильно
defer t.log.ErrorOrDebug(nil, "operation completed")

// ✅ Правильно
defer func() {
    t.log.ErrorOrDebug(nil, "operation completed")
}()
```

---

```go
package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/comerc/budva43/app/log"
)

type Try struct {
	log *log.Logger
}

func NewTry() *Try {
	return &Try{
		log: log.NewLogger(),
	}
}

func (t *Try) fn1() {
	defer func() {
		t.log.ErrorOrDebug(nil, "fn1")
	}()
	panic("test")
}

func (t *Try) fn2() {
	defer func() {
		t.log.ErrorOrDebug(nil, "fn2")
	}()
	t.fn1()
}

func (t *Try) executeTask(fn func()) {
	defer func() {
		if r := recover(); r != nil {
			err := fmt.Errorf("%v", r)
			t.log.ErrorOrDebug(err, "")
		}
	}()
	fn()
}

func main() {

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})))

	t := NewTry()
	t.executeTask(t.fn2)
}
```