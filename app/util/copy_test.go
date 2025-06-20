package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCopy(t *testing.T) {
	t.Parallel()

	type object struct {
		I int64
		A []string
		M map[string]string
	}
	obj1 := &object{}
	obj1.I = 1
	obj1.A = []string{"a", "b"}
	obj1.M = map[string]string{"a": "b"}
	obj2 := Copy(obj1)
	assert.Equal(t, obj2.I, int64(1))
	assert.Equal(t, obj2.A, []string{"a", "b"})
	assert.Equal(t, obj2.M, map[string]string{"a": "b"})
	obj2.I = 2
	obj2.A[0] = "c"
	obj2.M["a"] = "c"
	assert.Equal(t, obj1.I, int64(1), "obj1.I не изменился")
	assert.Equal(t, obj1.A, []string{"a", "b"}, "obj1.A не изменился")
	assert.Equal(t, obj1.M, map[string]string{"a": "b"}, "obj1.M не изменился")
}
