package queue

import (
	"container/list"
	"context"
	"log/slog"
	"time"
)

// Service предоставляет функциональность асинхронной очереди задач
type Service struct {
	log *slog.Logger
	//
	queue *list.List
}

// New создает новый экземпляр сервиса очереди
func New() *Service {
	return &Service{
		log: slog.With("module", "service.queue"),
		//
		queue: list.New(),
	}
}

// Start запускает обработчик очереди
func (s *Service) Start(ctx context.Context) error {
	s.log.Info("Запуск сервиса очереди")

	go s.runQueue(ctx)

	return nil
}

// Close останавливает сервис очереди
func (s *Service) Close() error {
	return nil
}

// Add добавляет задачу в очередь
func (s *Service) Add(fn func()) {
	s.queue.PushBack(fn)
}

// runQueue обрабатывает очередь задач
func (s *Service) runQueue(ctx context.Context) {
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
				// This will remove the allocated memory and avoid memory leaks
				s.queue.Remove(front)
			}
		}
	}
}
