package rate_limiter

import (
	"context"
	"sync"
	"time"

	"github.com/comerc/budva43/app/log"
)

const waitForForward = 3 * time.Second // чтобы бот успел отреагировать на сообщение

// Service управляет скоростью пересылки сообщений
type Service struct {
	log *log.Logger
	//
	mu            sync.Mutex
	lastForwarded map[int64]time.Time
}

// New создает новый сервис для управления скоростью пересылки сообщений
func New() *Service {
	return &Service{
		log: log.NewLogger("service.rate_limiter"),
		//
		lastForwarded: make(map[int64]time.Time),
	}
}

// WaitForForward добавляет задержку, чтобы бот успел отреагировать на сообщение
func (s *Service) WaitForForward(ctx context.Context, dstChatId int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	diff := time.Since(s.lastForwarded[dstChatId])
	if diff < waitForForward {
		// Освобождаем блокировку на время ожидания
		s.mu.Unlock()

		select {
		case <-ctx.Done():
			return
		case <-time.After(waitForForward - diff):
		}

		// Снова захватываем блокировку для записи
		s.mu.Lock()
	}
	s.lastForwarded[dstChatId] = time.Now()
}
