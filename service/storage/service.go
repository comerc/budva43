package engine_storage

import (
	"fmt"
	"log/slog"
	"strings"
)

// TODO: выполнить корректный перенос из budva32:
// - Интерфейс для работы с хранилищем более формализован, но некоторые методы могут работать иначе, чем в старой версии
// - Особенно это касается функций работы с идентификаторами сообщений
// - Функционал инкрементирования счетчиков может отличаться от старой версии

const (
	// Префиксы ключей для хранения в BadgerDB
	CopiedMessageIdsPrefix  = "copiedMsgIds"
	NewMessageIdPrefix      = "newMsgId"
	TmpMessageIdPrefix      = "tmpMsgId"
	ViewedMessagesPrefix    = "viewedMsgs"
	ForwardedMessagesPrefix = "forwardedMsgs"
)

//go:generate mockery --name=storageRepo --exported
type storageRepo interface {
	Set(key, value string) error
	Get(key string) (string, error)
	Delete(key string) error
}

// Service предоставляет методы для хранения данных, специфичных для engine
type Service struct {
	log *slog.Logger
	//
	repo storageRepo
}

// New создает новый экземпляр сервиса хранения данных
func New(repo storageRepo) *Service {
	return &Service{
		log: slog.With("module", "service.storage"),
		//
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

// SetCopiedMessageId сохраняет связь между оригинальным и скопированным сообщением
func (s *Service) SetCopiedMessageId(fromChatMessageId string, toChatMessageId string) error {
	key := fmt.Sprintf("%s:%s", CopiedMessageIdsPrefix, fromChatMessageId)

	var val string
	var err error

	// Получаем текущий список скопированных сообщений
	val, err = s.repo.Get(key)
	if err != nil && err.Error() != "key not found" {
		return fmt.Errorf("ошибка получения значения: %w", err)
	}

	// Добавляем новое сообщение в список
	result := []string{}
	if len(val) > 0 {
		result = strings.Split(val, ",")
	}

	// Добавляем новый Id и удаляем дубликаты
	result = append(result, toChatMessageId)
	result = distinct(result)

	// Сохраняем обновленный список
	err = s.repo.Set(key, strings.Join(result, ","))
	if err != nil {
		return fmt.Errorf("ошибка сохранения значения: %w", err)
	}

	return nil
}

// GetCopiedMessageIds получает идентификаторы скопированных сообщений по Id оригинала
func (s *Service) GetCopiedMessageIds(fromChatMessageId string) ([]string, error) {
	key := fmt.Sprintf("%s:%s", CopiedMessageIdsPrefix, fromChatMessageId)

	val, err := s.repo.Get(key)
	if err != nil {
		if err.Error() == "key not found" {
			return []string{}, nil
		}
		return nil, fmt.Errorf("ошибка получения значения: %w", err)
	}

	toChatMessageIds := []string{}
	if len(val) > 0 {
		toChatMessageIds = strings.Split(val, ",")
	}

	s.log.Debug("получены скопированные сообщения",
		"fromChatMessageId", fromChatMessageId,
		"toChatMessageIds", toChatMessageIds)

	return toChatMessageIds, nil
}

// DeleteCopiedMessageIds удаляет связь между оригинальным и скопированными сообщениями
func (s *Service) DeleteCopiedMessageIds(fromChatMessageId string) error {
	key := fmt.Sprintf("%s:%s", CopiedMessageIdsPrefix, fromChatMessageId)

	err := s.repo.Delete(key)
	if err != nil && err.Error() != "key not found" {
		return fmt.Errorf("ошибка удаления значения: %w", err)
	}

	return nil
}

// SetNewMessageId сохраняет соответствие между временным и постоянным Id сообщения
func (s *Service) SetNewMessageId(chatId, tmpMessageId, newMessageId int64) error {
	key := fmt.Sprintf("%s:%d:%d", NewMessageIdPrefix, chatId, tmpMessageId)

	err := s.repo.Set(key, fmt.Sprintf("%d", newMessageId))
	if err != nil {
		return fmt.Errorf("ошибка сохранения значения: %w", err)
	}

	return nil
}

// GetNewMessageId получает постоянный Id сообщения по временному
func (s *Service) GetNewMessageId(chatId, tmpMessageId int64) (int64, error) {
	key := fmt.Sprintf("%s:%d:%d", NewMessageIdPrefix, chatId, tmpMessageId)

	val, err := s.repo.Get(key)
	if err != nil {
		if err.Error() == "key not found" {
			return 0, nil
		}
		return 0, fmt.Errorf("ошибка получения значения: %w", err)
	}

	var newMessageId int64
	if _, err := fmt.Sscanf(val, "%d", &newMessageId); err != nil {
		return 0, fmt.Errorf("ошибка преобразования newMessageId: %w", err)
	}

	return newMessageId, nil
}

// DeleteNewMessageId удаляет соответствие между временным и постоянным Id сообщения
func (s *Service) DeleteNewMessageId(chatId, tmpMessageId int64) error {
	key := fmt.Sprintf("%s:%d:%d", NewMessageIdPrefix, chatId, tmpMessageId)

	err := s.repo.Delete(key)
	if err != nil && err.Error() != "key not found" {
		return fmt.Errorf("ошибка удаления значения: %w", err)
	}

	return nil
}

// SetTmpMessageId сохраняет соответствие между постоянным и временным Id сообщения
func (s *Service) SetTmpMessageId(chatId, newMessageId, tmpMessageId int64) error {
	key := fmt.Sprintf("%s:%d:%d", TmpMessageIdPrefix, chatId, newMessageId)

	err := s.repo.Set(key, fmt.Sprintf("%d", tmpMessageId))
	if err != nil {
		return fmt.Errorf("ошибка сохранения значения: %w", err)
	}

	return nil
}

// GetTmpMessageId получает временный Id сообщения по постоянному
func (s *Service) GetTmpMessageId(chatId, newMessageId int64) (int64, error) {
	key := fmt.Sprintf("%s:%d:%d", TmpMessageIdPrefix, chatId, newMessageId)

	val, err := s.repo.Get(key)
	if err != nil {
		if err.Error() == "key not found" {
			return 0, nil
		}
		return 0, fmt.Errorf("ошибка получения значения: %w", err)
	}

	var tmpMessageId int64
	if _, err := fmt.Sscanf(val, "%d", &tmpMessageId); err != nil {
		return 0, fmt.Errorf("ошибка преобразования tmpMessageId: %w", err)
	}

	return tmpMessageId, nil
}

// DeleteTmpMessageId удаляет соответствие между постоянным и временным Id сообщения
func (s *Service) DeleteTmpMessageId(chatId, newMessageId int64) error {
	key := fmt.Sprintf("%s:%d:%d", TmpMessageIdPrefix, chatId, newMessageId)

	err := s.repo.Delete(key)
	if err != nil && err.Error() != "key not found" {
		return fmt.Errorf("ошибка удаления значения: %w", err)
	}

	return nil
}

// IncrementViewedMessages увеличивает счетчик просмотренных сообщений
func (s *Service) IncrementViewedMessages(toChatId int64) error {
	key := fmt.Sprintf("%s:%d", ViewedMessagesPrefix, toChatId)

	val, err := s.repo.Get(key)
	if err != nil && err.Error() != "key not found" {
		return fmt.Errorf("ошибка получения значения: %w", err)
	}

	var count int64
	if len(val) > 0 {
		if _, err := fmt.Sscanf(val, "%d", &count); err != nil {
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

const answerMessageIdPrefix = "answerMsgId"

// SetAnswerMessageId устанавливает идентификатор сообщения ответа
func (s *Service) SetAnswerMessageId(dstChatId, tmpMessageId int64, fromChatMessageId string) {
	// TODO: выполнить корректный перенос из budva32
}

// GetAnswerMessageId возвращает идентификатор сообщения ответа
func (s *Service) GetAnswerMessageId(dstChatId, tmpMessageId int64) string {
	// TODO: выполнить корректный перенос из budva32
	return ""
}

// DeleteAnswerMessageId удаляет идентификатор сообщения ответа
func (s *Service) DeleteAnswerMessageId(dstChatId, tmpMessageId int64) {
	// TODO: выполнить корректный перенос из budva32
}

// IncrementForwardedMessages увеличивает счетчик пересланных сообщений
func (s *Service) IncrementForwardedMessages(toChatId int64) error {
	key := fmt.Sprintf("%s:%d", ForwardedMessagesPrefix, toChatId)

	val, err := s.repo.Get(key)
	if err != nil && err.Error() != "key not found" {
		return fmt.Errorf("ошибка получения значения: %w", err)
	}

	var count int64
	if len(val) > 0 {
		if _, err := fmt.Sscanf(val, "%d", &count); err != nil {
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
