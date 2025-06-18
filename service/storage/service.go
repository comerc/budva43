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
func (s *Service) SetCopiedMessageId(fromChatMessageId string, toChatMessageId string) {
	var (
		err error
		val string
	)
	defer func() {
		s.log.ErrorOrDebug(&err, "SetCopiedMessageId",
			"fromChatMessageId", fromChatMessageId,
			"toChatMessageId", toChatMessageId,
			"val", val,
		)
	}()

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

	key := fmt.Sprintf("%s:%s", copiedMessageIdsPrefix, fromChatMessageId)
	val, err = s.repo.GetSet(key, fn)
}

// GetCopiedMessageIds получает идентификаторы скопированных сообщений по Id оригинала
// TODO: входные параметры: сhatId, messageId (по аналогии с остальными методами)
func (s *Service) GetCopiedMessageIds(fromChatMessageId string) []string {
	var (
		err error
		val string
	)
	toChatMessageIds := []string{}
	defer s.log.ErrorOrDebug(&err, "GetCopiedMessageIds",
		"fromChatMessageId", fromChatMessageId,
		"toChatMessageIds", toChatMessageIds,
	)

	key := fmt.Sprintf("%s:%s", copiedMessageIdsPrefix, fromChatMessageId)
	val, err = s.repo.Get(key)
	if err != nil {
		return nil
	}

	if val != "" {
		// workaround https://stackoverflow.com/questions/28330908/how-to-string-split-an-empty-string-in-go
		toChatMessageIds = strings.Split(val, ",")
	}

	return toChatMessageIds
}

// DeleteCopiedMessageIds удаляет связь между оригинальным и скопированными сообщениями
// TODO: почему не используется?
func (s *Service) DeleteCopiedMessageIds(fromChatMessageId string) {
	var err error
	defer s.log.ErrorOrDebug(&err, "DeleteCopiedMessageIds",
		"fromChatMessageId", fromChatMessageId,
	)

	key := fmt.Sprintf("%s:%s", copiedMessageIdsPrefix, fromChatMessageId)
	err = s.repo.Delete(key)
}

// SetNewMessageId сохраняет соответствие между временным и постоянным Id сообщения
func (s *Service) SetNewMessageId(chatId, tmpMessageId, newMessageId int64) {
	var err error
	defer s.log.ErrorOrDebug(&err, "SetNewMessageId",
		"chatId", chatId,
		"tmpMessageId", tmpMessageId,
		"newMessageId", newMessageId,
	)

	key := fmt.Sprintf("%s:%d:%d", newMessageIdPrefix, chatId, tmpMessageId)
	err = s.repo.Set(key, fmt.Sprintf("%d", newMessageId))
}

// GetNewMessageId получает постоянный Id сообщения по временному
// TODO: почему не используется?
func (s *Service) GetNewMessageId(chatId, tmpMessageId int64) int64 {
	var (
		err error
		val string
	)
	newMessageId := int64(0)
	defer s.log.ErrorOrDebug(&err, "GetNewMessageId",
		"chatId", chatId,
		"tmpMessageId", tmpMessageId,
		"newMessageId", newMessageId,
	)

	key := fmt.Sprintf("%s:%d:%d", newMessageIdPrefix, chatId, tmpMessageId)
	val, err = s.repo.Get(key)
	if err != nil {
		return 0
	}

	newMessageId = util.ConvertToInt[int64](val)
	return newMessageId
}

// DeleteNewMessageId удаляет соответствие между временным и постоянным Id сообщения
func (s *Service) DeleteNewMessageId(chatId, tmpMessageId int64) {
	var err error
	defer s.log.ErrorOrDebug(&err, "DeleteNewMessageId",
		"chatId", chatId,
		"tmpMessageId", tmpMessageId,
	)

	key := fmt.Sprintf("%s:%d:%d", newMessageIdPrefix, chatId, tmpMessageId)
	err = s.repo.Delete(key)
}

// SetTmpMessageId сохраняет соответствие между постоянным и временным Id сообщения
func (s *Service) SetTmpMessageId(chatId, newMessageId, tmpMessageId int64) {
	var err error
	defer s.log.ErrorOrDebug(&err, "SetTmpMessageId",
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
		err error
		val string
	)
	tmpMessageId := int64(0)
	defer s.log.ErrorOrDebug(&err, "GetTmpMessageId",
		"chatId", chatId,
		"newMessageId", newMessageId,
		"tmpMessageId", tmpMessageId,
	)

	key := fmt.Sprintf("%s:%d:%d", tmpMessageIdPrefix, chatId, newMessageId)
	val, err = s.repo.Get(key)
	if err != nil {
		return 0
	}

	tmpMessageId = util.ConvertToInt[int64](val)
	return tmpMessageId
}

