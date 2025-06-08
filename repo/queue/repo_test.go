package queue

import (
	"context"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueueRepo(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	synctest.Run(func() {
		// Создаем репозиторий очереди
		queue := New()
		err := queue.Start(ctx)
		require.NoError(t, err)
		t.Cleanup(func() {
			queue.Close()
		})

		// Проверяем, что очередь изначально пуста
		assert.Equal(t, 0, queue.Len(), "Длина очереди должна быть 0")

		// Счетчик выполненных задач
		executed := 0

		// Добавляем несколько задач в очередь
		queue.Add(func() { executed++ })
		queue.Add(func() { executed++ })
		queue.Add(func() { executed++ })

		// Проверяем, что задачи добавились
		assert.Equal(t, 3, queue.Len(), "В очереди должно быть 3 задачи")

		// Добавляю смещение времени для тиков
		time.Sleep(1 * time.Millisecond)

		// В synctest время контролируется, поэтому мы можем точно знать,
		// когда задачи будут выполнены

		// Ждем первый тик таймера (1 секунда)
		time.Sleep(1 * time.Second)
		assert.Equal(t, 1, executed, "Должна быть выполнена одна задача после первого тика")
		assert.Equal(t, 2, queue.Len(), "В очереди должно остаться 2 задачи")

		// Ждем второй тик
		time.Sleep(1 * time.Second)
		assert.Equal(t, 2, executed, "Должны быть выполнены две задачи после второго тика")
		assert.Equal(t, 1, queue.Len(), "В очереди должна остаться 1 задача")

		// Ждем третий тик
		time.Sleep(1 * time.Second)
		assert.Equal(t, 3, executed, "Должны быть выполнены все три задачи после третьего тика")
		assert.Equal(t, 0, queue.Len(), "Очередь должна быть пуста после выполнения всех задач")

		// Проверяем, что больше задач нет
		time.Sleep(1 * time.Second)
		assert.Equal(t, 3, executed, "Количество выполненных задач не должно измениться")
		assert.Equal(t, 0, queue.Len(), "Очередь должна оставаться пустой")
	})
}
