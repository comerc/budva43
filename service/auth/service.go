package auth

import (
	"errors"
	"log/slog"
	"sync"
	"time"
)

//go:generate mockery --name=storageService --exported
type storageService interface {
	Set(key string, value []byte) error
	Get(key string) ([]byte, error)
	Delete(key string) error
}

//go:generate mockery --name=authenticator --exported
type authenticator interface {
	Authenticate(credentials string) (bool, error)
	GetToken() (string, error)
	RefreshToken() (string, error)
}

// Service предоставляет методы для аутентификации и авторизации
type Service struct {
	log *slog.Logger
	//
	authenticator authenticator
	storage       storageService
	sessions      map[string]time.Time
	sessionsLock  sync.RWMutex
}

// New создает новый экземпляр сервиса для аутентификации
func New(authenticator authenticator, storage storageService) *Service {
	return &Service{
		log: slog.With("module", "service.auth"),
		//
		authenticator: authenticator,
		storage:       storage,
		sessions:      make(map[string]time.Time),
	}
}

// Authenticate выполняет аутентификацию с заданными учетными данными
func (s *Service) Authenticate(credentials string) (string, error) {
	if s.authenticator == nil {
		return "", errors.New("authenticator is nil")
	}

	// Выполняем аутентификацию
	success, err := s.authenticator.Authenticate(credentials)
	if err != nil {
		return "", err
	}

	if !success {
		return "", errors.New("authentication failed")
	}

	// Получаем токен
	token, err := s.authenticator.GetToken()
	if err != nil {
		return "", err
	}

	// Сохраняем сессию
	s.sessionsLock.Lock()
	s.sessions[token] = time.Now().Add(24 * time.Hour) // Токен действителен 24 часа
	s.sessionsLock.Unlock()

	// Если есть хранилище, сохраняем токен в нем
	if s.storage != nil {
		err = s.storage.Set("auth:token", []byte(token))
		if err != nil {
			// Некритичная ошибка, можем продолжить
			s.log.Error("Failed to save token in storage", "err", err)
		}
	}

	return token, nil
}

// ValidateToken проверяет валидность токена
func (s *Service) ValidateToken(token string) (bool, error) {
	if token == "" {
		return false, errors.New("empty token")
	}

	s.sessionsLock.RLock()
	expiryTime, exists := s.sessions[token]
	s.sessionsLock.RUnlock()

	// Если токен существует и не истек срок его действия
	if exists && time.Now().Before(expiryTime) {
		return true, nil
	}

	// Если у нас есть хранилище, проверяем токен там
	if s.storage != nil {
		storedToken, err := s.storage.Get("auth:token")
		if err == nil && string(storedToken) == token {
			// Обновляем сессию в памяти
			s.sessionsLock.Lock()
			s.sessions[token] = time.Now().Add(24 * time.Hour)
			s.sessionsLock.Unlock()
			return true, nil
		}
	}

	return false, nil
}

// RefreshToken обновляет токен аутентификации
func (s *Service) RefreshToken(oldToken string) (string, error) {
	if s.authenticator == nil {
		return "", errors.New("authenticator is nil")
	}

	// Проверяем валидность старого токена
	valid, err := s.ValidateToken(oldToken)
	if err != nil {
		return "", err
	}

	if !valid {
		return "", errors.New("invalid or expired token")
	}

	// Получаем новый токен
	newToken, err := s.authenticator.RefreshToken()
	if err != nil {
		return "", err
	}

	// Обновляем сессию
	s.sessionsLock.Lock()
	delete(s.sessions, oldToken)
	s.sessions[newToken] = time.Now().Add(24 * time.Hour)
	s.sessionsLock.Unlock()

	// Если есть хранилище, обновляем токен в нем
	if s.storage != nil {
		err = s.storage.Set("auth:token", []byte(newToken))
		if err != nil {
			// Некритричная ошибка, можем продолжить
			s.log.Error("Failed to update token in storage", "err", err)
		}
	}

	return newToken, nil
}

// InvalidateToken аннулирует токен
func (s *Service) InvalidateToken(token string) error {
	if token == "" {
		return errors.New("empty token")
	}

	// Удаляем сессию
	s.sessionsLock.Lock()
	delete(s.sessions, token)
	s.sessionsLock.Unlock()

	// Если есть хранилище, удаляем токен из него
	if s.storage != nil {
		storedToken, err := s.storage.Get("auth:token")
		if err == nil && string(storedToken) == token {
			err = s.storage.Delete("auth:token")
			if err != nil {
				// Некритичная ошибка, можем продолжить
				s.log.Error("Failed to delete token from storage", "err", err)
			}
		}
	}

	return nil
}

// CleanupExpiredSessions очищает истекшие сессии
func (s *Service) CleanupExpiredSessions() {
	s.sessionsLock.Lock()
	defer s.sessionsLock.Unlock()

	now := time.Now()
	for token, expiryTime := range s.sessions {
		if now.After(expiryTime) {
			delete(s.sessions, token)
		}
	}
}
