package queue

import (
	"context"
	"log/slog"
)

// TODO: очередь задач должна быть реализована на основе "container/list"
// с паузами между задачами, чтобы tdlibClient успевал обрабатывать другие события
// (см. старую реализацию в _budva32/main0.go)

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
		queue: make(chan func(), 100), // TODO: в старой реализации не было ограничения на размер очереди
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
