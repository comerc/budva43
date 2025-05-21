package update_message_edited

import (
	"context"
	"log/slog"

	"github.com/zelenin/go-tdlib/client"
)

type telegramRepo interface {
	GetClient() *client.Client
}

type queueRepo interface {
	Add(task func())
}

type storageService interface {
	// SetCopiedMessageId(fromChatMessageId string, toChatMessageId string) error
	// GetCopiedMessageIds(fromChatMessageId string) ([]string, error)
	// DeleteCopiedMessageIds(fromChatMessageId string) error
	// SetNewMessageId(chatId, tmpMessageId, newMessageId int64) error
	// GetNewMessageId(chatId, tmpMessageId int64) (int64, error)
	// DeleteNewMessageId(chatId, tmpMessageId int64) error
	// SetTmpMessageId(chatId, newMessageId, tmpMessageId int64) error
	// GetTmpMessageId(chatId, newMessageId int64) (int64, error)
	// DeleteTmpMessageId(chatId, newMessageId int64) error
	// IncrementViewedMessages(toChatId int64, date string) error
	// // GetViewedMessages(toChatId int64, date string) (int64, error)
	// IncrementForwardedMessages(toChatId int64, date string) error
	// // GetForwardedMessages(toChatId int64, date string) (int64, error)
	// SetAnswerMessageId(dstChatId, tmpMessageId int64, fromChatMessageId string) error
}

type messageService interface {
	// GetFormattedText(message *client.Message) *client.FormattedText
	// IsSystemMessage(message *client.Message) bool
	// GetInputMessageContent(message *client.Message, formattedText *client.FormattedText) client.InputMessageContent
	// GetReplyMarkupData(message *client.Message) ([]byte, bool)
}

type mediaAlbumService interface {
	// AddMessage(key entity.MediaAlbumKey, message *client.Message) bool
	// GetLastReceivedDiff(key entity.MediaAlbumKey) time.Duration
	// PopMessages(key entity.MediaAlbumKey) []*client.Message
	// GetKey(forwardRuleId entity.ForwardRuleId, MediaAlbumId client.JsonInt64) entity.MediaAlbumKey
}

type Handler struct {
	log *slog.Logger
	ctx context.Context
	//
	telegramRepo       telegramRepo
	queueRepo          queueRepo
	storageService     storageService
	messageService     messageService
	mediaAlbumsService mediaAlbumService
}

func New(
	telegramRepo telegramRepo,
	queueRepo queueRepo,
	storageService storageService,
	messageService messageService,
	mediaAlbumsService mediaAlbumService,
) *Handler {
	return &Handler{
		log: slog.With("module", "handler.update_message_edited"),
		//
		telegramRepo:       telegramRepo,
		queueRepo:          queueRepo,
		storageService:     storageService,
		messageService:     messageService,
		mediaAlbumsService: mediaAlbumsService,
	}
}

// Run выполняет обрабатку обновления о редактировании сообщения
func (h *Handler) Run(update *client.UpdateMessageEdited) {
	// chatId := update.ChatId
	// messageId := update.MessageId

	// // Проверяем, является ли чат источником для какого-либо правила
	// if _, ok := isChatSource(chatId); !ok {
	// 	return
	// }

	// s.log.Debug("Обработка редактирования сообщения", "chatId", chatId, "messageId", messageId)

	// // Отправляем задачу в очередь
	// s.queueService.Add(func() {
	// 	// Формируем ключ для поиска скопированных сообщений
	// 	fromChatMessageId := fmt.Sprintf("%d:%d", chatId, messageId)

	// 	// Получаем идентификаторы скопированных сообщений
	// 	toChatMessageIds, err := s.storageService.GetCopiedMessageIds(fromChatMessageId)
	// 	if err != nil {
	// 		s.log.Error("Ошибка получения скопированных сообщений", "err", err)
	// 		return
	// 	}

	// 	if len(toChatMessageIds) == 0 {
	// 		s.log.Debug("Скопированные сообщения не найдены", "fromChatMessageId", fromChatMessageId)
	// 		return
	// 	}

	// 	tdlibClient := s.telegramRepo.GetClient()

	// 	// Получаем исходное сообщение
	// 	src, err := tdlibClient.GetMessage(&client.GetMessageRequest{
	// 		ChatId:    chatId,
	// 		MessageId: messageId,
	// 	})
	// 	if err != nil {
	// 		s.log.Error("Ошибка получения исходного сообщения", "err", err)
	// 		return
	// 	}

	// 	// Обрабатываем каждое скопированное сообщение
	// 	for _, toChatMessageId := range toChatMessageIds {
	// 		s.processSingleEdited(src, toChatMessageId)
	// 	}
	// })
}

