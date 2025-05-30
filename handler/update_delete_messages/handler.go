package update_delete_messages

import (
	"fmt"
	"strings"

	"github.com/zelenin/go-tdlib/client"

	"github.com/comerc/budva43/app/config"
	"github.com/comerc/budva43/app/log"
	"github.com/comerc/budva43/app/util"
)

type telegramRepo interface {
	GetClient() *client.Client
}

type queueRepo interface {
	Add(task func())
}

type storageService interface {
	GetCopiedMessageIds(fromChatMessageId string) []string
	DeleteCopiedMessageIds(fromChatMessageId string)
	GetNewMessageId(chatId, tmpMessageId int64) int64
	DeleteNewMessageId(chatId, tmpMessageId int64)
	DeleteTmpMessageId(chatId, newMessageId int64)
	DeleteAnswerMessageId(dstChatId, tmpMessageId int64)
}

type Handler struct {
	log *log.Logger
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
		log: log.NewLogger("handler.update_delete_messages"),
		//
		telegramRepo:   telegramRepo,
		queueRepo:      queueRepo,
		storageService: storageService,
	}
}

// Run выполняет обрабатку обновления об удалении сообщений
func (h *Handler) Run(update *client.UpdateDeleteMessages) {
	if !update.IsPermanent {
		return
	}

	chatId := update.ChatId
	if _, ok := config.Engine.UniqueSources[update.ChatId]; !ok {
		return
	}
	messageIds := update.MessageIds

	const maxRetries = 3
	retryCount := 0

	var fn func()
	fn = func() {
		data := h.collectData(chatId, messageIds)
		if data.needRepeat {
			retryCount++
			if retryCount >= maxRetries {
				// h.log.Error("max retries reached for message deletion",
				// 	"chatId", chatId,
				// 	"messageIds", messageIds,
				// )
				return
			}
			// h.log.Info("retrying message deletion",
			// 	"retryCount", retryCount,
			// 	"chatId", chatId,
			// 	"messageIds", messageIds,
			// )
			h.queueRepo.Add(fn) // переставляем в конец очереди
			return
		}

		h.deleteMessages(chatId, messageIds, data)
	}

	h.queueRepo.Add(fn)
}

type data struct {
	needRepeat       bool
	copiedMessageIds map[string][]string // fromChatMessageId -> []toChatMessageId
	newMessageIds    map[string]int64    // tmpChatMessageId -> newMessageId
}

// collectData собирает данные для удаления сообщений
func (h *Handler) collectData(chatId int64, messageIds []int64) *data {
	result := &data{
		copiedMessageIds: make(map[string][]string),
		newMessageIds:    make(map[string]int64),
	}
	// errs := []error{} // TODO: собирать все ошибки (для тестов)

	for _, messageId := range messageIds {
		fromChatMessageId := fmt.Sprintf("%d:%d", chatId, messageId)
		toChatMessageIds := h.storageService.GetCopiedMessageIds(fromChatMessageId)
		result.copiedMessageIds[fromChatMessageId] = toChatMessageIds

		for _, toChatMessageId := range toChatMessageIds {
			a := strings.Split(toChatMessageId, ":")
			// forwardRuleId := a[0]
			dstChatId := util.ConvertToInt[int64](a[1])
			tmpMessageId := util.ConvertToInt[int64](a[2])

			newMessageId := h.storageService.GetNewMessageId(dstChatId, tmpMessageId)
			if newMessageId == 0 {
				result = &data{needRepeat: true}
				return result
			}

			tmpChatMessageId := fmt.Sprintf("%d:%d", dstChatId, tmpMessageId)
			result.newMessageIds[tmpChatMessageId] = newMessageId
		}
	}

	return result
}

// deleteMessages удаляет сообщения
func (h *Handler) deleteMessages(chatId int64, messageIds []int64, data *data) {
	result := []string{}
	// errs := []error{} // TODO: собирать все ошибки (для тестов)
	defer func() {
		_ = result // TODO: костыль
	}()

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
				h.log.Error("forwardRule not found",
					"forwardRuleId", forwardRuleId,
					"fromChatMessageId", fromChatMessageId,
					"toChatMessageId", toChatMessageId,
				)
				continue
			}
			if !forwardRule.Indelible {
				continue
			}

			h.storageService.DeleteAnswerMessageId(dstChatId, tmpMessageId)

			tmpChatMessageId := fmt.Sprintf("%d:%d", dstChatId, tmpMessageId)
			newMessageId := data.newMessageIds[tmpChatMessageId]

			// TODO: может лучше удалять индексы _после_ удаления сообщения?
			h.storageService.DeleteTmpMessageId(dstChatId, newMessageId)
			h.storageService.DeleteNewMessageId(dstChatId, tmpMessageId)

			_, err := h.telegramRepo.GetClient().DeleteMessages(&client.DeleteMessagesRequest{
				ChatId:     dstChatId,
				MessageIds: []int64{newMessageId},
				Revoke:     true,
			})
			if err != nil {
				h.log.Error("DeleteMessages", "err", err)
				continue
			}

			result = append(result,
				fmt.Sprintf("%d:%d:%d", dstChatId, tmpMessageId, newMessageId))
		}

		if len(toChatMessageIds) > 0 {
			h.storageService.DeleteCopiedMessageIds(fromChatMessageId)
		}
	}
}
