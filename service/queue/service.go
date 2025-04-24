package queue

import (
	"context"
	"log/slog"
)

// Service предоставляет функциональность асинхронной очереди задач
type Service struct {
	log *slog.Logger
	//
	queue chan func()
}

// New создает новый экземпляр сервиса очереди
func New() *Service {
	return &Service{
		log: slog.With("module", "service.queue"),
		//
		queue: make(chan func(), 100),
	}
}

// Start запускает обработчик очереди
func (s *Service) Start(ctx context.Context) error {
	s.log.Info("Запуск сервиса очереди")

	go s.processQueue(ctx)

	return nil
}

// Close останавливает сервис очереди
func (s *Service) Close() error {
	close(s.queue)
	return nil
}

// Add добавляет задачу в очередь
func (s *Service) Add(task func()) {
	s.queue <- task
}

// processQueue обрабатывает очередь задач
func (s *Service) processQueue(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case fn, ok := <-s.queue:
			if !ok {
				return
			}
			fn()
		}
	}
}
