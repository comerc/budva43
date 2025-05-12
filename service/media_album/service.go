package media_album

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/zelenin/go-tdlib/client"

	"github.com/comerc/budva43/entity"
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
	mediaAlbums map[entity.MediaAlbumForwardKey]*mediaAlbum
}

// New создает новый сервис для управления медиа-альбомами
func New() *Service {
	return &Service{
		log: slog.With("module", "service.media_album"),
		//
		mediaAlbums: make(map[entity.MediaAlbumForwardKey]*mediaAlbum),
	}
}

// https://github.com/tdlib/td/issues/1482
// AddMessage добавляет сообщение в медиа-альбом
// Возвращает true, если это первое сообщение в медиа-альбоме
func (s *Service) AddMessage(forwardRuleId entity.ForwardRuleId, message *client.Message) bool {
	forwardKey := s.GetForwardKey(forwardRuleId, message.MediaAlbumId)
	item, ok := s.mediaAlbums[forwardKey]
	if !ok {
		item = &mediaAlbum{}
	}
	item.messages = append(item.messages, message)
	item.lastReceived = time.Now()
	s.mediaAlbums[forwardKey] = item
	return !ok
}

// GetLastReceivedDiff возвращает время, прошедшее с момента получения последнего сообщения в медиа-альбоме
func (s *Service) GetLastReceivedDiff(forwardKey entity.MediaAlbumForwardKey) time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()
	return time.Since(s.mediaAlbums[forwardKey].lastReceived)
}

// GetMessages возвращает сообщения медиа-альбома
func (s *Service) GetMessages(forwardKey entity.MediaAlbumForwardKey) []*client.Message {
	s.mu.Lock()
	defer s.mu.Unlock()
	messages := s.mediaAlbums[forwardKey].messages
	delete(s.mediaAlbums, forwardKey)
	return messages
}

// GetForwardKey возвращает ключ для пересылаемого медиа-альбома
func (s *Service) GetForwardKey(forwardRuleId entity.ForwardRuleId, MediaAlbumId client.JsonInt64) entity.MediaAlbumForwardKey {
	return fmt.Sprintf("%s:%d", forwardRuleId, MediaAlbumId)
}
