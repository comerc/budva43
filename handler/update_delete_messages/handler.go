package update_delete_messages

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/zelenin/go-tdlib/client"

	"github.com/comerc/budva43/config"
	"github.com/comerc/budva43/util"
)

type telegramRepo interface {
	GetClient() *client.Client
}

type queueRepo interface {
	Add(task func())
}

type storageService interface {
	GetCopiedMessageIds(fromChatMessageId string) ([]string, error)
	DeleteCopiedMessageIds(fromChatMessageId string) error
	GetNewMessageId(chatId, tmpMessageId int64) (int64, error)
	DeleteNewMessageId(chatId, tmpMessageId int64) error
	DeleteTmpMessageId(chatId, newMessageId int64) error
	DeleteAnswerMessageId(dstChatId, tmpMessageId int64) error
}

type Handler struct {
	log *slog.Logger
	//
	telegramRepo   telegramRepo
	queueRepo      queueRepo
	storageService storageService
}

func New(
	telegramRepo telegramRepo,
	queueRepo queueRepo,
	storageService storageService,
) *Handler {
	return &Handler{
		log: slog.With("module", "handler.update_delete_messages"),
		//
		telegramRepo:   telegramRepo,
		queueRepo:      queueRepo,
		storageService: storageService,
	}
}

func (s *Handler) Run(update *client.UpdateDeleteMessages) {
	if !update.IsPermanent {
		return
	}

	if _, ok := config.Engine.UniqueSources[update.ChatId]; !ok {
		return
	}

	s.process(update.ChatId, update.MessageIds)
}

func (s *Handler) process(chatId int64, messageIds []int64) {
	const maxRetries = 3
	retryCount := 0

	var fn func()
	fn = func() {
		data, err := s.collectData(chatId, messageIds)
		if err != nil {
			retryCount++
			if retryCount >= maxRetries {
				s.log.Error("max retries reached for message deletion",
					"chatId", chatId,
					"messageIds", messageIds,
				)
				return
			}
			s.log.Info("retrying message deletion",
				"retryCount", retryCount,
				"chatId", chatId,
				"messageIds", messageIds,
			)
			s.queueRepo.Add(fn) // переставляем в конец очереди
			return
		}

		result := s.deleteMessages(chatId, messageIds, data)
		s.log.Info("deleteMessages",
			"chatId", chatId,
			"messageIds", messageIds,
			"result", result,
		)
	}

	s.queueRepo.Add(fn)
}

type data struct {
	copiedMessageIds map[string][]string // fromChatMessageId -> []toChatMessageId
	newMessageIds    map[string]int64    // tmpChatMessageId -> newMessageId
}

func (s *Handler) collectData(chatId int64, messageIds []int64) (*data, error) {
	data := &data{
		copiedMessageIds: make(map[string][]string),
		newMessageIds:    make(map[string]int64),
	}

	for _, messageId := range messageIds {
		fromChatMessageId := fmt.Sprintf("%d:%d", chatId, messageId)
		toChatMessageIds, _ := s.storageService.GetCopiedMessageIds(fromChatMessageId)
		data.copiedMessageIds[fromChatMessageId] = toChatMessageIds

		for _, toChatMessageId := range toChatMessageIds {
			a := strings.Split(toChatMessageId, ":")
			// forwardRuleId := a[0]
			dstChatId := util.ConvertToInt[int64](a[1])
			tmpMessageId := util.ConvertToInt[int64](a[2])

			newMessageId, _ := s.storageService.GetNewMessageId(dstChatId, tmpMessageId)
			if newMessageId == 0 {
				return nil, fmt.Errorf("newMessageId не найден")
			}

			tmpChatMessageId := fmt.Sprintf("%d:%d", dstChatId, tmpMessageId)
			data.newMessageIds[tmpChatMessageId] = newMessageId
		}
	}

	return data, nil
}

func (s *Handler) deleteMessages(chatId int64, messageIds []int64, data *data) []string {
	result := make([]string, 0)

	for _, messageId := range messageIds {
		fromChatMessageId := fmt.Sprintf("%d:%d", chatId, messageId)
		toChatMessageIds := data.copiedMessageIds[fromChatMessageId]

		for _, toChatMessageId := range toChatMessageIds {
			a := strings.Split(toChatMessageId, ":")
			forwardRuleId := a[0]
			dstChatId := util.ConvertToInt[int64](a[1])
			tmpMessageId := util.ConvertToInt[int64](a[2])

			forwardRule, ok := config.Engine.ForwardRules[forwardRuleId]
			if !ok {
				s.log.Error("forwardRule not found",
					"forwardRuleId", forwardRuleId,
					"fromChatMessageId", fromChatMessageId,
					"toChatMessageId", toChatMessageId,
				)
				continue
			}
			if !forwardRule.Indelible {
				continue
			}

			_ = s.storageService.DeleteAnswerMessageId(dstChatId, tmpMessageId)

			tmpChatMessageId := fmt.Sprintf("%d:%d", dstChatId, tmpMessageId)
			newMessageId := data.newMessageIds[tmpChatMessageId]

			// TODO: может лучше удалять индексы после удаления сообщения?
			_ = s.storageService.DeleteTmpMessageId(dstChatId, newMessageId)
			_ = s.storageService.DeleteNewMessageId(dstChatId, tmpMessageId)

			_, err := s.telegramRepo.GetClient().DeleteMessages(&client.DeleteMessagesRequest{
				ChatId:     dstChatId,
				MessageIds: []int64{newMessageId},
				Revoke:     true,
			})
			if err != nil {
				s.log.Error("DeleteMessages", "err", err)
				continue
			}

			result = append(result,
				fmt.Sprintf("%d:%d:%d", dstChatId, tmpMessageId, newMessageId))
		}

		if len(toChatMessageIds) > 0 {
			_ = s.storageService.DeleteCopiedMessageIds(fromChatMessageId)
		}
	}

	return result
}
