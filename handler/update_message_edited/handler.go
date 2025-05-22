package update_message_edited

import (
	"context"
	"fmt"
	"log"
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
	ctx context.Context
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
	repeat := 0
	var fn func()
	fn = func() {
		var (
			result []string
			src    *client.Message
		)
		fromChatMessageId := fmt.Sprintf("%d:%d", chatId, messageId)
		toChatMessageIds, _ := h.storageService.GetCopiedMessageIds(fromChatMessageId)
		defer func() {
			args := []any{
				"fromChatMessageId", fromChatMessageId,
				"toChatMessageIds", toChatMessageIds,
			}
			if src != nil {
				args = append(args, "contentType", src.Content.MessageContentType())
			}
			if len(result) > 0 {
				args = append(args, "result", result)
			}
			h.log.Info("UpdateMessageEdited", args...)
		}()
		if len(toChatMessageIds) == 0 {
			return
		}
		var newMessageIds = make(map[string]int64)
		isUpdateMessageSendSucceeded := true
		for _, toChatMessageId := range toChatMessageIds {
			a := strings.Split(toChatMessageId, ":")
			// forwardRuleId := a[0]
			dstChatId := util.ConvertToInt[int64](a[1])
			tmpMessageId := util.ConvertToInt[int64](a[2])
			newMessageId, _ := h.storageService.GetNewMessageId(dstChatId, tmpMessageId)
			if newMessageId == 0 {
				isUpdateMessageSendSucceeded = false
				break
			}
			tmpChatMessageId := fmt.Sprintf("%d:%d", dstChatId, tmpMessageId)
			newMessageIds[tmpChatMessageId] = newMessageId
		}
		if !isUpdateMessageSendSucceeded {
			repeat++
			if repeat < 3 {
				log.Print("isUpdateMessageSendSucceeded > repeat: ", repeat)
				h.queueRepo.Add(fn)
			} else {
				log.Print("isUpdateMessageSendSucceeded > repeat limit !!!")
			}
			return
		}
		src, err := h.telegramRepo.GetClient().GetMessage(&client.GetMessageRequest{
			ChatId:    chatId,
			MessageId: messageId,
		})
		if err != nil {
			h.log.Error("GetMessage", "err", err)
			return
		}
		// TODO: isAnswer
		_, hasReplyMarkupData := h.messageService.GetReplyMarkupData(src)
		srcFormattedText := h.messageService.GetFormattedText(src)
		log.Printf("srcChatId: %d srcId: %d hasText: %t MediaAlbumId: %d", src.ChatId, src.Id, srcFormattedText.Text != "", src.MediaAlbumId)
		checkFns := make(map[int64]func())
		for _, toChatMessageId := range toChatMessageIds {
			a := strings.Split(toChatMessageId, ":")
			forwardRuleId := a[0]
			dstChatId := util.ConvertToInt[int64](a[1])
			tmpMessageId := util.ConvertToInt[int64](a[2])
			formattedText := util.Copy(srcFormattedText)
			forwardRule, ok := config.Engine.ForwardRules[forwardRuleId]
			if !ok {
				h.log.Error("forwardRule not found",
					"forwardRuleId", forwardRuleId,
					"fromChatMessageId", fromChatMessageId,
					"toChatMessageId", toChatMessageId,
				)
				continue
			}
			if forwardRule.CopyOnce {
				continue
			}
			if (forwardRule.SendCopy || src.CanBeSaved) &&
				h.filtersModeService.Map(formattedText, forwardRule) == entity.FiltersCheck {

				_, ok := checkFns[forwardRule.Check]
				if !ok {
					checkFns[forwardRule.Check] = func() {
						const isSendCopy = false // обязательно надо форвардить, иначе не видно текущего сообщения
						h.forwarderService.ForwardMessages([]*client.Message{src}, src.ChatId, forwardRule.Check, isSendCopy, forwardRuleId)
					}
				}
				continue
			}
			// hasFiltersCheck := false
			// testChatId := dstChatId
			// for _, forward := range configData.Forwards {
			// 	if src.ChatId == forward.From && (forward.SendCopy || src.CanBeForwarded) {
			// 		for _, dstChatId := range forward.To {
			// 			if testChatId == dstChatId {
			// 				if checkFilters(formattedText, forward) == FiltersCheck {
			// 					hasFiltersCheck = true
			// 					_, ok := checkFns[forward.Check]
			// 					if !ok {
			// 						checkFns[forward.Check] = func() {
			// 							const isSendCopy = false // обязательно надо форвардить, иначе не видно текущего сообщения
			// 							forwardNewMessages(tdlibClient, []*client.Message{src}, src.ChatId, forward.Check, isSendCopy)
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
			newMessageId := newMessageIds[tmpChatMessageId]
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
					log.Print("EditMessageText > ", err)
				}
				log.Printf("EditMessageText > dst: %#v", dst)
			case *client.MessageVoiceNote:
				dst, err := h.telegramRepo.GetClient().EditMessageCaption(&client.EditMessageCaptionRequest{
					ChatId:    dstChatId,
					MessageId: newMessageId,
					Caption:   formattedText,
				})
				if err != nil {
					log.Print("EditMessageCaption > ", err)
				}
				log.Printf("EditMessageCaption > dst: %#v", dst)
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
		for check, fn := range checkFns {
			if fn == nil {
				log.Printf("check: %d is nil", check)
				continue
			}
			log.Printf("check: %d is fn()", check)
			fn()
		}
	}
	h.queueRepo.Add(fn)
}