// processSingleEdited обрабатывает редактирование сообщения
func (h *Handler) processSingleEdited(message *client.Message, toChatMessageId string) {
	// // Разбираем toChatMessageId
	// ruleId, dstChatId, dstMessageId, err := parseToChatMessageId(toChatMessageId)
	// if err != nil {
	// 	s.log.Error("Ошибка разбора toChatMessageId", "toChatMessageId", toChatMessageId, "err", err)
	// 	return
	// }

	// s.log.Debug("Обработка редактирования сообщения",
	// 	"srcChatId", message.ChatId,
	// 	"srcMessageId", message.Id,
	// 	"dstChatId", dstChatId,
	// 	"dstMessageId", dstMessageId,
	// 	"ruleId", ruleId)

	// // Получаем правило форвардинга
	// rule, ok := config.Engine.Forwards[ruleId]
	// if !ok {
	// 	s.log.Error("Правило форвардинга не найдено", "ruleId", ruleId)
	// 	return
	// }

	// // Если установлен флаг CopyOnce, не обрабатываем редактирование
	// if rule.CopyOnce {
	// 	s.log.Debug("Сообщение скопировано однократно, редактирование не применяется",
	// 		"ruleId", ruleId,
	// 		"dstChatId", dstChatId)
	// 	return
	// }

	// tdlibClient := s.telegramRepo.GetClient()

	// // Получаем оригинальное сообщение
	// srcMessage, err := tdlibClient.GetMessage(&client.GetMessageRequest{
	// 	ChatId:    message.ChatId,
	// 	MessageId: message.Id,
	// })
	// if err != nil {
	// 	s.log.Error("Ошибка получения исходного сообщения", "err", err)
	// 	return
	// }

	// // Получаем контент сообщения
	// formattedText, contentType := s.messageService.GetContent(srcMessage)
	// if contentType == "" {
	// 	s.log.Error("Неподдерживаемый тип сообщения при редактировании")
	// 	return
	// }

	// // Применяем трансформации к тексту
	// if err := s.transformService.ReplaceMyselfLinks(formattedText, srcMessage.ChatId, dstChatId); err != nil {
	// 	s.log.Error("Ошибка при замене ссылок", "err", err)
	// }
	// if err := s.transformService.ReplaceFragments(formattedText, dstChatId); err != nil {
	// 	s.log.Error("Ошибка при замене фрагментов", "err", err)
	// }

	// // В зависимости от типа контента, применяем соответствующее редактирование
	// switch contentType {
	// case client.TypeMessageText:
	// 	// Редактирование текста
	// 	_, err = tdlibClient.EditMessageText(&client.EditMessageTextRequest{
	// 		ChatId:    dstChatId,
	// 		MessageId: dstMessageId,
	// 		InputMessageContent: &client.InputMessageText{
	// 			Text: formattedText,
	// 		},
	// 	})
	// case client.TypeMessagePhoto:
	// 	// TODO: почему только тут применяется getInputMessageContent() ?
	// 	content := getInputMessageContent(srcMessage.Content, formattedText, contentType)
	// 	_, err = tdlibClient.EditMessageMedia(&client.EditMessageMediaRequest{
	// 		ChatId:              dstChatId,
	// 		MessageId:           dstMessageId,
	// 		InputMessageContent: content,
	// 	})
	// case client.TypeMessageVideo, client.TypeMessageDocument, client.TypeMessageAudio, client.TypeMessageAnimation:
	// 	// TODO: реализовать?
	// case client.TypeMessageVoiceNote:
	// 	// Редактирование подписи медиа
	// 	_, err = tdlibClient.EditMessageCaption(&client.EditMessageCaptionRequest{
	// 		ChatId:    dstChatId,
	// 		MessageId: dstMessageId,
	// 		Caption:   formattedText,
	// 	})
	// default:
	// 	err = fmt.Errorf("неподдерживаемый тип контента: %s", contentType)
	// }

	// if err != nil {
	// 	s.log.Error("Ошибка редактирования сообщения", "err", err)
	// } else {
	// 	s.log.Debug("Сообщение успешно отредактировано",
	// 		"dstChatId", dstChatId,
	// 		"dstMessageId", dstMessageId)
	// }
}
