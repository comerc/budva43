package engine_storage

import (
	"fmt"
	"log/slog"
	"strings"
)

const (
	// Префиксы ключей для хранения в BadgerDB
	CopiedMessageIDsPrefix  = "copiedMsgIds"
	NewMessageIDPrefix      = "newMsgId"
	TmpMessageIDPrefix      = "tmpMsgId"
	ViewedMessagesPrefix    = "viewedMsgs"
	ForwardedMessagesPrefix = "forwardedMsgs"
)

//go:generate mockery --name=storageRepo --exported
type storageRepo interface {
	Set(key, value string) error
	Get(key string) ([]byte, error)
	Delete(key string) error
}

// Service предоставляет методы для хранения данных, специфичных для engine
type Service struct {
	log  *slog.Logger
	repo storageRepo
}

// New создает новый экземпляр сервиса хранения данных
func New(repo storageRepo) *Service {
	return &Service{
		log:  slog.With("module", "service.storage"),
		repo: repo,
	}
}

// distinct удаляет дубликаты из слайса строк
func distinct(slice []string) []string {
	keys := make(map[string]bool)
	var list []string
	for _, entry := range slice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

// SetCopiedMessageID сохраняет связь между оригинальным и скопированным сообщением
func (s *Service) SetCopiedMessageID(fromChatMessageID string, toChatMessageID string) error {
	key := fmt.Sprintf("%s:%s", CopiedMessageIDsPrefix, fromChatMessageID)

	var val []byte
	var err error

	// Получаем текущий список скопированных сообщений
	val, err = s.repo.Get(key)
	if err != nil && err.Error() != "key not found" {
		return fmt.Errorf("ошибка получения значения: %w", err)
	}

	// Добавляем новое сообщение в список
	result := []string{}
	if len(val) > 0 {
		result = strings.Split(string(val), ",")
	}

	// Добавляем новый ID и удаляем дубликаты
	result = append(result, toChatMessageID)
	result = distinct(result)

	// Сохраняем обновленный список
	err = s.repo.Set(key, strings.Join(result, ","))
	if err != nil {
		return fmt.Errorf("ошибка сохранения значения: %w", err)
	}

	return nil
}

// GetCopiedMessageIDs получает идентификаторы скопированных сообщений по ID оригинала
func (s *Service) GetCopiedMessageIDs(fromChatMessageID string) ([]string, error) {
	key := fmt.Sprintf("%s:%s", CopiedMessageIDsPrefix, fromChatMessageID)

	val, err := s.repo.Get(key)
	if err != nil {
		if err.Error() == "key not found" {
			return []string{}, nil
		}
		return nil, fmt.Errorf("ошибка получения значения: %w", err)
	}

	toChatMessageIDs := []string{}
	if len(val) > 0 {
		toChatMessageIDs = strings.Split(string(val), ",")
	}

	s.log.Debug("получены скопированные сообщения",
		"fromChatMessageID", fromChatMessageID,
		"toChatMessageIDs", toChatMessageIDs)

	return toChatMessageIDs, nil
}

// DeleteCopiedMessageIDs удаляет связь между оригинальным и скопированными сообщениями
func (s *Service) DeleteCopiedMessageIDs(fromChatMessageID string) error {
	key := fmt.Sprintf("%s:%s", CopiedMessageIDsPrefix, fromChatMessageID)

	err := s.repo.Delete(key)
	if err != nil && err.Error() != "key not found" {
		return fmt.Errorf("ошибка удаления значения: %w", err)
	}

	return nil
}

// SetNewMessageID сохраняет соответствие между временным и постоянным ID сообщения
func (s *Service) SetNewMessageID(chatID, tmpMessageID, newMessageID int64) error {
	key := fmt.Sprintf("%s:%d:%d", NewMessageIDPrefix, chatID, tmpMessageID)

	err := s.repo.Set(key, fmt.Sprintf("%d", newMessageID))
	if err != nil {
		return fmt.Errorf("ошибка сохранения значения: %w", err)
	}

	return nil
}

// GetNewMessageID получает постоянный ID сообщения по временному
func (s *Service) GetNewMessageID(chatID, tmpMessageID int64) (int64, error) {
	key := fmt.Sprintf("%s:%d:%d", NewMessageIDPrefix, chatID, tmpMessageID)

	val, err := s.repo.Get(key)
	if err != nil {
		if err.Error() == "key not found" {
			return 0, nil
		}
		return 0, fmt.Errorf("ошибка получения значения: %w", err)
	}

	var newMessageID int64
	if _, err := fmt.Sscanf(string(val), "%d", &newMessageID); err != nil {
		return 0, fmt.Errorf("ошибка преобразования newMessageID: %w", err)
	}

	return newMessageID, nil
}

// DeleteNewMessageID удаляет соответствие между временным и постоянным ID сообщения
func (s *Service) DeleteNewMessageID(chatID, tmpMessageID int64) error {
	key := fmt.Sprintf("%s:%d:%d", NewMessageIDPrefix, chatID, tmpMessageID)

	err := s.repo.Delete(key)
	if err != nil && err.Error() != "key not found" {
		return fmt.Errorf("ошибка удаления значения: %w", err)
	}

	return nil
}

// SetTmpMessageID сохраняет соответствие между постоянным и временным ID сообщения
func (s *Service) SetTmpMessageID(chatID, newMessageID, tmpMessageID int64) error {
	key := fmt.Sprintf("%s:%d:%d", TmpMessageIDPrefix, chatID, newMessageID)

	err := s.repo.Set(key, fmt.Sprintf("%d", tmpMessageID))
	if err != nil {
		return fmt.Errorf("ошибка сохранения значения: %w", err)
	}

	return nil
}

// GetTmpMessageID получает временный ID сообщения по постоянному
func (s *Service) GetTmpMessageID(chatID, newMessageID int64) (int64, error) {
	key := fmt.Sprintf("%s:%d:%d", TmpMessageIDPrefix, chatID, newMessageID)

	val, err := s.repo.Get(key)
	if err != nil {
		if err.Error() == "key not found" {
			return 0, nil
		}
		return 0, fmt.Errorf("ошибка получения значения: %w", err)
	}

	var tmpMessageID int64
	if _, err := fmt.Sscanf(string(val), "%d", &tmpMessageID); err != nil {
		return 0, fmt.Errorf("ошибка преобразования tmpMessageID: %w", err)
	}

	return tmpMessageID, nil
}

// DeleteTmpMessageID удаляет соответствие между постоянным и временным ID сообщения
func (s *Service) DeleteTmpMessageID(chatID, newMessageID int64) error {
	key := fmt.Sprintf("%s:%d:%d", TmpMessageIDPrefix, chatID, newMessageID)

	err := s.repo.Delete(key)
	if err != nil && err.Error() != "key not found" {
		return fmt.Errorf("ошибка удаления значения: %w", err)
	}

	return nil
}

// IncrementViewedMessages увеличивает счетчик просмотренных сообщений
func (s *Service) IncrementViewedMessages(toChatID int64) error {
	key := fmt.Sprintf("%s:%d", ViewedMessagesPrefix, toChatID)

	val, err := s.repo.Get(key)
	if err != nil && err.Error() != "key not found" {
		return fmt.Errorf("ошибка получения значения: %w", err)
	}

	var count int64
	if len(val) > 0 {
		if _, err := fmt.Sscanf(string(val), "%d", &count); err != nil {
			return fmt.Errorf("ошибка преобразования счетчика просмотренных сообщений: %w", err)
		}
	}

	count++

	err = s.repo.Set(key, fmt.Sprintf("%d", count))
	if err != nil {
		return fmt.Errorf("ошибка сохранения значения: %w", err)
	}

	return nil
}

// IncrementForwardedMessages увеличивает счетчик пересланных сообщений
func (s *Service) IncrementForwardedMessages(toChatID int64) error {
	key := fmt.Sprintf("%s:%d", ForwardedMessagesPrefix, toChatID)

	val, err := s.repo.Get(key)
	if err != nil && err.Error() != "key not found" {
		return fmt.Errorf("ошибка получения значения: %w", err)
	}

	var count int64
	if len(val) > 0 {
		if _, err := fmt.Sscanf(string(val), "%d", &count); err != nil {
			return fmt.Errorf("ошибка преобразования счетчика пересланных сообщений: %w", err)
		}
	}

	count++

	err = s.repo.Set(key, fmt.Sprintf("%d", count))
	if err != nil {
		return fmt.Errorf("ошибка сохранения значения: %w", err)
	}

	return nil
}
