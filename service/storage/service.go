package engine_storage

import (
	"fmt"
	"strings"

	"github.com/comerc/budva43/app/log"
	"github.com/comerc/budva43/app/util"
)

const (
	// Префиксы ключей для хранения в BadgerDB
	copiedMessageIdsPrefix  = "copiedMsgIds"
	newMessageIdPrefix      = "newMsgId"
	tmpMessageIdPrefix      = "tmpMsgId"
	viewedMessagesPrefix    = "viewedMsgs"
	forwardedMessagesPrefix = "forwardedMsgs"
	answerMessageIdPrefix   = "answerMsgId"
)

//go:generate mockery --name=storageRepo --exported
type storageRepo interface {
	GetSet(key string, fn func(val string) (string, error)) (string, error)
	Set(key, value string) error
	Get(key string) (string, error)
	Delete(key string) error
	Increment(key string) (uint64, error)
}

// Service предоставляет методы для хранения данных, специфичных для engine
type Service struct {
	log *log.Logger
	//
	repo storageRepo
}

// New создает новый экземпляр сервиса хранения данных
func New(repo storageRepo) *Service {
	return &Service{
		log: log.NewLogger("service.storage"),
		//
		repo: repo,
	}
}

// SetCopiedMessageId сохраняет связь между оригинальным и скопированным сообщением
func (s *Service) SetCopiedMessageId(chatId, messageId int64, toChatMessageId string) {
	var (
		err    error
		val    string
		result []string
	)
	defer s.log.ErrorOrDebug(&err, "",
		"chatId", chatId,
		"messageId", messageId,
		"toChatMessageId", toChatMessageId,
		"result", &result,
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

	key := fmt.Sprintf("%s:%d:%d", copiedMessageIdsPrefix, chatId, messageId)
	val, err = s.repo.GetSet(key, fn)
	result = strings.Split(val, ",")
}

// GetCopiedMessageIds получает идентификаторы скопированных сообщений по Id оригинала
func (s *Service) GetCopiedMessageIds(chatId, messageId int64) []string {
	var (
		err    error
		val    string
		result []string
	)
	defer s.log.ErrorOrDebug(&err, "",
		"chatId", chatId,
		"messageId", messageId,
		"result", &result,
	)

	key := fmt.Sprintf("%s:%d:%d", copiedMessageIdsPrefix, chatId, messageId)
	val, err = s.repo.Get(key)
	if err != nil {
		return nil
	}

	if val != "" {
		// workaround https://stackoverflow.com/questions/28330908/how-to-string-split-an-empty-string-in-go
		result = strings.Split(val, ",")
	}

	return result
}

// DeleteCopiedMessageIds удаляет связь между оригинальным и скопированными сообщениями
func (s *Service) DeleteCopiedMessageIds(chatId, messageId int64) {
	var err error
	defer s.log.ErrorOrDebug(&err, "",
		"chatId", chatId,
		"messageId", messageId,
	)

	key := fmt.Sprintf("%s:%d:%d", copiedMessageIdsPrefix, chatId, messageId)
	err = s.repo.Delete(key)
}

// SetNewMessageId сохраняет соответствие между временным и постоянным Id сообщения
func (s *Service) SetNewMessageId(chatId, tmpMessageId, newMessageId int64) {
	var err error
	defer s.log.ErrorOrDebug(&err, "",
		"chatId", chatId,
		"tmpMessageId", tmpMessageId,
		"newMessageId", newMessageId,
	)

	key := fmt.Sprintf("%s:%d:%d", newMessageIdPrefix, chatId, tmpMessageId)
	err = s.repo.Set(key, fmt.Sprintf("%d", newMessageId))
}

// GetNewMessageId получает постоянный Id сообщения по временному
func (s *Service) GetNewMessageId(chatId, tmpMessageId int64) int64 {
	var (
		err    error
		val    string
		result int64
	)
	defer s.log.ErrorOrDebug(&err, "",
		"chatId", chatId,
		"tmpMessageId", tmpMessageId,
		"result", &result,
	)

	key := fmt.Sprintf("%s:%d:%d", newMessageIdPrefix, chatId, tmpMessageId)
	val, err = s.repo.Get(key)
	if err != nil {
		return 0
	}

	result = util.ConvertToInt[int64](val)
	return result
}

// DeleteNewMessageId удаляет соответствие между временным и постоянным Id сообщения
func (s *Service) DeleteNewMessageId(chatId, tmpMessageId int64) {
	var err error
	defer s.log.ErrorOrDebug(&err, "",
		"chatId", chatId,
		"tmpMessageId", tmpMessageId,
	)

	key := fmt.Sprintf("%s:%d:%d", newMessageIdPrefix, chatId, tmpMessageId)
	err = s.repo.Delete(key)
}

// SetTmpMessageId сохраняет соответствие между постоянным и временным Id сообщения
func (s *Service) SetTmpMessageId(chatId, newMessageId, tmpMessageId int64) {
	var err error
	defer s.log.ErrorOrDebug(&err, "",
		"chatId", chatId,
		"newMessageId", newMessageId,
		"tmpMessageId", tmpMessageId,
	)

	key := fmt.Sprintf("%s:%d:%d", tmpMessageIdPrefix, chatId, newMessageId)
	err = s.repo.Set(key, fmt.Sprintf("%d", tmpMessageId))
}

// GetTmpMessageId получает временный Id сообщения по постоянному
func (s *Service) GetTmpMessageId(chatId, newMessageId int64) int64 {
	var (
		err    error
		val    string
		result int64
	)
	defer s.log.ErrorOrDebug(&err, "",
		"chatId", chatId,
		"newMessageId", newMessageId,
		"result", &result,
	)

	key := fmt.Sprintf("%s:%d:%d", tmpMessageIdPrefix, chatId, newMessageId)
	val, err = s.repo.Get(key)
	if err != nil {
		return 0
	}

	result = util.ConvertToInt[int64](val)
	return result
}

// DeleteTmpMessageId удаляет соответствие между постоянным и временным Id сообщения
func (s *Service) DeleteTmpMessageId(chatId, newMessageId int64) {
	var err error
	defer s.log.ErrorOrDebug(&err, "",
		"chatId", chatId,
		"newMessageId", newMessageId,
	)

	key := fmt.Sprintf("%s:%d:%d", tmpMessageIdPrefix, chatId, newMessageId)
	err = s.repo.Delete(key)
}

// IncrementViewedMessages увеличивает счетчик просмотренных сообщений
func (s *Service) IncrementViewedMessages(toChatId int64, date string) {
	var (
		err    error
		result uint64
	)
	defer s.log.ErrorOrDebug(&err, "",
		"toChatId", toChatId,
		"date", date,
		"result", &result,
	)

	if date == "" { // внешняя date нужна для тестирования
		date = util.GetCurrentDate()
	}
	key := fmt.Sprintf("%s:%d:%s", viewedMessagesPrefix, toChatId, date)
	result, err = s.repo.Increment(key)
}

// GetViewedMessages получает количество просмотренных сообщений
func (s *Service) GetViewedMessages(toChatId int64, date string) int64 {
	var (
		err    error
		val    string
		result int64
	)
	defer s.log.ErrorOrDebug(&err, "",
		"toChatId", toChatId,
		"date", date,
		"result", &result,
	)

	key := fmt.Sprintf("%s:%d:%s", viewedMessagesPrefix, toChatId, date)
	val, err = s.repo.Get(key)
	if err != nil || val == "" {
		return 0
	}

	result = util.ConvertToInt[int64](val)
	return result
}

// IncrementForwardedMessages увеличивает счетчик пересланных сообщений
func (s *Service) IncrementForwardedMessages(toChatId int64, date string) {
	var (
		err    error
		result uint64
	)
	defer s.log.ErrorOrDebug(&err, "",
		"toChatId", toChatId,
		"date", date,
		"result", &result,
	)

	if date == "" { // внешняя date нужна для тестирования
		date = util.GetCurrentDate()
	}
	key := fmt.Sprintf("%s:%d:%s", forwardedMessagesPrefix, toChatId, date)
	result, err = s.repo.Increment(key)
}

// GetForwardedMessages получает количество пересланных сообщений
func (s *Service) GetForwardedMessages(toChatId int64, date string) int64 {
	var (
		err    error
		val    string
		result int64
	)
	defer s.log.ErrorOrDebug(&err, "",
		"toChatId", toChatId,
		"date", date,
		"result", &result,
	)

	if date == "" { // внешняя date нужна для тестирования
		date = util.GetCurrentDate()
	}
	key := fmt.Sprintf("%s:%d:%s", forwardedMessagesPrefix, toChatId, date)
	val, err = s.repo.Get(key)
	if err != nil || val == "" {
		return 0
	}

	result = util.ConvertToInt[int64](val)
	return result
}

// SetAnswerMessageId устанавливает идентификатор сообщения ответа
func (s *Service) SetAnswerMessageId(dstChatId, tmpMessageId, chatId, messageId int64) {
	var err error
	defer s.log.ErrorOrDebug(&err, "",
		"dstChatId", dstChatId,
		"tmpMessageId", tmpMessageId,
		"chatId", chatId,
		"messageId", messageId,
	)

	fromChatMessageId := fmt.Sprintf("%d:%d", chatId, messageId)
	key := fmt.Sprintf("%s:%d:%d", answerMessageIdPrefix, dstChatId, tmpMessageId)
	err = s.repo.Set(key, fromChatMessageId)
}

// GetAnswerMessageId возвращает идентификатор сообщения ответа
func (s *Service) GetAnswerMessageId(dstChatId, tmpMessageId int64) string {
	var (
		err    error
		result string
	)
	defer s.log.ErrorOrDebug(&err, "",
		"dstChatId", dstChatId,
		"tmpMessageId", tmpMessageId,
		"result", &result,
	)

	key := fmt.Sprintf("%s:%d:%d", answerMessageIdPrefix, dstChatId, tmpMessageId)
	result, err = s.repo.Get(key)
	if err != nil {
		return ""
	}

	return result // fromChatMessageId
}

// DeleteAnswerMessageId удаляет идентификатор сообщения ответа
func (s *Service) DeleteAnswerMessageId(dstChatId, tmpMessageId int64) {
	var err error
	defer s.log.ErrorOrDebug(&err, "",
		"dstChatId", dstChatId,
		"tmpMessageId", tmpMessageId,
	)

	key := fmt.Sprintf("%s:%d:%d", answerMessageIdPrefix, dstChatId, tmpMessageId)
	err = s.repo.Delete(key)
}
