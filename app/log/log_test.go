package log

import (
	"errors"
	"log/slog"
	"testing"

	"github.com/comerc/spylog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
		s.log.DebugOrError("message", &err, args...)
	}()

	err = NewError("error", "arg0", "val0")
	err = WrapError(err, "arg1", "val1")
}

func TestSomeMethod(t *testing.T) {
	var o *SomeObject
	spylogHandler := spylog.GetModuleLogHandler("module_name", t.Name(), func() {
		o = NewSomeObject() // вызываем функцию-конструктор в обёртке spylogHandler
	})
	o.SomeMethod() // вызываем тестируемый метод

	require.True(t, len(spylogHandler.Records) == 1)
	record0 := spylogHandler.Records[0]

	assert.Equal(t, slog.LevelError, record0.Level)
	assert.Equal(t, "error", record0.Message)
	assert.Equal(t, "val0", spylog.GetAttrValue(record0, "arg0"))
	assert.Equal(t, "val1", spylog.GetAttrValue(record0, "arg1"))
	assert.Equal(t, "val2", spylog.GetAttrValue(record0, "arg2"))
	assert.Equal(t, "app/log/log_test.go:36 log.(*SomeObject).NestedMethod",
		spylog.GetAttrValue(record0, "stack[0]"))
}

func TestUnwrappedError(t *testing.T) {
	type OtherObject struct {
		log *Logger
	}

	var o *OtherObject
	spylogHandler := spylog.GetModuleLogHandler("module_name", t.Name(), func() {
		o = &OtherObject{
			log: NewLogger("module_name"),
		}
	})

	err := errors.New("unwrapped error")
	o.log.InfoOrError("message", &err, "arg", "val")

	require.True(t, len(spylogHandler.Records) == 1)
	record0 := spylogHandler.Records[0]

	assert.Equal(t, slog.LevelError, record0.Level)
	assert.Equal(t, "unwrapped error", record0.Message)
	assert.Equal(t, "val", spylog.GetAttrValue(record0, "arg"))
	assert.Equal(t, "app/log/log_test.go:72 log.TestUnwrappedError",
		spylog.GetAttrValue(record0, "stack[0]"))
}
