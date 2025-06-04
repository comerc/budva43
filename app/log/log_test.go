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
			"arg1", "val1",
			"arg2", "val2",
		}
		s.log.DebugOrError("message", &err, args...)
	}()

	err = NewError("error", "arg0", "val0")
	err = WrapError(err, "arg3", "val3")
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
	assert.Equal(t, "val3", spylog.GetAttrValue(record0, "arg3"))
	assert.Equal(t, "app/log/log_test.go:37 log.(*SomeObject).NestedMethod",
		spylog.GetAttrValue(record0, "stack[0]"))
}

func TestSimpleError(t *testing.T) {
	type OtherObject struct {
		log *Logger
	}

	var o *OtherObject
	spylogHandler := spylog.GetModuleLogHandler("module_name", t.Name(), func() {
		o = &OtherObject{
			log: NewLogger("module_name"),
		}
	})

	err := errors.New("simple error")
	o.log.InfoOrError("message", &err)

	require.True(t, len(spylogHandler.Records) == 1)
	record0 := spylogHandler.Records[0]

	assert.Equal(t, slog.LevelError, record0.Level)
	assert.Equal(t, "simple error", record0.Message)
	assert.Equal(t, "app/log/log_test.go:74 log.TestSimpleError",
		spylog.GetAttrValue(record0, "stack[0]"))
}

func TestWrapError(t *testing.T) {
	tests := []struct {
		name          string
		err           error
		args          []any
		expectedError string
		expectedArgs  []any
	}{
		{
			name:          "nil error",
			err:           nil,
			args:          []any{"arg1", "val1"},
			expectedError: "err is nil",
			expectedArgs:  []any{"arg1", "val1"},
		},
		{
			name:          "not CustomError",
			err:           errors.New("error"),
			args:          []any{"arg1", "val1"},
			expectedError: "error",
			expectedArgs:  []any{"arg1", "val1"},
		},
		{
			name:          "CustomError",
			err:           NewError("error", "arg0", "val0"),
			args:          []any{"arg1", "val1"},
			expectedError: "error",
			expectedArgs:  []any{"arg0", "val0", "arg1", "val1"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err, ok := WrapError(test.err, test.args...).(*CustomError)
			assert.True(t, ok)
			assert.Equal(t, test.expectedError, err.Error())
			assert.Equal(t, test.expectedArgs, err.Args)
			assert.NotNil(t, err.Stack)
		})
	}
}
