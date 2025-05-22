package update_message_edited

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/comerc/budva43/config"
	"github.com/comerc/budva43/entity"
	"github.com/comerc/budva43/util"
	"github.com/zelenin/go-tdlib/client"
)

type telegramRepo interface {
	GetClient() *client.Client
}

type queueRepo interface {
	Add(task func())
}

type storageService interface {
	GetCopiedMessageIds(fromChatMessageId string) ([]string, error)
	GetNewMessageId(chatId, tmpMessageId int64) (int64, error)
	SetAnswerMessageId(dstChatId, tmpMessageId int64, fromChatMessageId string) error
	DeleteAnswerMessageId(dstChatId, tmpMessageId int64) error
}

type messageService interface {
	GetFormattedText(message *client.Message) *client.FormattedText
	GetInputMessageContent(message *client.Message, formattedText *client.FormattedText) client.InputMessageContent
	GetReplyMarkupData(message *client.Message) ([]byte, bool)
}

type transformService interface {
	Transform(formattedText *client.FormattedText, withSources bool, src *client.Message, dstChatId int64) error
}

type filtersModeService interface {
	Map(formattedText *client.FormattedText, forwardRule *entity.ForwardRule) entity.FiltersMode
}

type forwarderService interface {
	ForwardMessages(messages []*client.Message, srcChatId, dstChatId int64, isSendCopy bool, forwardRuleId string) error
}

type Handler struct {
	log *slog.Logger
	//
	telegramRepo       telegramRepo
	queueRepo          queueRepo
	storageService     storageService
	messageService     messageService
	transformService   transformService
	filtersModeService filtersModeService
	forwarderService   forwarderService
}

func New(
	telegramRepo telegramRepo,
	queueRepo queueRepo,
	storageService storageService,
	messageService messageService,
	transformService transformService,
	filtersModeService filtersModeService,
	forwarderService forwarderService,
) *Handler {
	return &Handler{
		log: slog.With("module", "handler.update_message_edited"),
		//
		telegramRepo:       telegramRepo,
		queueRepo:          queueRepo,
		storageService:     storageService,
		messageService:     messageService,
		transformService:   transformService,
		filtersModeService: filtersModeService,
		forwarderService:   forwarderService,
	}
}

// Run выполняет обрабатку обновления о редактировании сообщения
func (h *Handler) Run(update *client.UpdateMessageEdited) {
	chatId := update.ChatId
	if _, ok := config.Engine.UniqueSources[chatId]; !ok {
		return
	}
	messageId := update.MessageId

	const maxRetries = 3
	retryCount := 0

	var fn func()
	fn = func() {
		data, err := h.collectData(chatId, messageId)
		if err != nil {
			h.log.Error("collectData", "err", err)
			return
		}
		if data.needRepeat {
			retryCount++
			if retryCount >= maxRetries {
				h.log.Error("max retries reached for message edit",
					"chatId", chatId,
					"messageId", messageId,
				)
				return
			}
			h.log.Info("retrying message edit",
				"retryCount", retryCount,
				"chatId", chatId,
				"messageId", messageId,
			)
			h.queueRepo.Add(fn)
			return
		}

		result, _ := h.editMessages(chatId, messageId, data)
		h.log.Info("editMessages",
			"chatId", chatId,
			"messageId", messageId,
			"result", result,
		)
	}

	h.queueRepo.Add(fn)
}

type data struct {
	needRepeat       bool
	copiedMessageIds []string         // []toChatMessageId
	newMessageIds    map[string]int64 // tmpChatMessageId -> newMessageId
}

// collectData собирает данные для редактирования сообщений
func (h *Handler) collectData(chatId, messageId int64) (*data, error) {
	result := &data{}
	// errs := []error{} // TODO: собирать все ошибки (для тестов)

	fromChatMessageId := fmt.Sprintf("%d:%d", chatId, messageId)
	toChatMessageIds, _ := h.storageService.GetCopiedMessageIds(fromChatMessageId)
	result.copiedMessageIds = toChatMessageIds

	result.newMessageIds = make(map[string]int64)

	for _, toChatMessageId := range toChatMessageIds {
		a := strings.Split(toChatMessageId, ":")
		// forwardRuleId := a[0]
		dstChatId := util.ConvertToInt[int64](a[1])
		tmpMessageId := util.ConvertToInt[int64](a[2])

		newMessageId, err := h.storageService.GetNewMessageId(dstChatId, tmpMessageId)
		if err != nil {
			result = &data{needRepeat: true}
			return result, nil
		}

		tmpChatMessageId := fmt.Sprintf("%d:%d", dstChatId, tmpMessageId)
		result.newMessageIds[tmpChatMessageId] = newMessageId
	}

	return result, nil
}

