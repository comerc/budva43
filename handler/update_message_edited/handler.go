package update_message_edited

import (
	"fmt"
	"strings"

	"github.com/zelenin/go-tdlib/client"

	"github.com/comerc/budva43/app/config"
	"github.com/comerc/budva43/app/entity"
	"github.com/comerc/budva43/app/log"
	"github.com/comerc/budva43/app/util"
)

//go:generate mockery --name=telegramRepo --exported
type telegramRepo interface {
	// tdlibClient methods
	GetMessage(*client.GetMessageRequest) (*client.Message, error)
	EditMessageText(*client.EditMessageTextRequest) (*client.Message, error)
	EditMessageCaption(*client.EditMessageCaptionRequest) (*client.Message, error)
}

//go:generate mockery --name=queueRepo --exported
type queueRepo interface {
	Add(fn func())
}

//go:generate mockery --name=storageService --exported
type storageService interface {
	GetCopiedMessageIds(chatId, messageId int64) []string
	GetNewMessageId(chatId, tmpMessageId int64) int64
	SetAnswerMessageId(dstChatId, tmpMessageId, chatId, messageId int64)
	DeleteAnswerMessageId(dstChatId, tmpMessageId int64)
}

//go:generate mockery --name=messageService --exported
type messageService interface {
	GetFormattedText(message *client.Message) *client.FormattedText
	GetInputMessageContent(message *client.Message, formattedText *client.FormattedText) client.InputMessageContent
	GetReplyMarkupData(message *client.Message) []byte
}

//go:generate mockery --name=transformService --exported
type transformService interface {
	Transform(formattedText *client.FormattedText, withSources bool, src *client.Message, dstChatId int64, engineConfig *entity.EngineConfig)
}

//go:generate mockery --name=filtersModeService --exported
type filtersModeService interface {
	Map(formattedText *client.FormattedText, forwardRule *entity.ForwardRule) entity.FiltersMode
}

//go:generate mockery --name=forwarderService --exported
type forwarderService interface {
	ForwardMessages(messages []*client.Message, filtersMode entity.FiltersMode, srcChatId, dstChatId int64, isSendCopy bool, forwardRuleId string, engineConfig *entity.EngineConfig)
}

type Handler struct {
	log *log.Logger
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
		log: log.NewLogger("handler.update_message_edited"),
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
	engineConfig := config.Engine // копируем, см. WATCH-CONFIG.md

	chatId := update.ChatId
	if _, ok := engineConfig.UniqueSources[chatId]; !ok {
		return
	}
	messageId := update.MessageId

	const maxRetries = 3
	retryCount := 0

	var fn func()
	fn = func() {
		var err error
		defer h.log.ErrorOrDebug(&err, "Run",
			"retryCount", &retryCount,
			"chatId", chatId,
			"messageId", messageId,
		)

		data := h.collectData(chatId, messageId)
		if data.needRepeat {
			retryCount++
			if retryCount >= maxRetries {
				err = log.NewError("max retries reached for message edit")
				return
			}
			h.queueRepo.Add(fn) // переставляем в конец очереди
			return
		}

		h.editMessages(chatId, messageId, data, engineConfig)
	}

	h.queueRepo.Add(fn)
}

type data struct {
	needRepeat       bool
	copiedMessageIds []string         // []toChatMessageId
	newMessageIds    map[string]int64 // tmpChatMessageId -> newMessageId
}

// collectData собирает данные для редактирования сообщений
func (h *Handler) collectData(chatId, messageId int64) *data {
	toChatMessageIds := h.storageService.GetCopiedMessageIds(chatId, messageId)
	result := &data{
		copiedMessageIds: toChatMessageIds,
		newMessageIds:    make(map[string]int64),
	}

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

	return result
}

// editMessages редактирует сообщения
func (h *Handler) editMessages(chatId, messageId int64, data *data, engineConfig *entity.EngineConfig) {
	var (
		err          error
		mediaAlbumId int64
		result       []string
	)
	defer h.log.ErrorOrDebug(&err, "editMessages",
		"chatId", chatId,
		"messageId", messageId,
		"mediaAlbumId", &mediaAlbumId,
		"result", &result,
	)

	toChatMessageIds := data.copiedMessageIds

	var src *client.Message
	src, err = h.telegramRepo.GetMessage(&client.GetMessageRequest{
		ChatId:    chatId,
		MessageId: messageId,
	})
	if err != nil {
		return
	}
	// TODO: isAnswer
	replyMarkupData := h.messageService.GetReplyMarkupData(src)
	srcFormattedText := h.messageService.GetFormattedText(src)
	mediaAlbumId = int64(src.MediaAlbumId)

	checkFns := make(map[int64]func())

	for _, toChatMessageId := range toChatMessageIds {
		func() {
			var err error
			forwardRuleId := ""
			defer h.log.ErrorOrDebug(&err, "editMessages",
				"chatId", chatId,
				"messageId", messageId,
				"toChatMessageId", toChatMessageId,
				"forwardRuleId", &forwardRuleId,
			)

			a := strings.Split(toChatMessageId, ":")
			forwardRuleId = a[0]
			dstChatId := util.ConvertToInt[int64](a[1])
			tmpMessageId := util.ConvertToInt[int64](a[2])

			forwardRule, ok := engineConfig.ForwardRules[forwardRuleId]
			if !ok {
				err = log.NewError("forwardRule not found")
				return
			}
			if forwardRule.CopyOnce {
				return
			}

			var formattedText *client.FormattedText
			formattedText, err = util.DeepCopy(srcFormattedText)
			if err != nil {
				err = log.WrapError(err) // внешняя ошибка
				return
			}
			if (forwardRule.SendCopy || src.CanBeSaved) &&
				h.filtersModeService.Map(formattedText, forwardRule) == entity.FiltersCheck {
				_, ok := checkFns[forwardRule.Check]
				if !ok {
					checkFns[forwardRule.Check] = func() {
						const isSendCopy = false // обязательно надо форвардить, иначе не видно текущего сообщения
						h.forwarderService.ForwardMessages([]*client.Message{src}, "", chatId, forwardRule.Check, isSendCopy, forwardRuleId, engineConfig)
					}
				}
				return
			}

			// TODO: почему не используется?
			// hasFiltersCheck := false
			// testChatId := dstChatId
			// var src *client.Message
			// for _, forwardRule := range engineConfig.ForwardRules {
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
			// 	return
			// }

			withSources := true
			h.transformService.Transform(formattedText, withSources, src, dstChatId, engineConfig)

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
				_, err = h.telegramRepo.EditMessageText(&client.EditMessageTextRequest{
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
				// if err != nil {
				//   //ничего не делаем, просто логируем ошибку
				// }
			case *client.MessageVoiceNote:
				_, err = h.telegramRepo.EditMessageCaption(&client.EditMessageCaptionRequest{
					ChatId:    dstChatId,
					MessageId: newMessageId,
					Caption:   formattedText,
				})
				// if err != nil {
				//   //ничего не делаем, просто логируем ошибку
				// }
			default:
				return
			}
			// TODO: isAnswer
			if len(replyMarkupData) > 0 {
				h.storageService.SetAnswerMessageId(dstChatId, tmpMessageId, chatId, messageId)
			} else {
				h.storageService.DeleteAnswerMessageId(dstChatId, tmpMessageId)
			}
		}()
	}

	for _, fn := range checkFns {
		fn()
	}
}
