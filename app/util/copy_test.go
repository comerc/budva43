package util

import (
	"fmt"
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
		I int64
		A []string
		M map[string]string
	}
	obj1 := &object{}
	obj1.I = 1
	obj1.A = []string{"a", "b"}
	obj1.M = map[string]string{"a": "b"}
	obj2, err := DeepCopy(obj1)
	require.NoError(t, err)
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

func TestCopyComparison(t *testing.T) {
	t.Parallel()

	type innerObj struct {
		value int
	}

	type object struct {
		child *innerObj
		value int
	}

	obj1 := &object{
		child: &innerObj{value: 42},
		value: 1,
	}

	// SimpleCopy - shallow copy
	simpleObj := SimpleCopy(obj1)

	// DeepCopy - deep copy (но указатели на другие объекты могут быть nil после JSON)
	deepObj, err := DeepCopy(obj1)
	require.NoError(t, err)

	// Проверяем, что все указатели на основные объекты разные
	assert.NotSame(t, obj1, simpleObj, "obj1 и simpleObj должны быть разными")
	assert.NotSame(t, obj1, deepObj, "obj1 и deepObj должны быть разными")
	assert.NotSame(t, simpleObj, deepObj, "simpleObj и deepObj должны быть разными")

	// При SimpleCopy указатель на child остается тем же
	assert.Same(t, obj1.child, simpleObj.child, "SimpleCopy: child должен указывать на тот же объект")

	// При DeepCopy создается новый объект child
	assert.NotSame(t, obj1.child, deepObj.child, "DeepCopy: child должен указывать на другой объект")
	assert.Equal(t, obj1.child.value, deepObj.child.value, "но значения должны быть одинаковыми")

	// Изменяем значение в child
	obj1.child.value = 100

	// При SimpleCopy изменение влияет на копию
	assert.Equal(t, 100, simpleObj.child.value, "SimpleCopy: изменение obj1.child влияет на simpleObj.child")

	// При DeepCopy изменение НЕ влияет на копию
	assert.Equal(t, 42, deepObj.child.value, "DeepCopy: изменение obj1.child НЕ влияет на deepObj.child")
}

func TestPointerAddresses(t *testing.T) {
	t.Parallel()

	type object struct {
		child *object
	}

	obj1 := &object{}
	obj1.child = &object{}
	obj2 := SimpleCopy(obj1)

	// Демонстрируем разницу адресов
	t.Logf("obj1 адрес: %p", obj1)
	t.Logf("obj2 адрес: %p", obj2)
	t.Logf("obj1.child адрес: %p", obj1.child)
	t.Logf("obj2.child адрес: %p", obj2.child)

	// Основной объект: разные адреса
	assert.NotEqual(t, fmt.Sprintf("%p", obj1), fmt.Sprintf("%p", obj2),
		"Адреса основных объектов должны быть разными")

	// Поле child: одинаковые адреса (shallow copy)
	assert.Equal(t, fmt.Sprintf("%p", obj1.child), fmt.Sprintf("%p", obj2.child),
		"Адреса obj1.child и obj2.child должны быть одинаковыми")
}
