package engine_storage

import (
	"fmt"
	"strings"

	"github.com/comerc/budva43/util"
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
	Increment(key string) (string, error)
}

// Service предоставляет методы для хранения данных, специфичных для engine
type Service struct {
	log *util.Logger
	//
	repo storageRepo
}

// New создает новый экземпляр сервиса хранения данных
func New(repo storageRepo) *Service {
	return &Service{
		log: util.NewLogger("service.storage"),
		//
		repo: repo,
	}
}

// SetCopiedMessageId сохраняет связь между оригинальным и скопированным сообщением
func (s *Service) SetCopiedMessageId(fromChatMessageId string, toChatMessageId string) {
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
	val, err := s.repo.GetSet(key, fn)
	if err != nil {
		// s.log.Error("SetCopiedMessageId", "err", err)
	}

	_ = val // TODO: костыль
	// s.log.Debug("SetCopiedMessageId",
	// 	"fromChatMessageId", fromChatMessageId,
	// 	"toChatMessageId", toChatMessageId,
	// 	"val", val,
	// )
}

// GetCopiedMessageIds получает идентификаторы скопированных сообщений по Id оригинала
// TODO: входные параметры: сhatId, messageId (по аналогии с остальными методами)
func (s *Service) GetCopiedMessageIds(fromChatMessageId string) []string {
	key := fmt.Sprintf("%s:%s", copiedMessageIdsPrefix, fromChatMessageId)
	val, err := s.repo.Get(key)
	if err != nil {
		return nil
	}

	toChatMessageIds := []string{}
	if val != "" {
		// workaround https://stackoverflow.com/questions/28330908/how-to-string-split-an-empty-string-in-go
		toChatMessageIds = strings.Split(val, ",")
	}

	// s.log.Debug("GetCopiedMessageIds",
	// 	"fromChatMessageId", fromChatMessageId,
	// 	"toChatMessageIds", toChatMessageIds,
	// )

	return toChatMessageIds
}

// DeleteCopiedMessageIds удаляет связь между оригинальным и скопированными сообщениями
// TODO: почему не используется?
func (s *Service) DeleteCopiedMessageIds(fromChatMessageId string) {
	key := fmt.Sprintf("%s:%s", copiedMessageIdsPrefix, fromChatMessageId)
	err := s.repo.Delete(key)
	if err != nil {
		// s.log.Error("DeleteCopiedMessageIds", "err", err)
		return
	}

	// s.log.Debug("DeleteCopiedMessageIds",
	// 	"fromChatMessageId", fromChatMessageId,
	// )
}

// SetNewMessageId сохраняет соответствие между временным и постоянным Id сообщения
func (s *Service) SetNewMessageId(chatId, tmpMessageId, newMessageId int64) {
	key := fmt.Sprintf("%s:%d:%d", newMessageIdPrefix, chatId, tmpMessageId)
	err := s.repo.Set(key, fmt.Sprintf("%d", newMessageId))
	if err != nil {
		// s.log.Error("SetNewMessageId", "err", err)
	}

	// s.log.Debug("SetNewMessageId",
	// 	"chatId", chatId,
	// 	"tmpMessageId", tmpMessageId,
	// 	"newMessageId", newMessageId,
	// )
}

// GetNewMessageId получает постоянный Id сообщения по временному
// TODO: почему не используется?
func (s *Service) GetNewMessageId(chatId, tmpMessageId int64) int64 {
	key := fmt.Sprintf("%s:%d:%d", newMessageIdPrefix, chatId, tmpMessageId)
	val, err := s.repo.Get(key)
	if err != nil {
		// s.log.Error("GetNewMessageId", "err", err)
		return 0
	}

	newMessageId := util.ConvertToInt[int64](val)

	// s.log.Debug("GetNewMessageId",
	// 	"chatId", chatId,
	// 	"tmpMessageId", tmpMessageId,
	// 	"newMessageId", newMessageId,
	// )
	return newMessageId
}

// DeleteNewMessageId удаляет соответствие между временным и постоянным Id сообщения
func (s *Service) DeleteNewMessageId(chatId, tmpMessageId int64) {
	key := fmt.Sprintf("%s:%d:%d", newMessageIdPrefix, chatId, tmpMessageId)
	err := s.repo.Delete(key)
	if err != nil {
		// s.log.Error("DeleteNewMessageId", "err", err)
		return
	}

	// s.log.Debug("DeleteNewMessageId",
	// 	"chatId", chatId,
	// 	"tmpMessageId", tmpMessageId,
	// )
}

// SetTmpMessageId сохраняет соответствие между постоянным и временным Id сообщения
func (s *Service) SetTmpMessageId(chatId, newMessageId, tmpMessageId int64) {
	key := fmt.Sprintf("%s:%d:%d", tmpMessageIdPrefix, chatId, newMessageId)
	err := s.repo.Set(key, fmt.Sprintf("%d", tmpMessageId))
	if err != nil {
		// s.log.Error("SetTmpMessageId", "err", err)
		return
	}

	// s.log.Debug("SetTmpMessageId",
	// 	"chatId", chatId,
	// 	"newMessageId", newMessageId,
	// 	"tmpMessageId", tmpMessageId,
	// )
}

// GetTmpMessageId получает временный Id сообщения по постоянному
func (s *Service) GetTmpMessageId(chatId, newMessageId int64) int64 {
	key := fmt.Sprintf("%s:%d:%d", tmpMessageIdPrefix, chatId, newMessageId)
	val, err := s.repo.Get(key)
	if err != nil {
		// s.log.Error("GetTmpMessageId", "err", err)
		return 0
	}

	tmpMessageId := util.ConvertToInt[int64](val)

	// s.log.Debug("GetTmpMessageId",
	// 	"chatId", chatId,
	// 	"newMessageId", newMessageId,
	// 	"tmpMessageId", tmpMessageId,
	// )
	return tmpMessageId
}

