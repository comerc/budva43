package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSimpleCopy(t *testing.T) {
	t.Parallel()

	type object struct {
		child *object
	}
	obj1 := &object{}
	obj1.child = &object{}
	obj2 := SimpleCopy(obj1)

	// Проверяем, что указатели на основные объекты разные (разные адреса)
	require.NotSame(t, obj1, obj2, "obj1 и obj2 должны быть разными указателями")

	// Проверяем, что указатель на поле child остался тем же (shallow copy)
	assert.Same(t, obj1.child, obj2.child, "obj1.child и obj2.child должны указывать на один объект")
}

func TestDeepCopy(t *testing.T) {
	t.Parallel()

	type object struct {
		I int
		A []string
		M map[string]string
		O *object
		P *object
	}
	obj1 := &object{}
	obj1.I = 1
	obj1.A = []string{"a", "b"}
	obj1.M = map[string]string{"a": "b"}
	obj1.O = &object{I: 1}
	obj1.P = nil
	obj2, err := DeepCopy(obj1)
	require.NoError(t, err)
	assert.Equal(t, obj2.I, 1)
	assert.Equal(t, obj2.A, []string{"a", "b"})
	assert.Equal(t, obj2.M, map[string]string{"a": "b"})
	assert.NotSame(t, obj1.O, obj2.O, "obj1.O и obj2.O должны быть разными указателями")
	assert.Same(t, obj1.P, obj2.P, "obj1.P и obj2.P должны быть одинаковыми указателями на nil")
	obj2.I = 2
	obj2.A[0] = "c"
	obj2.M["a"] = "c"
	obj2.O.I = 2
	obj2.P = &object{}
	assert.Equal(t, obj1.I, 1, "obj1.I не изменился")
	assert.Equal(t, obj1.A, []string{"a", "b"}, "obj1.A не изменился")
	assert.Equal(t, obj1.M, map[string]string{"a": "b"}, "obj1.M не изменился")
	assert.Equal(t, obj1.O.I, 1, "obj1.O не изменился")
	assert.Nil(t, obj1.P, "obj1.P не изменился")
}
