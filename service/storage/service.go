package storage

import (
	"context"
	"errors"
	"time"
)

//go:generate mockery --name=storageRepository --exported
type storageRepository interface {
	Get(key []byte) ([]byte, error)
	Set(key, value []byte) error
	SetWithTTL(key, value []byte, ttl time.Duration) error
	Delete(key []byte) error
	Close() error
}

// StorageService предоставляет методы для работы с хранилищем
type StorageService struct {
	repository storageRepository
}

// NewStorageService создает новый экземпляр сервиса для работы с хранилищем
func NewStorageService(repository storageRepository) *StorageService {
	return &StorageService{
		repository: repository,
	}
}

// Get получает значение по ключу
func (s *StorageService) Get(key string) ([]byte, error) {
	if s.repository == nil {
		return nil, errors.New("storage repository is nil")
	}
	return s.repository.Get([]byte(key))
}

// Set устанавливает значение по ключу
func (s *StorageService) Set(key string, value []byte) error {
	if s.repository == nil {
		return errors.New("storage repository is nil")
	}
	return s.repository.Set([]byte(key), value)
}

// SetWithTTL устанавливает значение по ключу с временем жизни
func (s *StorageService) SetWithTTL(key string, value []byte, ttl time.Duration) error {
	if s.repository == nil {
		return errors.New("storage repository is nil")
	}
	return s.repository.SetWithTTL([]byte(key), value, ttl)
}

// Delete удаляет значение по ключу
func (s *StorageService) Delete(key string) error {
	if s.repository == nil {
		return errors.New("storage repository is nil")
	}
	return s.repository.Delete([]byte(key))
}

// SaveMedia сохраняет медиа-файл в хранилище
func (s *StorageService) SaveMedia(key string, content []byte) error {
	return s.Set("media:"+key, content)
}

// GetMedia получает медиа-файл из хранилища
func (s *StorageService) GetMedia(key string) ([]byte, bool, error) {
	content, err := s.Get("media:" + key)
	if err != nil {
		if errors.Is(err, errors.New("key not found")) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return content, true, nil
}

// SaveConfig сохраняет конфигурацию в хранилище
func (s *StorageService) SaveConfig(key string, content []byte) error {
	return s.Set("config:"+key, content)
}

// GetConfig получает конфигурацию из хранилища
func (s *StorageService) GetConfig(key string) ([]byte, error) {
	return s.Get("config:" + key)
}

// Close закрывает соединение с хранилищем
func (s *StorageService) Close() error {
	if s.repository == nil {
		return nil
	}
	return s.repository.Close()
}

// Shutdown корректно завершает работу с хранилищем
func (s *StorageService) Shutdown(ctx context.Context) error {
	// Можно добавить дополнительную логику завершения работы
	// Например, ожидание завершения всех операций

	// Но в простейшем случае просто закрываем соединение
	return s.Close()
}
