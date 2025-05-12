package media_album

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/comerc/budva43/entity"
	"github.com/zelenin/go-tdlib/client"
)

type Key = string

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
	mediaAlbums map[Key]*mediaAlbum
}

// New создает новый сервис для управления медиа-альбомами
func New() *Service {
	return &Service{
		log: slog.With("module", "service.media_album"),
		//
		mediaAlbums: make(map[Key]*mediaAlbum),
	}
}

// https://github.com/tdlib/td/issues/1482
// AddMessage добавляет сообщение в медиа-альбом
// Возвращает true, если это первое сообщение в медиа-альбоме
func (s *Service) AddMessage(forwardRuleId entity.ForwardRuleId, message *client.Message) bool {
	key := s.GetKey(forwardRuleId, message.MediaAlbumId)
	item, ok := s.mediaAlbums[key]
	if !ok {
		item = &mediaAlbum{}
	}
	item.messages = append(item.messages, message)
	item.lastReceived = time.Now()
	s.mediaAlbums[key] = item
	return !ok
}

// GetLastReceivedDiff возвращает время, прошедшее с момента получения последнего сообщения в медиа-альбоме
func (s *Service) GetLastReceivedDiff(key string) time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()
	return time.Since(s.mediaAlbums[key].lastReceived)
}

// GetMessages возвращает сообщения медиа-альбома
func (s *Service) GetMessages(key string) []*client.Message {
	// TODO: выполнить корректный перенос из budva32
	return nil
}

// GetKey возвращает ключ для пересылаемого медиа-альбома
func (s *Service) GetKey(forwardRuleId entity.ForwardRuleId, MediaAlbumId client.JsonInt64) Key {
	return fmt.Sprintf("%s:%d", forwardRuleId, MediaAlbumId)
}
