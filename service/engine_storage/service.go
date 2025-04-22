package engine_storage

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/comerc/budva43/config"
	"github.com/comerc/budva43/entity"
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
	Set(key string, value []byte) error
	Get(key string) ([]byte, error)
	Delete(key string) error
}

// Service предоставляет методы для хранения данных, специфичных для engine
type Service struct {
	log  *slog.Logger
	repo storageRepo
}

// New создает новый экземпляр сервиса хранения данных engine
func New(repo storageRepo) *Service {
	return &Service{
		log:  slog.With("module", "service.engine_storage"),
		repo: repo,
	}
}

// ParseToChatMessageID парсит строку toChatMessageID в формате "ruleID:chatID:messageID"
func ParseToChatMessageID(toChatMessageID string) (ruleID string, chatID int64, messageID int64, err error) {
	parts := strings.Split(toChatMessageID, ":")
	if len(parts) != 3 {
		return "", 0, 0, fmt.Errorf("неверный формат toChatMessageID: %s", toChatMessageID)
	}

	ruleID = parts[0]

	var chatIDInt, messageIDInt int
	if _, err := fmt.Sscanf(parts[1], "%d", &chatIDInt); err != nil {
		return "", 0, 0, fmt.Errorf("ошибка преобразования chatID: %w", err)
	}

	if _, err := fmt.Sscanf(parts[2], "%d", &messageIDInt); err != nil {
		return "", 0, 0, fmt.Errorf("ошибка преобразования messageID: %w", err)
	}

	return ruleID, int64(chatIDInt), int64(messageIDInt), nil
}

// GetRuleByID находит правило по его идентификатору
func (s *Service) GetRuleByID(ruleID string) (rule entity.ForwardRule, ok bool) {
	rule, ok = config.Engine.Forwards[ruleID]
	return
}

// SetCopiedMessageID сохраняет связь между оригинальным и скопированным сообщением
func (s *Service) SetCopiedMessageID(fromChatMessageID string, toChatMessageID string) error {
	// Реализация будет добавлена позже
	return nil
}

// GetCopiedMessageIDs получает идентификаторы скопированных сообщений по ID оригинала
func (s *Service) GetCopiedMessageIDs(fromChatMessageID string) ([]string, error) {
	// Реализация будет добавлена позже
	return nil, nil
}

// DeleteCopiedMessageIDs удаляет связь между оригинальным и скопированными сообщениями
func (s *Service) DeleteCopiedMessageIDs(fromChatMessageID string) error {
	// Реализация будет добавлена позже
	return nil
}

// SetNewMessageID сохраняет соответствие между временным и постоянным ID сообщения
func (s *Service) SetNewMessageID(chatID, tmpMessageID, newMessageID int64) error {
	// Реализация будет добавлена позже
	return nil
}

// GetNewMessageID получает постоянный ID сообщения по временному
func (s *Service) GetNewMessageID(chatID, tmpMessageID int64) (int64, error) {
	// Реализация будет добавлена позже
	return 0, nil
}

// DeleteNewMessageID удаляет соответствие между временным и постоянным ID сообщения
func (s *Service) DeleteNewMessageID(chatID, tmpMessageID int64) error {
	// Реализация будет добавлена позже
	return nil
}

// SetTmpMessageID сохраняет соответствие между постоянным и временным ID сообщения
func (s *Service) SetTmpMessageID(chatID, newMessageID, tmpMessageID int64) error {
	// Реализация будет добавлена позже
	return nil
}

// GetTmpMessageID получает временный ID сообщения по постоянному
func (s *Service) GetTmpMessageID(chatID, newMessageID int64) (int64, error) {
	// Реализация будет добавлена позже
	return 0, nil
}

// DeleteTmpMessageID удаляет соответствие между постоянным и временным ID сообщения
func (s *Service) DeleteTmpMessageID(chatID, newMessageID int64) error {
	// Реализация будет добавлена позже
	return nil
}

// IncrementViewedMessages увеличивает счетчик просмотренных сообщений
func (s *Service) IncrementViewedMessages(toChatID int64) error {
	// Реализация будет добавлена позже
	return nil
}

// IncrementForwardedMessages увеличивает счетчик пересланных сообщений
func (s *Service) IncrementForwardedMessages(toChatID int64) error {
	// Реализация будет добавлена позже
	return nil
}
