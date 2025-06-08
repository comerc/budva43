package queue

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"testing/synctest"
	"time"

	"github.com/comerc/spylog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/comerc/budva43/app/config"
	"github.com/comerc/budva43/app/log"
)

func TestMain(m *testing.M) {
	config.Init()
	log.Init()
	spylog.Init(slog.Default())
	os.Exit(m.Run())
}

func TestQueueRepo(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)

	synctest.Run(func() {
		// Создаем репозиторий очереди
		var queueRepo *Repo
		spylogHandler := spylog.GetModuleLogHandler("repo.queue", t.Name(), func() {
			queueRepo = New() // вызываем функцию-конструктор в обёртке spylogHandler
		})

		err := queueRepo.Start(ctx)
		require.NoError(t, err)
		t.Cleanup(func() {
			queueRepo.Close()
		})

		// Счетчик выполненных задач
		executed := 0

		// Добавляем задачи, одна из которых вызывает панику
		queueRepo.Add(func() { executed++ })
		queueRepo.Add(func() {
			executed++
			panic("Alarm!")
		})
		queueRepo.Add(func() { executed++ })
		require.Equal(t, 3, queueRepo.Len(), "В очереди должно быть 3 задачи")

		// Добавляем смещение времени для тиков
		time.Sleep(1 * time.Nanosecond)

		// Ждем выполнения всех задач
		time.Sleep(1 * time.Second)
		assert.Equal(t, 1, executed, "Первая задача должна выполниться")
		time.Sleep(1 * time.Second)
		assert.Equal(t, 2, executed, "Задача с паникой должна выполниться, но не сломать очередь")
		time.Sleep(1 * time.Second)
		assert.Equal(t, 3, executed, "Третья задача должна выполниться после паники")

		// Проверяем запись в лог
		require.True(t, len(spylogHandler.Records) == 1)
		record0 := spylogHandler.Records[0]
		assert.Equal(t, "Alarm!", record0.Message)

		// Завершаем контекст
		cancel()
	})
}
