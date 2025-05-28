package util

import (
	"fmt"
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

	err = fmt.Errorf("error")
	args = append(args, "arg1", "val1", "arg2", "val2")
}

func TestSomeMethod(t *testing.T) {
	var o *SomeObject
	logHandler := spylog.GetModuleLogHandler("module_name", t.Name(), func() {
		o = NewSomeObject() // вызываем функцию-конструктор в обёртке logHandler
	})
	o.SomeMethod() // вызываем тестируемый метод

	require.True(t, len(logHandler.Records) == 1)
	record0 := logHandler.Records[0]

	assert.Equal(t, "message", record0.Message)
	assert.Equal(t, "error", spylog.GetAttrValue(record0, "err"))
	assert.Equal(t, "val1", spylog.GetAttrValue(record0, "arg1"))
	assert.Equal(t, "val2", spylog.GetAttrValue(record0, "arg2"))
	stack := []string{
		"util/log_test.go:35 util.(*SomeObject).NestedMethod",
		"util/log_test.go:23 util.(*SomeObject).SomeMethod",
		"util/log_test.go:42 util.TestSomeMethod",
	}
	for i, call := range stack {
		assert.Equal(t, call, spylog.GetAttrValue(record0, fmt.Sprintf("stack[%d]", i)))
	}
}
