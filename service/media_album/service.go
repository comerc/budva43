package media_album

import (
	"fmt"
	"sync"
	"time"

	"github.com/zelenin/go-tdlib/client"

	"github.com/comerc/budva43/app/domain"
	"github.com/comerc/budva43/app/log"
)

// mediaAlbum представляет группу сообщений, составляющих медиа-альбом
type mediaAlbum struct {
	messages     []*client.Message
	lastReceived time.Time
}

// Service управляет медиа-альбомами
type Service struct {
	log *log.Logger
	//
	mu          sync.Mutex
	mediaAlbums map[domain.MediaAlbumKey]*mediaAlbum
}

// New создает новый сервис для управления медиа-альбомами
func New() *Service {
	return &Service{
		log: log.NewLogger(),
		//
		mediaAlbums: make(map[domain.MediaAlbumKey]*mediaAlbum),
	}
}

// https://github.com/tdlib/td/issues/1482
// AddMessage добавляет сообщение в медиа-альбом
// Возвращает true, если это первое сообщение в медиа-альбоме
func (s *Service) AddMessage(key domain.MediaAlbumKey, message *client.Message) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
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
func (s *Service) GetLastReceivedDiff(key domain.MediaAlbumKey) time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()
	return time.Since(s.mediaAlbums[key].lastReceived)
}

// PopMessages возвращает сообщения медиа-альбома и удаляет его
func (s *Service) PopMessages(key domain.MediaAlbumKey) []*client.Message {
	s.mu.Lock()
	defer s.mu.Unlock()
	messages := s.mediaAlbums[key].messages
	delete(s.mediaAlbums, key)
	return messages
}

// GetKey возвращает ключ для пересылаемого медиа-альбома
func (s *Service) GetKey(forwardRuleId domain.ForwardRuleId, mediaAlbumId client.JsonInt64) domain.MediaAlbumKey {
	return fmt.Sprintf("%s:%d", forwardRuleId, mediaAlbumId)
}
