package engine

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"slices"
	"time"

	"github.com/zelenin/go-tdlib/client"

	"github.com/comerc/budva43/config"
	"github.com/comerc/budva43/entity"
	"github.com/comerc/budva43/util"
)

type telegramRepo interface {
	GetClient() *client.Client
	GetClientDone() <-chan any
}

type updateNewMessageHandler interface {
	Run(ctx context.Context, update *client.UpdateNewMessage)
}

type updateDeleteMessagesHandler interface {
	Run(update *client.UpdateDeleteMessages)
}

type queueRepo interface {
	Add(task func())
}

type storageService interface {
	SetCopiedMessageId(fromChatMessageId string, toChatMessageId string) error
	GetCopiedMessageIds(fromChatMessageId string) ([]string, error)
	DeleteCopiedMessageIds(fromChatMessageId string) error
	SetNewMessageId(chatId, tmpMessageId, newMessageId int64) error
	GetNewMessageId(chatId, tmpMessageId int64) (int64, error)
	DeleteNewMessageId(chatId, tmpMessageId int64) error
	SetTmpMessageId(chatId, newMessageId, tmpMessageId int64) error
	GetTmpMessageId(chatId, newMessageId int64) (int64, error)
	DeleteTmpMessageId(chatId, newMessageId int64) error
	IncrementViewedMessages(toChatId int64, date string) error
	// GetViewedMessages(toChatId int64, date string) (int64, error)
	IncrementForwardedMessages(toChatId int64, date string) error
	// GetForwardedMessages(toChatId int64, date string) (int64, error)
	SetAnswerMessageId(dstChatId, tmpMessageId int64, fromChatMessageId string) error
}

type messageService interface {
	GetFormattedText(message *client.Message) *client.FormattedText
	IsSystemMessage(message *client.Message) bool
	GetInputMessageContent(message *client.Message, formattedText *client.FormattedText) client.InputMessageContent
	GetReplyMarkupData(message *client.Message) ([]byte, bool)
}

type mediaAlbumService interface {
	AddMessage(key entity.MediaAlbumKey, message *client.Message) bool
	GetLastReceivedDiff(key entity.MediaAlbumKey) time.Duration
	PopMessages(key entity.MediaAlbumKey) []*client.Message
	GetKey(forwardRuleId entity.ForwardRuleId, MediaAlbumId client.JsonInt64) entity.MediaAlbumKey
}

type transformService interface {
	Transform(formattedText *client.FormattedText, isFirstMessageInAlbum bool, src *client.Message, dstChatId int64) error
}

type rateLimiterService interface {
	WaitForForward(ctx context.Context, dstChatId int64)
}

// Service предоставляет функциональность движка пересылки сообщений
type Service struct {
	log *slog.Logger
	ctx context.Context
	//
	telegramRepo                telegramRepo
	queueRepo                   queueRepo
	storageService              storageService
	messageService              messageService
	mediaAlbumsService          mediaAlbumService
	transformService            transformService
	rateLimiterService          rateLimiterService
	updateNewMessageHandler     updateNewMessageHandler
	updateDeleteMessagesHandler updateDeleteMessagesHandler
}

// New создает новый экземпляр сервиса engine
func New(
	telegramRepo telegramRepo,
	queueRepo queueRepo,
	storageService storageService,
	messageService messageService,
	mediaAlbumsService mediaAlbumService,
	transformService transformService,
	rateLimiterService rateLimiterService,
	updateNewMessageHandler updateNewMessageHandler,
	updateDeleteMessagesHandler updateDeleteMessagesHandler,
) *Service {
	return &Service{
		log: slog.With("module", "service.engine"),
		//
		telegramRepo:                telegramRepo,
		queueRepo:                   queueRepo,
		storageService:              storageService,
		messageService:              messageService,
		mediaAlbumsService:          mediaAlbumsService,
		transformService:            transformService,
		rateLimiterService:          rateLimiterService,
		updateNewMessageHandler:     updateNewMessageHandler,
		updateDeleteMessagesHandler: updateDeleteMessagesHandler,
	}
}

// Start запускает обработчик обновлений от Telegram
func (s *Service) Start(ctx context.Context) error {
	s.log.Info("Запуск сервиса engine")
	s.ctx = ctx

	return nil

	// Проверяем конфигурацию
	if err := s.validateConfig(); err != nil {
		return fmt.Errorf("ошибка валидации конфигурации: %w", err)
	}

	// Обогащаем конфигурацию
	if err := s.enrichConfig(); err != nil {
		return fmt.Errorf("ошибка обогащения конфигурации: %w", err)
	}

	go s.run()

	return nil
}

