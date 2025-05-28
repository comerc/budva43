package rate_limiter

import (
	"context"
	"sync"
	"time"

	"github.com/comerc/budva43/util"
)

// TODO: зочем? у нас уже есть таймер в service/queue/service.go, который отвечает за задержку между сообщениями

const waitForForward = 3 * time.Second // чтобы бот успел отреагировать на сообщение

// Service управляет скоростью пересылки сообщений
type Service struct {
	log *util.Logger
	//
	mu            sync.Mutex
	lastForwarded map[int64]time.Time
}

// New создает новый сервис для управления скоростью пересылки сообщений
func New() *Service {
	return &Service{
		log: util.NewLogger("service.rate_limiter"),
		//
		lastForwarded: make(map[int64]time.Time),
	}
}

// getLastForwardedDiff возвращает время, прошедшее с момента последней пересылки сообщений в целевой чат
func (s *Service) getLastForwardedDiff(dstChatId int64) time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()
	return time.Since(s.lastForwarded[dstChatId])
}

// setLastForwarded устанавливает время последней пересылки сообщений в целевой чат
func (s *Service) setLastForwarded(dstChatId int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastForwarded[dstChatId] = time.Now()
}

// WaitForForward добавляет задержку, чтобы бот успел отреагировать на сообщение
func (s *Service) WaitForForward(ctx context.Context, dstChatId int64) {
	diff := s.getLastForwardedDiff(dstChatId)
	if diff < waitForForward {
		select {
		case <-ctx.Done():
			return
		case <-time.After(waitForForward - diff):
		}
	}
	s.setLastForwarded(dstChatId)
}