// DeleteTmpMessageId удаляет соответствие между постоянным и временным Id сообщения
func (s *Service) DeleteTmpMessageId(chatId, newMessageId int64) {
	var err error
	defer s.log.ErrorOrDebug(&err, "DeleteTmpMessageId",
		"chatId", chatId,
		"newMessageId", newMessageId,
	)

	key := fmt.Sprintf("%s:%d:%d", tmpMessageIdPrefix, chatId, newMessageId)
	err = s.repo.Delete(key)
}

// IncrementViewedMessages увеличивает счетчик просмотренных сообщений
func (s *Service) IncrementViewedMessages(toChatId int64, date string) {
	var (
		err error
		val uint64
	)
	defer func() {
		s.log.ErrorOrDebug(&err, "IncrementViewedMessages",
			"toChatId", toChatId,
			"date", date,
			"val", val,
		)
	}()

	if date == "" { // внешняя date нужна для тестирования
		date = util.GetCurrentDate()
	}
	key := fmt.Sprintf("%s:%d:%s", viewedMessagesPrefix, toChatId, date)
	val, err = s.repo.Increment(key)
}

// GetViewedMessages получает количество просмотренных сообщений
func (s *Service) GetViewedMessages(toChatId int64, date string) int64 {
	var (
		err error
		val string
	)
	viewed := int64(0)
	defer s.log.ErrorOrDebug(&err, "GetViewedMessages",
		"toChatId", toChatId,
		"date", date,
		"viewed", viewed,
	)

	key := fmt.Sprintf("%s:%d:%s", viewedMessagesPrefix, toChatId, date)
	val, err = s.repo.Get(key)
	if err != nil || val == "" {
		return 0
	}

	viewed = util.ConvertToInt[int64](val)
	return viewed
}

// IncrementForwardedMessages увеличивает счетчик пересланных сообщений
func (s *Service) IncrementForwardedMessages(toChatId int64, date string) {
	var (
		err error
		val uint64
	)
	defer func() {
		s.log.ErrorOrDebug(&err, "IncrementForwardedMessages",
			"toChatId", toChatId,
			"date", date,
			"val", val,
		)
	}()

	if date == "" { // внешняя date нужна для тестирования
		date = util.GetCurrentDate()
	}
	key := fmt.Sprintf("%s:%d:%s", forwardedMessagesPrefix, toChatId, date)
	val, err = s.repo.Increment(key)
}

// GetForwardedMessages получает количество пересланных сообщений
func (s *Service) GetForwardedMessages(toChatId int64, date string) int64 {
	var (
		err error
		val string
	)
	forwarded := int64(0)
	defer s.log.ErrorOrDebug(&err, "GetForwardedMessages",
		"toChatId", toChatId,
		"date", date,
		"forwarded", forwarded,
	)

	if date == "" { // внешняя date нужна для тестирования
		date = util.GetCurrentDate()
	}
	key := fmt.Sprintf("%s:%d:%s", forwardedMessagesPrefix, toChatId, date)
	val, err = s.repo.Get(key)
	if err != nil || val == "" {
		return 0
	}

	forwarded = util.ConvertToInt[int64](val)
	return forwarded
}

// SetAnswerMessageId устанавливает идентификатор сообщения ответа
func (s *Service) SetAnswerMessageId(dstChatId, tmpMessageId int64, fromChatMessageId string) {
	var err error
	defer s.log.ErrorOrDebug(&err, "SetAnswerMessageId",
		"dstChatId", dstChatId,
		"tmpMessageId", tmpMessageId,
		"fromChatMessageId", fromChatMessageId,
	)

	key := fmt.Sprintf("%s:%d:%d", answerMessageIdPrefix, dstChatId, tmpMessageId)
	err = s.repo.Set(key, fromChatMessageId)
}

// GetAnswerMessageId возвращает идентификатор сообщения ответа
func (s *Service) GetAnswerMessageId(dstChatId, tmpMessageId int64) string {
	var (
		err error
		val string
	)
	fromChatMessageId := ""
	defer s.log.ErrorOrDebug(&err, "GetAnswerMessageId",
		"dstChatId", dstChatId,
		"tmpMessageId", tmpMessageId,
		"fromChatMessageId", fromChatMessageId,
	)

	key := fmt.Sprintf("%s:%d:%d", answerMessageIdPrefix, dstChatId, tmpMessageId)
	val, err = s.repo.Get(key)
	if err != nil {
		return ""
	}

	fromChatMessageId = val
	return fromChatMessageId
}

// DeleteAnswerMessageId удаляет идентификатор сообщения ответа
func (s *Service) DeleteAnswerMessageId(dstChatId, tmpMessageId int64) {
	var err error
	defer s.log.ErrorOrDebug(&err, "DeleteAnswerMessageId",
		"dstChatId", dstChatId,
		"tmpMessageId", tmpMessageId,
	)

	key := fmt.Sprintf("%s:%d:%d", answerMessageIdPrefix, dstChatId, tmpMessageId)
	err = s.repo.Delete(key)
}
