package engine_storage

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/comerc/budva43/util"
)

// TODO: выполнить корректный перенос из budva32:
// - Интерфейс для работы с хранилищем более формализован, но некоторые методы могут работать иначе, чем в старой версии
// - Особенно это касается функций работы с идентификаторами сообщений
// - Функционал инкрементирования счетчиков может отличаться от старой версии

const (
	// Префиксы ключей для хранения в BadgerDB
	copiedMessageIdsPrefix  = "copiedMsgIds"
	newMessageIdPrefix      = "newMsgId"
	tmpMessageIdPrefix      = "tmpMsgId"
	viewedMessagesPrefix    = "viewedMsgs"
	forwardedMessagesPrefix = "forwardedMsgs"
)

//go:generate mockery --name=storageRepo --exported
type storageRepo interface {
	GetSet(key string, fn func(val string) (string, error)) (string, error)
	Set(key, value string) error
	Get(key string) (string, error)
	Delete(key string) error
	Increment(key string) (string, error)
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

// SetCopiedMessageId сохраняет связь между оригинальным и скопированным сообщением
func (s *Service) SetCopiedMessageId(fromChatMessageId string, toChatMessageId string) error {
	key := fmt.Sprintf("%s:%s", copiedMessageIdsPrefix, fromChatMessageId)
	var (
		err error
		val string
	)

	fn := func(val string) (string, error) {
		var ss []string
		if val != "" {
			// workaround https://stackoverflow.com/questions/28330908/how-to-string-split-an-empty-string-in-go
			ss = strings.Split(val, ",")
		}
		ss = append(ss, toChatMessageId)
		ss = util.Distinct(ss)
		return strings.Join(ss, ","), nil
	}

	val, err = s.repo.GetSet(key, fn)

	if err != nil {
		s.log.Error("SetCopiedMessageId", "err", err)
		return fmt.Errorf("SetCopiedMessageId: %w", err)
	}

	s.log.Debug("SetCopiedMessageId",
		"fromChatMessageId", fromChatMessageId,
		"toChatMessageId", toChatMessageId,
		"val", val,
	)
	return nil
}

// GetCopiedMessageIds получает идентификаторы скопированных сообщений по Id оригинала
func (s *Service) GetCopiedMessageIds(fromChatMessageId string) ([]string, error) {
	key := fmt.Sprintf("%s:%s", copiedMessageIdsPrefix, fromChatMessageId)

	val, err := s.repo.Get(key)
	if err != nil {
		return nil, fmt.Errorf("GetCopiedMessageIds: %w", err)
	}

	toChatMessageIds := []string{}
	if val != "" {
		// workaround https://stackoverflow.com/questions/28330908/how-to-string-split-an-empty-string-in-go
		toChatMessageIds = strings.Split(val, ",")
	}

	s.log.Debug("GetCopiedMessageIds",
		"fromChatMessageId", fromChatMessageId,
		"toChatMessageIds", toChatMessageIds,
	)

	return toChatMessageIds, nil
}

// DeleteCopiedMessageIds удаляет связь между оригинальным и скопированными сообщениями
func (s *Service) DeleteCopiedMessageIds(fromChatMessageId string) error {
	key := fmt.Sprintf("%s:%s", copiedMessageIdsPrefix, fromChatMessageId)

	err := s.repo.Delete(key)
	if err != nil {
		s.log.Error("DeleteCopiedMessageIds", "err", err)
		return fmt.Errorf("DeleteCopiedMessageIds: %w", err)
	}

	s.log.Debug("DeleteCopiedMessageIds",
		"fromChatMessageId", fromChatMessageId,
	)
	return nil
}

// SetNewMessageId сохраняет соответствие между временным и постоянным Id сообщения
func (s *Service) SetNewMessageId(chatId, tmpMessageId, newMessageId int64) error {
	key := fmt.Sprintf("%s:%d:%d", newMessageIdPrefix, chatId, tmpMessageId)

	err := s.repo.Set(key, fmt.Sprintf("%d", newMessageId))
	if err != nil {
		s.log.Error("SetNewMessageId", "err", err)
		return fmt.Errorf("SetNewMessageId: %w", err)
	}

	s.log.Debug("SetNewMessageId",
		"chatId", chatId,
		"tmpMessageId", tmpMessageId,
		"newMessageId", newMessageId,
	)
	return nil
}

// GetNewMessageId получает постоянный Id сообщения по временному
func (s *Service) GetNewMessageId(chatId, tmpMessageId int64) (int64, error) {
	key := fmt.Sprintf("%s:%d:%d", newMessageIdPrefix, chatId, tmpMessageId)

	val, err := s.repo.Get(key)
	if err != nil {
		s.log.Error("GetNewMessageId", "err", err)
		return 0, fmt.Errorf("GetNewMessageId: %w", err)
	}

	newMessageId := util.ConvertToInt[int64](val)

	s.log.Debug("GetNewMessageId",
		"chatId", chatId,
		"tmpMessageId", tmpMessageId,
		"newMessageId", newMessageId,
	)
	return newMessageId, nil
}

// DeleteNewMessageId удаляет соответствие между временным и постоянным Id сообщения
func (s *Service) DeleteNewMessageId(chatId, tmpMessageId int64) error {
	key := fmt.Sprintf("%s:%d:%d", newMessageIdPrefix, chatId, tmpMessageId)

	err := s.repo.Delete(key)
	if err != nil {
		s.log.Error("DeleteNewMessageId", "err", err)
		return fmt.Errorf("DeleteNewMessageId: %w", err)
	}

	s.log.Debug("DeleteNewMessageId",
		"chatId", chatId,
		"tmpMessageId", tmpMessageId,
	)
	return nil
}

// SetTmpMessageId сохраняет соответствие между постоянным и временным Id сообщения
func (s *Service) SetTmpMessageId(chatId, newMessageId, tmpMessageId int64) error {
	key := fmt.Sprintf("%s:%d:%d", tmpMessageIdPrefix, chatId, newMessageId)

	err := s.repo.Set(key, fmt.Sprintf("%d", tmpMessageId))
	if err != nil {
		s.log.Error("SetTmpMessageId", "err", err)
		return fmt.Errorf("SetTmpMessageId: %w", err)
	}

	s.log.Debug("SetTmpMessageId",
		"chatId", chatId,
		"newMessageId", newMessageId,
		"tmpMessageId", tmpMessageId,
	)
	return nil
}

// GetTmpMessageId получает временный Id сообщения по постоянному
func (s *Service) GetTmpMessageId(chatId, newMessageId int64) (int64, error) {
	key := fmt.Sprintf("%s:%d:%d", tmpMessageIdPrefix, chatId, newMessageId)

	val, err := s.repo.Get(key)
	if err != nil {
		s.log.Error("GetTmpMessageId", "err", err)
		return 0, fmt.Errorf("GetTmpMessageId: %w", err)
	}

	tmpMessageId := util.ConvertToInt[int64](val)

	s.log.Debug("GetTmpMessageId",
		"chatId", chatId,
		"newMessageId", newMessageId,
		"tmpMessageId", tmpMessageId,
	)
	return tmpMessageId, nil
}

// DeleteTmpMessageId удаляет соответствие между постоянным и временным Id сообщения
func (s *Service) DeleteTmpMessageId(chatId, newMessageId int64) error {
	key := fmt.Sprintf("%s:%d:%d", tmpMessageIdPrefix, chatId, newMessageId)

	err := s.repo.Delete(key)
	if err != nil {
		s.log.Error("DeleteTmpMessageId", "err", err)
		return fmt.Errorf("DeleteTmpMessageId: %w", err)
	}

	s.log.Debug("DeleteTmpMessageId",
		"chatId", chatId,
		"newMessageId", newMessageId,
	)
	return nil
}

// IncrementViewedMessages увеличивает счетчик просмотренных сообщений
func (s *Service) IncrementViewedMessages(toChatId int64) error {
	date := time.Now().UTC().Format("2006-01-02")
	key := fmt.Sprintf("%s:%d:%s", viewedMessagesPrefix, toChatId, date)
	var (
		err error
		val string
	)

	val, err = s.repo.Increment(key)
	if err != nil {
		s.log.Error("IncrementViewedMessages", "err", err)
		return fmt.Errorf("IncrementViewedMessages: %w", err)
	}

	s.log.Debug("IncrementViewedMessages",
		"toChatId", toChatId,
		"val", val,
	)
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
	key := fmt.Sprintf("%s:%d", forwardedMessagesPrefix, toChatId)

	val, err := s.repo.Get(key)
	if err != nil {
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
