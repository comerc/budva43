package queue

import (
	"container/list"
	"context"
	"log/slog"
	"time"
)

// Repo предоставляет функциональность асинхронной очереди задач
type Repo struct {
	log *slog.Logger
	//
	queue *list.List
}

// New создает новый экземпляр сервиса очереди
func New() *Repo {
	return &Repo{
		log: slog.With("module", "repo.queue"),
		//
		queue: list.New(),
	}
}

// Start запускает обработчик очереди
func (s *Repo) Start(ctx context.Context) error {
	go s.run(ctx)

	return nil
}

// Close останавливает сервис очереди
func (s *Repo) Close() error {
	return nil
}

// Add добавляет задачу в очередь
func (s *Repo) Add(fn func()) {
	s.queue.PushBack(fn)
}

// run обрабатывает очередь задач
func (s *Repo) run(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			front := s.queue.Front()
			if front != nil {
				fn := front.Value.(func())
				fn()
				// Это позволит удалить выделенную память и избежать утечек памяти
				s.queue.Remove(front)
			}
		}
	}
}
