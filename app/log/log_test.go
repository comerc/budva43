package log

import (
	"errors"
	"log/slog"
	"os"
	"testing"

	"github.com/comerc/spylog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	// Исключение: используется без app/spylog - циклическая зависимость
	spylog.Init(slog.Default()) // init() app/log before slog.Default()
	os.Exit(m.Run())
}

type SomeObject struct {
	log *Logger
}

func NewSomeObject() *SomeObject {
	return &SomeObject{
		log: NewLogger("module_name"),
	}
}

func (s *SomeObject) SomeMethod() {
	s.NestedMethod()
}

func (s *SomeObject) NestedMethod() {
	var err error
	defer func() {
		args := []any{
			"arg2", "val2",
		}
		s.log.ErrorOrDebug(&err, "message", args...)
	}()

	err = NewError("error", "arg0", "val0")
	err = WrapError(err, "arg1", "val1")
}

func TestSomeMethod(t *testing.T) {
	t.Parallel()

	var o *SomeObject
	spylogHandler := spylog.GetModuleLogHandler("module_name", t.Name(), func() {
		o = NewSomeObject() // вызываем функцию-конструктор в обёртке spylogHandler
	})
	o.SomeMethod() // вызываем тестируемый метод

	records := spylogHandler.GetRecords()
	require.True(t, len(records) == 1)

	assert.Equal(t, slog.LevelError, records[0].Level)
	assert.Equal(t, "error", records[0].Message)
	assert.Equal(t, "val0", spylog.GetAttrValue(records[0], "arg0"))
	assert.Equal(t, "val1", spylog.GetAttrValue(records[0], "arg1"))
	assert.Equal(t, "val2", spylog.GetAttrValue(records[0], "arg2"))
	assert.Equal(t, "app/log/log_test.go:43 log.(*SomeObject).NestedMethod",
		spylog.GetAttrValue(records[0], "source"))
}

type SomeError struct {
	error
}

func TestUnwrappedError(t *testing.T) {
	t.Parallel()

	type OtherObject struct {
		log *Logger
	}

	var o *OtherObject
	spylogHandler := spylog.GetModuleLogHandler("module_name", t.Name(), func() {
		o = &OtherObject{
			log: NewLogger("module_name"),
		}
	})

	var err error
	err = &SomeError{
		error: errors.New("unwrapped error"),
	}
	o.log.InfoOrError("message", &err, "arg", "val")

	records := spylogHandler.GetRecords()
	require.True(t, len(records) == 1)

	assert.Equal(t, slog.LevelError, records[0].Level)
	assert.Equal(t, "unwrapped error", records[0].Message)
	assert.Equal(t, "val", spylog.GetAttrValue(records[0], "arg"))
	assert.Equal(t, "log.SomeError", spylog.GetAttrValue(records[0], "type"))
	assert.Equal(t, "app/log/log_test.go:90 log.TestUnwrappedError",
		spylog.GetAttrValue(records[0], "source"))
}
