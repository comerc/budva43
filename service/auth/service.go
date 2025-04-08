package auth

import (
	"errors"
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

// AuthService предоставляет методы для аутентификации и авторизации
type AuthService struct {
	authenticator authenticator
	storage       storageService
	sessions      map[string]time.Time
	sessionsLock  sync.RWMutex
}

// NewAuthService создает новый экземпляр сервиса для аутентификации
func NewAuthService(authenticator authenticator, storage storageService) *AuthService {
	return &AuthService{
		authenticator: authenticator,
		storage:       storage,
		sessions:      make(map[string]time.Time),
	}
}

// Authenticate выполняет аутентификацию с заданными учетными данными
func (s *AuthService) Authenticate(credentials string) (string, error) {
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
			// Нелитричная ошибка, можем продолжить
			// Логировать ошибку
		}
	}

	return token, nil
}

// ValidateToken проверяет валидность токена
func (s *AuthService) ValidateToken(token string) (bool, error) {
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
func (s *AuthService) RefreshToken(oldToken string) (string, error) {
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
			// Нелитричная ошибка, можем продолжить
			// Логировать ошибку
		}
	}

	return newToken, nil
}

// InvalidateToken аннулирует токен
func (s *AuthService) InvalidateToken(token string) error {
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
				// Нелитричная ошибка, можем продолжить
				// Логировать ошибку
			}
		}
	}

	return nil
}

// CleanupExpiredSessions очищает истекшие сессии
func (s *AuthService) CleanupExpiredSessions() {
	s.sessionsLock.Lock()
	defer s.sessionsLock.Unlock()

	now := time.Now()
	for token, expiryTime := range s.sessions {
		if now.After(expiryTime) {
			delete(s.sessions, token)
		}
	}
}
