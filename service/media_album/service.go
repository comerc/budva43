package media_album

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/zelenin/go-tdlib/client"
)

// MediaAlbum представляет группу сообщений, составляющих медиа-альбом
type MediaAlbum struct {
	Messages     []*client.Message
	LastReceived time.Time
	ForwardKey   string
}

// Service управляет медиа-альбомами
type Service struct {
	log         *slog.Logger
	mu          sync.Mutex
	mediaAlbums map[string]*MediaAlbum
}

// New создает новый сервис для управления медиа-альбомами
func New() *Service {
	return &Service{
		log:         slog.With("module", "service.media_album"),
		mediaAlbums: make(map[string]*MediaAlbum),
	}
}

// AddMessage добавляет сообщение в медиа-альбом
// Возвращает true, если это первое сообщение в альбоме
func (s *Service) AddMessage(forwardKey string, message *client.Message) bool {
	if message.MediaAlbumId == 0 {
		return false
	}

	key := s.getKey(forwardKey, message.MediaAlbumId)

	s.mu.Lock()
	defer s.mu.Unlock()

	album, ok := s.mediaAlbums[key]
	if !ok {
		album = &MediaAlbum{
			Messages:     make([]*client.Message, 0),
			LastReceived: time.Now(),
			ForwardKey:   forwardKey,
		}
		s.mediaAlbums[key] = album
		album.Messages = append(album.Messages, message)
		return true
	}

	album.Messages = append(album.Messages, message)
	album.LastReceived = time.Now()
	return false
}

// GetLastReceivedDiff возвращает время, прошедшее с момента получения последнего сообщения в альбоме
func (s *Service) GetLastReceivedDiff(forwardKey string, albumID client.JsonInt64) time.Duration {
	key := s.getKey(forwardKey, albumID)

	s.mu.Lock()
	defer s.mu.Unlock()

	album, ok := s.mediaAlbums[key]
	if !ok {
		return 24 * time.Hour // Очень большое значение, если альбом не найден
	}

	return time.Since(album.LastReceived)
}

// GetMessages возвращает сообщения альбома
func (s *Service) GetMessages(forwardKey string, albumID client.JsonInt64) []*client.Message {
	key := s.getKey(forwardKey, albumID)

	s.mu.Lock()
	defer s.mu.Unlock()

	album, ok := s.mediaAlbums[key]
	if !ok {
		return nil
	}

	// Делаем копию сообщений, чтобы избежать проблем с конкурентным доступом
	result := make([]*client.Message, len(album.Messages))
	copy(result, album.Messages)

	// Удаляем альбом после чтения
	delete(s.mediaAlbums, key)

	return result
}

// getKey возвращает ключ для хранения альбома
func (s *Service) getKey(forwardKey string, albumID client.JsonInt64) string {
	return forwardKey + ":" + fmt.Sprintf("%d", int64(albumID))
}
