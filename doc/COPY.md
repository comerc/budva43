# Copy()

## Ограничения:

### 1. **Ограничения сериализации msgpack:**
- **Интерфейсы:** Не может корректно обрабатывать поля типа `interface{}` с произвольными типами
- **Каналы и функции:** Поля типа `chan` и `func` игнорируются 
- **Неэкспортируемые поля:** Приватные поля структур не сериализуются
- **Циклические ссылки:** Может привести к бесконечной рекурсии или панике

### 2. **Ограничения типов:**
- **Карты с не-строковыми ключами:** Ключи карт должны быть строками или конвертируемыми в строки
- **Пользовательские типы:** Типы без поддержки msgpack сериализации вызовут панику

### 3. **Производительность:**
- Медленная операция из-за маршалинга/демаршалинга через msgpack
- Выделение памяти для промежуточного буфера байтов
- Использование рефлексии внутри msgpack

### 4. **Обработка ошибок:**
- При любой ошибке сериализации происходит `log.Panic` (жёсткое завершение)

## Преимущества:

### 1. **Глубокое копирование:**
```go
func Copy[T any](from *T) *T {
	var err error
	var b []byte
	b, err = msgpack.Marshal(from)
	if err != nil {
		log.Panic("Copy: ", err)
	}
	to := new(T)
	err = msgpack.Unmarshal(b, to)
	if err != nil {
		log.Panic("Copy: ", err)
	}
	return to
}
```

`Copy` создаёт **полную независимую копию** всех вложенных структур, срезов, карт и указателей.

### 2. **Изоляция данных:**
Как показывает тест в `primitive_test.go`, после использования `Copy` изменения в копии не влияют на оригинал:

```go
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
```

## Рекомендации по использованию:

- **Используйте `Copy`** когда нужно полное клонирование сложных структур с вложенными объектами, но убедитесь что все типы поддерживают msgpack сериализацию