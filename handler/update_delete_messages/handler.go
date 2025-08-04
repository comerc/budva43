package update_delete_messages

import (
	"fmt"
	"strings"

	"github.com/zelenin/go-tdlib/client"

	"github.com/comerc/budva43/app/config"
	"github.com/comerc/budva43/app/domain"
	"github.com/comerc/budva43/app/log"
	"github.com/comerc/budva43/app/util"
)

//go:generate mockery --name=telegramRepo --exported
type telegramRepo interface {
	// tdlibClient methods
	DeleteMessages(*client.DeleteMessagesRequest) (*client.Ok, error)
}

//go:generate mockery --name=queueRepo --exported
type queueRepo interface {
	Add(fn func())
}

//go:generate mockery --name=storageService --exported
type storageService interface {
	GetCopiedMessageIds(chatId, messageId int64) []string
	DeleteCopiedMessageIds(chatId, messageId int64)
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
		log: log.NewLogger(),
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

	engineConfig := config.Engine // копируем, см. WATCH-CONFIG.md

	chatId := update.ChatId
	if _, ok := engineConfig.UniqueSources[update.ChatId]; !ok {
		return
	}
	messageIds := update.MessageIds

	const maxRetries = 3
	retryCount := 0

	var fn func()
	fn = func() {
		var err error
		defer func() {
			h.log.ErrorOrDebug(err, "",
				"retryCount", retryCount,
				"chatId", chatId,
				"messageIds", messageIds,
			)
		}()

		data := h.collectData(chatId, messageIds)
		if data.needRepeat {
			retryCount++
			if retryCount >= maxRetries {
				err = log.NewError("max retries reached for message deletion")
				return
			}
			h.queueRepo.Add(fn) // переставляем в конец очереди
			return
		}

		h.deleteMessages(chatId, messageIds, data, engineConfig)
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

	for _, messageId := range messageIds {
		toChatMessageIds := h.storageService.GetCopiedMessageIds(chatId, messageId)
		fromChatMessageId := fmt.Sprintf("%d:%d", chatId, messageId)
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
func (h *Handler) deleteMessages(chatId int64, messageIds []int64, data *data, engineConfig *domain.EngineConfig) {
	var err error
	result := []string{}
	defer func() {
		h.log.ErrorOrDebug(err, "",
			"chatId", chatId,
			"messageIds", messageIds,
			"result", result,
		)
	}()

	for _, messageId := range messageIds {
		fromChatMessageId := fmt.Sprintf("%d:%d", chatId, messageId)
		toChatMessageIds := data.copiedMessageIds[fromChatMessageId]

		for _, toChatMessageId := range toChatMessageIds {
			func() {
				var err error
				forwardRuleId := ""
				defer func() {
					h.log.ErrorOrDebug(err, "",
						"chatId", chatId,
						"messageId", messageId,
						"toChatMessageId", toChatMessageId,
						"forwardRuleId", forwardRuleId,
					)
				}()

				a := strings.Split(toChatMessageId, ":")
				forwardRuleId = a[0]
				dstChatId := util.ConvertToInt[int64](a[1])
				tmpMessageId := util.ConvertToInt[int64](a[2])

				forwardRule, ok := engineConfig.ForwardRules[forwardRuleId]
				if !ok {
					err = log.NewError("forwardRule not found")
					return
				}
				if forwardRule.Indelible {
					return
				}

				h.storageService.DeleteAnswerMessageId(dstChatId, tmpMessageId)

				tmpChatMessageId := fmt.Sprintf("%d:%d", dstChatId, tmpMessageId)
				newMessageId := data.newMessageIds[tmpChatMessageId]

				// TODO: может лучше удалять индексы _после_ удаления сообщения?
				h.storageService.DeleteTmpMessageId(dstChatId, newMessageId)
				h.storageService.DeleteNewMessageId(dstChatId, tmpMessageId)

				_, err = h.telegramRepo.DeleteMessages(&client.DeleteMessagesRequest{
					ChatId:     dstChatId,
					MessageIds: []int64{newMessageId},
					Revoke:     true,
				})
				if err != nil {
					return
				}

				result = append(result,
					fmt.Sprintf("%d:%d:%d", dstChatId, tmpMessageId, newMessageId))
			}()
		}

		if len(toChatMessageIds) > 0 {
			h.storageService.DeleteCopiedMessageIds(chatId, messageId)
		}
	}
}