// editMessages редактирует сообщения
func (h *Handler) editMessages(chatId, messageId int64, data *data) ([]string, error) {
	result := []string{}
	// errs := []error{} // TODO: собирать все ошибки (для тестов)

	var err error
	fromChatMessageId := fmt.Sprintf("%d:%d", chatId, messageId)
	toChatMessageIds := data.copiedMessageIds

	src, err := h.telegramRepo.GetClient().GetMessage(&client.GetMessageRequest{
		ChatId:    chatId,
		MessageId: messageId,
	})
	if err != nil {
		return nil, err
	}
	// TODO: isAnswer
	_, hasReplyMarkupData := h.messageService.GetReplyMarkupData(src)
	srcFormattedText := h.messageService.GetFormattedText(src)
	h.log.Info("editMessages",
		"chatId", src.ChatId,
		"messageId", src.Id,
		"hasText", srcFormattedText.Text != "",
		"mediaAlbumId", src.MediaAlbumId,
	)

	checkFns := make(map[int64]func())

	for _, toChatMessageId := range toChatMessageIds {
		a := strings.Split(toChatMessageId, ":")
		forwardRuleId := a[0]
		dstChatId := util.ConvertToInt[int64](a[1])
		tmpMessageId := util.ConvertToInt[int64](a[2])

		forwardRule, ok := config.Engine.ForwardRules[forwardRuleId]
		if !ok {
			h.log.Error("forwardRule not found",
				"forwardRuleId", forwardRuleId,
				// "fromChatMessageId", fromChatMessageId,
				"toChatMessageId", toChatMessageId,
			)
			continue
		}

		if forwardRule.CopyOnce {
			continue
		}

		formattedText := util.Copy(srcFormattedText)
		if (forwardRule.SendCopy || src.CanBeSaved) &&
			h.filtersModeService.Map(formattedText, forwardRule) == entity.FiltersCheck {
			_, ok := checkFns[forwardRule.Check]
			if !ok {
				checkFns[forwardRule.Check] = func() {
					const isSendCopy = false // обязательно надо форвардить, иначе не видно текущего сообщения
					h.forwarderService.ForwardMessages([]*client.Message{src}, chatId, forwardRule.Check, isSendCopy, forwardRuleId)
				}
			}
			continue
		}

		// TODO: почему не используется?
		// hasFiltersCheck := false
		// testChatId := dstChatId
		// var src *client.Message
		// for _, forwardRule := range config.Engine.ForwardRules {
		// 	if src.ChatId == forwardRule.From && (forwardRule.SendCopy || src.CanBeSaved) {
		// 		for _, dstChatId := range forwardRule.To {
		// 			if testChatId == dstChatId {
		// 				if h.filtersModeService.Map(formattedText, forwardRule) == entity.FiltersCheck {
		// 					hasFiltersCheck = true
		// 					_, ok := checkFns[forwardRule.Check]
		// 					if !ok {
		// 						checkFns[forwardRule.Check] = func() {
		// 							const isSendCopy = false // обязательно надо форвардить, иначе не видно текущего сообщения
		// 							h.forwarderService.ForwardMessages([]*client.Message{src}, chatId, forwardRule.Check, isSendCopy, forwardRule.Id)
		// 						}
		// 					}
		// 				}
		// 			}
		// 		}
		// 	}
		// }
		// if hasFiltersCheck {
		// 	continue
		// }

		withSources := true
		if err := h.transformService.Transform(formattedText, withSources, src, dstChatId); err != nil {
			h.log.Error("Transform", "err", err)
		}

		tmpChatMessageId := fmt.Sprintf("%d:%d", dstChatId, tmpMessageId)
		newMessageId := data.newMessageIds[tmpChatMessageId]
		result = append(result, fmt.Sprintf("toChatMessageId: %s, newMessageId: %d", toChatMessageId, newMessageId))

		switch src.Content.(type) {
		case
			*client.MessageText,
			*client.MessageAnimation,
			*client.MessageDocument,
			*client.MessageAudio,
			*client.MessageVideo,
			*client.MessagePhoto:
			content := h.messageService.GetInputMessageContent(src, formattedText)
			dst, err := h.telegramRepo.GetClient().EditMessageText(&client.EditMessageTextRequest{
				ChatId:              dstChatId,
				MessageId:           newMessageId,
				InputMessageContent: content,
				// ReplyMarkup: func() client.ReplyMarkup {
				// 	if src.Content.(type).MessageContentType() == client.TypeMessageText {
				// 		return src.ReplyMarkup // это не надо, юзер-бот игнорит изменение
				// 	}
				// 	return nil
				// }(),
			})
			if err != nil {
				h.log.Error("EditMessageText", "err", err)
			}
			h.log.Info("EditMessageText", "dst", dst)
		case *client.MessageVoiceNote:
			dst, err := h.telegramRepo.GetClient().EditMessageCaption(&client.EditMessageCaptionRequest{
				ChatId:    dstChatId,
				MessageId: newMessageId,
				Caption:   formattedText,
			})
			if err != nil {
				h.log.Error("EditMessageCaption", "err", err)
			}
			h.log.Info("EditMessageCaption", "dst", dst)
		default:
			continue
		}
		// TODO: isAnswer
		if hasReplyMarkupData {
			h.storageService.SetAnswerMessageId(dstChatId, tmpMessageId, fromChatMessageId)
		} else {
			h.storageService.DeleteAnswerMessageId(dstChatId, tmpMessageId)
		}
	}

	for _, fn := range checkFns {
		fn()
	}

	return result, nil
}
