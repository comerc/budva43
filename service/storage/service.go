package storage

import (
	"context"
	"errors"
	"log/slog"
	"time"
)

//go:generate mockery --name=storageRepo --exported
type storageRepo interface {
	Get(key []byte) ([]byte, error)
	Set(key, value []byte) error
	SetWithTTL(key, value []byte, ttl time.Duration) error
	Delete(key []byte) error
	Close() error
}

// Service предоставляет методы для работы с хранилищем
type Service struct {
	log *slog.Logger
	//
	repo storageRepo
}

// New создает новый экземпляр сервиса для работы с хранилищем
func New(repo storageRepo) *Service {
	return &Service{
		log: slog.With("module", "service.storage"),
		//
		repo: repo,
	}
}

// Get получает значение по ключу
func (s *Service) Get(key string) ([]byte, error) {
	if s.repo == nil {
		return nil, errors.New("storage repo is nil")
	}
	return s.repo.Get([]byte(key))
}

// Set устанавливает значение по ключу
func (s *Service) Set(key string, value []byte) error {
	if s.repo == nil {
		return errors.New("storage repo is nil")
	}
	return s.repo.Set([]byte(key), value)
}

// SetWithTTL устанавливает значение по ключу с временем жизни
func (s *Service) SetWithTTL(key string, value []byte, ttl time.Duration) error {
	if s.repo == nil {
		return errors.New("storage repo is nil")
	}
	return s.repo.SetWithTTL([]byte(key), value, ttl)
}

// Delete удаляет значение по ключу
func (s *Service) Delete(key string) error {
	if s.repo == nil {
		return errors.New("storage repo is nil")
	}
	return s.repo.Delete([]byte(key))
}

// SaveMedia сохраняет медиа-файл в хранилище
func (s *Service) SaveMedia(key string, content []byte) error {
	return s.Set("media:"+key, content)
}

// GetMedia получает медиа-файл из хранилища
func (s *Service) GetMedia(key string) ([]byte, bool, error) {
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
func (s *Service) SaveConfig(key string, content []byte) error {
	return s.Set("config:"+key, content)
}

// GetConfig получает конфигурацию из хранилища
func (s *Service) GetConfig(key string) ([]byte, error) {
	return s.Get("config:" + key)
}

// Close закрывает соединение с хранилищем
func (s *Service) Close() error {
	if s.repo == nil {
		return nil
	}
	return s.repo.Close()
}

// Shutdown корректно завершает работу с хранилищем
func (s *Service) Shutdown(ctx context.Context) error {
	// Можно добавить дополнительную логику завершения работы
	// Например, ожидание завершения всех операций

	// Но в простейшем случае просто закрываем соединение
	return s.Close()
}