// DeleteTmpMessageId удаляет соответствие между постоянным и временным Id сообщения
func (s *Service) DeleteTmpMessageId(chatId, newMessageId int64) {
	key := fmt.Sprintf("%s:%d:%d", tmpMessageIdPrefix, chatId, newMessageId)
	err := s.repo.Delete(key)
	if err != nil {
		// s.log.Error("DeleteTmpMessageId", "err", err)
		return
	}

	// s.log.Debug("DeleteTmpMessageId",
	// 	"chatId", chatId,
	// 	"newMessageId", newMessageId,
	// )
}

// IncrementViewedMessages увеличивает счетчик просмотренных сообщений
func (s *Service) IncrementViewedMessages(toChatId int64, date string) {
	if date == "" { // внешняя date нужна для тестирования
		date = util.GetCurrentDate()
	}
	key := fmt.Sprintf("%s:%d:%s", viewedMessagesPrefix, toChatId, date)
	val, err := s.repo.Increment(key)
	if err != nil {
		// s.log.Error("IncrementViewedMessages", "err", err)
		return
	}

	_ = val // TODO: костыль
	// s.log.Debug("IncrementViewedMessages",
	// 	"toChatId", toChatId,
	// 	"date", date,
	// 	"val", val,
	// )
}

// GetViewedMessages получает количество просмотренных сообщений
func (s *Service) GetViewedMessages(toChatId int64, date string) int64 {
	key := fmt.Sprintf("%s:%d:%s", viewedMessagesPrefix, toChatId, date)
	val, err := s.repo.Get(key)
	if err != nil {
		return 0
	}

	var viewed int64
	if val == "" {
		viewed = 0
	} else {
		viewed = util.ConvertToInt[int64](val)
	}

	// s.log.Debug("GetViewedMessages",
	// 	"toChatId", toChatId,
	// 	"date", date,
	// 	"viewed", viewed,
	// )
	return viewed
}

// IncrementForwardedMessages увеличивает счетчик пересланных сообщений
func (s *Service) IncrementForwardedMessages(toChatId int64, date string) {
	if date == "" { // внешняя date нужна для тестирования
		date = util.GetCurrentDate()
	}
	key := fmt.Sprintf("%s:%d:%s", forwardedMessagesPrefix, toChatId, date)
	val, err := s.repo.Increment(key)
	if err != nil {
		// s.log.Error("IncrementForwardedMessages", "err", err)
		return
	}

	_ = val // TODO: костыль
	// s.log.Debug("IncrementForwardedMessages",
	// 	"toChatId", toChatId,
	// 	"date", date,
	// 	"val", val,
	// )
}

// GetForwardedMessages получает количество пересланных сообщений
func (s *Service) GetForwardedMessages(toChatId int64, date string) int64 {
	if date == "" { // внешняя date нужна для тестирования
		date = util.GetCurrentDate()
	}
	key := fmt.Sprintf("%s:%d:%s", forwardedMessagesPrefix, toChatId, date)
	val, err := s.repo.Get(key)
	if err != nil {
		return 0
	}

	var forwarded int64
	if val == "" {
		forwarded = 0
	} else {
		forwarded = util.ConvertToInt[int64](val)
	}

	// s.log.Debug("GetForwardedMessages",
	// 	"toChatId", toChatId,
	// 	"date", date,
	// 	"forwarded", forwarded,
	// )
	return forwarded
}

// SetAnswerMessageId устанавливает идентификатор сообщения ответа
func (s *Service) SetAnswerMessageId(dstChatId, tmpMessageId int64, fromChatMessageId string) {
	key := fmt.Sprintf("%s:%d:%d", answerMessageIdPrefix, dstChatId, tmpMessageId)
	err := s.repo.Set(key, fromChatMessageId)
	if err != nil {
		// s.log.Error("SetAnswerMessageId", "err", err)
		return
	}

	// s.log.Debug("SetAnswerMessageId",
	// 	"dstChatId", dstChatId,
	// 	"tmpMessageId", tmpMessageId,
	// 	"fromChatMessageId", fromChatMessageId,
	// )
}

// GetAnswerMessageId возвращает идентификатор сообщения ответа
func (s *Service) GetAnswerMessageId(dstChatId, tmpMessageId int64) string {
	key := fmt.Sprintf("%s:%d:%d", answerMessageIdPrefix, dstChatId, tmpMessageId)
	val, err := s.repo.Get(key)
	if err != nil {
		// s.log.Error("GetAnswerMessageId", "err", err)
		return ""
	}

	// s.log.Debug("GetAnswerMessageId",
	// 	"dstChatId", dstChatId,
	// 	"tmpMessageId", tmpMessageId,
	// 	"fromChatMessageId", val,
	// )
	return val
}

// DeleteAnswerMessageId удаляет идентификатор сообщения ответа
func (s *Service) DeleteAnswerMessageId(dstChatId, tmpMessageId int64) {
	key := fmt.Sprintf("%s:%d:%d", answerMessageIdPrefix, dstChatId, tmpMessageId)
	err := s.repo.Delete(key)
	if err != nil {
		// s.log.Error("DeleteAnswerMessageId", "err", err)
		return
	}

	// s.log.Debug("DeleteAnswerMessageId",
	// 	"dstChatId", dstChatId,
	// 	"tmpMessageId", tmpMessageId,
	// )
}
