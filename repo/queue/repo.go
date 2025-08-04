package queue

import (
	"container/list"
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/comerc/budva43/app/log"
)

// Repo предоставляет функциональность асинхронной очереди задач
type Repo struct {
	log *log.Logger
	//
	queue *list.List
	mu    sync.RWMutex
}

// New создает новый экземпляр сервиса очереди
func New() *Repo {
	return &Repo{
		log: log.NewLogger(),
		//
		queue: list.New(),
	}
}

// StartContext запускает обработчик очереди
func (s *Repo) StartContext(ctx context.Context) error {

	go s.run(ctx)

	return nil
}

// Close останавливает сервис очереди
func (s *Repo) Close() error {
	return nil
}

// Add добавляет задачу в очередь
func (s *Repo) Add(fn func()) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.queue.PushBack(fn)
}

// Len возвращает количество задач в очереди
func (s *Repo) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.queue.Len()
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
			s.processQueue()
		}
	}
}

// processQueue обрабатывает одну задачу из очереди
func (s *Repo) processQueue() {
	s.mu.Lock()
	defer s.mu.Unlock()

	front := s.queue.Front()
	if front != nil {
		fn := front.Value.(func())
		// Это позволит удалить выделенную память и избежать утечек памяти
		s.queue.Remove(front)
		s.executeTask(fn)
	}
}

// executeTask безопасно выполняет задачу с recovery
func (s *Repo) executeTask(fn func()) {
	defer func() {
		if r := recover(); r != nil {
			err := fmt.Errorf("%v", r)
			s.log.ErrorOrDebug(err, "")
		}
	}()
	fn()
}
