package media_album

import (
	"log/slog"
	"sync"
	"time"

	"github.com/zelenin/go-tdlib/client"
)

// mediaAlbum представляет группу сообщений, составляющих медиа-альбом
type mediaAlbum struct {
	messages     []*client.Message
	lastReceived time.Time
}

// Service управляет медиа-альбомами
type Service struct {
	log *slog.Logger
	//
	mu          sync.Mutex
	mediaAlbums map[string]*mediaAlbum
}

// New создает новый сервис для управления медиа-альбомами
func New() *Service {
	return &Service{
		log: slog.With("module", "service.media_album"),
		//
		mediaAlbums: make(map[string]*mediaAlbum),
	}
}

// AddMessage добавляет сообщение в медиа-альбом
// Возвращает true, если это первое сообщение в альбоме
func (s *Service) AddMessage(forwardKey string, message *client.Message) bool {
	// TODO: выполнить корректный перенос из budva32
	return false
}

// GetLastReceivedDiff возвращает время, прошедшее с момента получения последнего сообщения в альбоме
func (s *Service) GetLastReceivedDiff(key string) time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()
	return time.Since(s.mediaAlbums[key].lastReceived)
}

// GetMessages возвращает сообщения альбома
func (s *Service) GetMessages(key string) []*client.Message {
	// TODO: выполнить корректный перенос из budva32
	return nil
}
