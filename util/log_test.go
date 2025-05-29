package util

import (
	"errors"
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
	var (
		err  error
		args []any
	)
	defer s.log.InfoOrError("message", &err, &args)

	err = WithCall(errors.New("error"))
	args = append(args, "arg1", "val1", "arg2", "val2")
}

func TestSomeMethod(t *testing.T) {
	var o *SomeObject
	spylogHandler := spylog.GetModuleLogHandler("module_name", t.Name(), func() {
		o = NewSomeObject() // вызываем функцию-конструктор в обёртке logHandler
	})
	o.SomeMethod() // вызываем тестируемый метод

	require.True(t, len(spylogHandler.Records) == 1)
	record0 := spylogHandler.Records[0]

	assert.Equal(t, "message", record0.Message)
	assert.Equal(t, "error", spylog.GetAttrValue(record0, "err"))
	assert.Equal(t, "val1", spylog.GetAttrValue(record0, "arg1"))
	assert.Equal(t, "val2", spylog.GetAttrValue(record0, "arg2"))
	assert.Equal(t, "util/log_test.go:33 util.(*SomeObject).NestedMethod",
		spylog.GetAttrValue(record0, "call"), "WithCall() не работает")
}
