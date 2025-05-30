package forwarded_to

import (
	"sync"

	"github.com/comerc/budva43/app/log"
)

type Service struct {
	log *log.Logger
	//
	mu sync.Mutex
}

func New() *Service {
	return &Service{
		log: log.NewLogger("service.forwarded_to"),
	}
}

// Init инициализирует forwardedTo для новых чатов
func (s *Service) Init(forwardedTo map[int64]bool, dstChatIds []int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, dstChatId := range dstChatIds {
		_, ok := forwardedTo[dstChatId]
		if !ok {
			forwardedTo[dstChatId] = false
		}
	}
}

// Add добавляет чат в forwardedTo, если он еще не добавлен
func (s *Service) Add(forwardedTo map[int64]bool, dstChatId int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !forwardedTo[dstChatId] {
		forwardedTo[dstChatId] = true
		return true
	}
	return false
}
