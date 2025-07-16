package queue

import (
	"context"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/comerc/budva43/app/domain"
	"github.com/comerc/budva43/app/testing/spylog"
)

func Test(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)

	synctest.Run(func() {
		// Создаем репозиторий очереди
		var queueRepo *Repo
		spylogHandler := spylog.GetHandler(t.Name(), func() {
			queueRepo = New() // вызываем функцию-конструктор в обёртке spylogHandler
		})

		err := queueRepo.StartContext(ctx)
		require.NoError(t, err)
		t.Cleanup(func() {
			queueRepo.Close()
		})

		var engineConfig *domain.EngineConfig

		// Счетчик выполненных задач
		executed := 0

		// Добавляем задачи, одна из которых вызывает панику
		var fn func()
		fn = func() {
			executed++
			panic("Alarm!")
		}
		queueRepo.Add(fn)

		engineConfig = newEngineConfig(-123)
		engineConfig1 := engineConfig // копируем, см. WATCH-CONFIG.md
		fn = func() {
			assert.Equal(t, int64(-123), engineConfig1.ForwardRules["rule1"].From,
				"Замыкается engineConfig1")
			executed++
		}
		queueRepo.Add(fn)

		engineConfig = newEngineConfig(-321)
		engineConfig2 := engineConfig // копируем, см. WATCH-CONFIG.md
		fn = func() {
			assert.Equal(t, int64(-321), engineConfig2.ForwardRules["rule1"].From,
				"Замыкается engineConfig2")
			executed++
		}
		queueRepo.Add(fn)

		require.Equal(t, 3, queueRepo.Len(), "В очереди должно быть 3 задачи")

		// Добавляем смещение времени для тиков
		time.Sleep(1 * time.Nanosecond)

		// Ждем выполнения всех задач
		time.Sleep(1 * time.Second)
		assert.Equal(t, 1, executed, "Задача с паникой должна выполниться, но не сломать очередь")
		time.Sleep(1 * time.Second)
		assert.Equal(t, 2, executed, "Вторая задача должна выполниться")
		time.Sleep(1 * time.Second)
		assert.Equal(t, 3, executed, "Третья задача должна выполниться")

		// Проверяем запись в лог
		records := spylogHandler.GetRecords()
		require.Equal(t, len(records), 1)
		record := records[0]
		assert.Equal(t, "Alarm!", record.Message)

		// Завершаем контекст
		cancel()
	})
}

func newEngineConfig(from int64) *domain.EngineConfig {
	return &domain.EngineConfig{
		ForwardRules: map[string]*domain.ForwardRule{
			"rule1": {
				From: from,
			},
		},
	}
}