// Close останавливает сервис
func (s *Service) Close() error {
	return nil
}

// validateConfig проверяет корректность конфигурации
func (s *Service) validateConfig() error {
	for chatId, dsc := range config.Engine.Destinations {
		for _, replaceFragment := range dsc.ReplaceFragments {
			if util.RuneCountForUTF16(replaceFragment.From) != util.RuneCountForUTF16(replaceFragment.To) {
				return fmt.Errorf("длина исходного и заменяемого текста должна быть одинаковой: %s -> %s", replaceFragment.From, replaceFragment.To)
			}
		}
		s.log.Info("Валидированы настройки замены фрагментов", "chatId", chatId, "replacements", len(dsc.ReplaceFragments))
	}

	re := regexp.MustCompile("[:,]") // TODO: зачем нужна эта проверка? (предположительно для badger)
	for forwardRuleId, forwardRule := range config.Engine.ForwardRules {
		if re.FindString(forwardRuleId) != "" {
			return fmt.Errorf("нельзя использовать [:,] в идентификаторе правила: %s", forwardRuleId)
		}
		for _, dstChatId := range forwardRule.To {
			if forwardRule.From == dstChatId {
				return fmt.Errorf("идентификатор получателя не может совпадать с идентификатором источника: %d", dstChatId)
			}
		}
		s.log.Info("Валидировано правило пересылки", "forwardRuleId", forwardRuleId, "from", forwardRule.From, "to", forwardRule.To)
	}

	return nil
}

// enrichConfig обогащает конфигурацию
func (s *Service) enrichConfig() error {
	config.Engine.UniqueSources = make(map[entity.ChatId]struct{})
	tmpOrderedForwardRules := make([]entity.ForwardRuleId, 0)
	for key, destination := range config.Engine.Destinations {
		destination.ChatId = key
	}
	for key, source := range config.Engine.Sources {
		source.ChatId = key
	}
	for key, forwardRule := range config.Engine.ForwardRules {
		forwardRule.Id = key
		if _, ok := config.Engine.Sources[forwardRule.From]; !ok {
			config.Engine.Sources[forwardRule.From] = &entity.Source{
				ChatId: forwardRule.From,
			}
		}
		config.Engine.UniqueSources[forwardRule.From] = struct{}{}
		tmpOrderedForwardRules = append(tmpOrderedForwardRules, forwardRule.Id)
	}
	slices.Sort(tmpOrderedForwardRules)
	config.Engine.OrderedForwardRules = slices.Compact(tmpOrderedForwardRules)
	return nil
}

// run запускает обработчик обновлений от Telegram
func (s *Service) run() {
	// Ждём авторизации клиента и получаем канал обновлений от Telegram
	select {
	case <-s.ctx.Done():
		return
	case <-s.telegramRepo.GetClientDone():
		listener := s.telegramRepo.GetClient().GetListener()
		defer listener.Close()
		s.handleUpdates(listener)
	}
}

// handleUpdates обрабатывает обновления от Telegram
func (s *Service) handleUpdates(listener *client.Listener) {
	for {
		select {
		case <-s.ctx.Done():
			return
		case update, ok := <-listener.Updates:
			if !ok {
				return
			}

			if update.GetClass() != client.ClassUpdate {
				continue
			}

			switch updateByType := update.(type) {
			case *client.UpdateNewMessage:
				s.updateNewMessageHandler.Run(s.ctx, updateByType)
			case *client.UpdateMessageEdited:
				s.handleUpdateMessageEdited(updateByType)
			case *client.UpdateDeleteMessages:
				s.updateDeleteMessagesHandler.Run(updateByType)
			case *client.UpdateMessageSendSucceeded:
				s.handleUpdateMessageSendSucceeded(updateByType)
			}
		}
	}
}

// handleUpdateMessageEdited обрабатывает обновление о редактировании сообщения
func (s *Service) handleUpdateMessageEdited(update *client.UpdateMessageEdited) {
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

// handleUpdateMessageSendSucceeded обрабатывает обновление об успешной отправке сообщения
func (s *Service) handleUpdateMessageSendSucceeded(update *client.UpdateMessageSendSucceeded) {
	message := update.Message
	tmpMessageId := update.OldMessageId
	fn := func() {
		_ = s.storageService.SetNewMessageId(message.ChatId, tmpMessageId, message.Id)
		_ = s.storageService.SetTmpMessageId(message.ChatId, message.Id, tmpMessageId)
		s.log.Info("handleUpdateMessageSendSucceeded")
	}
	s.queueRepo.Add(fn)
}
