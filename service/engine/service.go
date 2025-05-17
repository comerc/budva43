package engine

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"slices"
	"sync"
	"time"

	"github.com/zelenin/go-tdlib/client"

	"github.com/comerc/budva43/config"
	"github.com/comerc/budva43/entity"
	"github.com/comerc/budva43/util"
)

// TODO: serviceEngine слишком большой, что можно вынести в другие сервисы?

type telegramRepo interface {
	GetClient() *client.Client
	GetClientDone() <-chan any
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
}

type messageService interface {
	GetContent(message *client.Message) (*client.FormattedText, string)
	IsSystemMessage(message *client.Message) bool
	GetInputMessageContent(message *client.Message, formattedText *client.FormattedText) client.InputMessageContent
	// GetContentType(message *client.Message) string
	// SendMessage(chatId int64, text string) (*client.Message, error)
	// ForwardMessage(fromChatId, messageId, toChatId int64) (*client.Message, error)
	// SendMessageAlbum(chatId int64, contents []client.InputMessageContent) (*client.Messages, error)
	// ForwardMessages(fromChatId int64, messageIds []int64, toChatId int64) (*client.Messages, error)
	// EditMessageText(chatId, messageId int64, text *client.FormattedText) (*client.Message, error)
	// EditMessageMedia(chatId, messageId int64, content client.InputMessageContent) (*client.Message, error)
	// EditMessageCaption(chatId, messageId int64, caption *client.FormattedText) (*client.Message, error)
	// DeleteMessages(chatId int64, messageIds []int64) error
	// GetMessage(chatId, messageId int64) (*client.Message, error)
}

type mediaAlbumService interface {
	AddMessage(forwardRuleId entity.ForwardRuleId, message *client.Message) bool
	GetLastReceivedDiff(key entity.MediaAlbumForwardKey) time.Duration
	GetMessages(key entity.MediaAlbumForwardKey) []*client.Message
}

type transformService interface {
	ReplaceMyselfLinks(formattedText *client.FormattedText, srcChatId, dstChatId int64) error
	ReplaceFragments(formattedText *client.FormattedText, dstChatId int64) error
	AddSources(formattedText *client.FormattedText, message *client.Message, dstChatId int64) error
}

type rateLimiterService interface {
	WaitForForward(ctx context.Context, dstChatId int64)
}

// Service предоставляет функциональность движка пересылки сообщений
type Service struct {
	log *slog.Logger
	//
	telegramRepo       telegramRepo
	queueRepo          queueRepo
	storageService     storageService
	messageService     messageService
	mediaAlbumsService mediaAlbumService
	transformService   transformService
	rateLimiterService rateLimiterService
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
) *Service {
	return &Service{
		log: slog.With("module", "service.engine"),
		//
		telegramRepo:       telegramRepo,
		queueRepo:          queueRepo,
		storageService:     storageService,
		messageService:     messageService,
		mediaAlbumsService: mediaAlbumsService,
		transformService:   transformService,
		rateLimiterService: rateLimiterService,
	}
}

// Start запускает обработчик обновлений от Telegram
func (s *Service) Start(ctx context.Context) error {
	s.log.Info("Запуск сервиса engine")

	return nil

	// Проверяем конфигурацию
	if err := s.validateConfig(); err != nil {
		return fmt.Errorf("ошибка валидации конфигурации: %w", err)
	}

	// Обогащаем конфигурацию
	if err := s.enrichConfig(); err != nil {
		return fmt.Errorf("ошибка обогащения конфигурации: %w", err)
	}

	go s.run(ctx)

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
	for ruleId, rule := range config.Engine.ForwardRules {
		if re.FindString(ruleId) != "" {
			return fmt.Errorf("нельзя использовать [:,] в идентификаторе правила: %s", ruleId)
		}
		for _, dstChatId := range rule.To {
			if rule.From == dstChatId {
				return fmt.Errorf("идентификатор получателя не может совпадать с идентификатором источника: %d", dstChatId)
			}
		}
		s.log.Info("Валидировано правило пересылки", "ruleId", ruleId, "from", rule.From, "to", rule.To)
	}

	return nil
}

// enrichConfig обогащает конфигурацию
func (s *Service) enrichConfig() error {
	config.Engine.UniqueFrom = make(map[entity.ChatId]struct{})
	for key, val := range config.Engine.Destinations {
		val.ChatId = key
	}
	for key, val := range config.Engine.Sources {
		val.ChatId = key
	}
	for key, rule := range config.Engine.ForwardRules {
		rule.Id = key
		if _, ok := config.Engine.Sources[rule.From]; !ok {
			config.Engine.Sources[rule.From] = &entity.Source{
				ChatId: rule.From,
			}
		}
		config.Engine.UniqueFrom[rule.From] = struct{}{}
	}
	return nil
}

// run запускает обработчик обновлений от Telegram
func (s *Service) run(ctx context.Context) {
	// Ждём авторизации клиента и получаем канал обновлений от Telegram
	select {
	case <-ctx.Done():
		return
	case <-s.telegramRepo.GetClientDone():
		listener := s.telegramRepo.GetClient().GetListener()
		defer listener.Close()
		s.handleUpdates(ctx, listener)
	}
}

// handleUpdates обрабатывает обновления от Telegram
func (s *Service) handleUpdates(ctx context.Context, listener *client.Listener) {
	for {
		select {
		case <-ctx.Done():
			return
		case update, ok := <-listener.Updates:
			if !ok {
				return
			}

			if update.GetClass() != client.ClassUpdate {
				continue
			}

			switch updateType := update.(type) {
			case *client.UpdateNewMessage:
				s.handleUpdateNewMessage(updateType)
			case *client.UpdateMessageEdited:
				s.handleUpdateMessageEdited(updateType)
			case *client.UpdateDeleteMessages:
				s.handleUpdateDeleteMessages(updateType)
			case *client.UpdateMessageSendSucceeded:
				s.handleUpdateMessageSendSucceeded(updateType)
			}
		}
	}
}

// handleUpdateNewMessage обрабатывает обновление о новом сообщении
func (s *Service) handleUpdateNewMessage(update *client.UpdateNewMessage) {
	// message := update.Message
	// chatId := message.ChatId

	// // Проверяем, является ли чат источником для какого-либо правила
	// isSourceChat := false
	// var forwardRules []struct {
	// 	ruleId string
	// 	rule   entity.ForwardRule
	// }

	// for ruleId, rule := range config.Engine.Forwards {
	// 	if rule.From == chatId && rule.Status == entity.RuleStatusActive {
	// 		isSourceChat = true
	// 		forwardRules = append(forwardRules, struct {
	// 			ruleId string
	// 			rule   entity.ForwardRule
	// 		}{ruleId, rule})
	// 	}
	// }

	// if !isSourceChat {
	// 	return
	// }

	// // Проверяем удаление системных сообщений
	// if shouldDelete := config.Engine.DeleteSystemMessages[chatId]; shouldDelete {
	// 	if s.messageService.IsSystemMessage(message) {
	// 		go func() {
	// 			tdlibClient := s.telegramRepo.GetClient()
	// 			_, err := tdlibClient.DeleteMessages(&client.DeleteMessagesRequest{
	// 				ChatId:     chatId,
	// 				MessageIds: []int64{message.Id},
	// 				Revoke:     true, // Удаляем для всех участников, а не только для себя
	// 			})
	// 			if err != nil {
	// 				s.log.Error("Ошибка удаления системного сообщения", "err", err)
	// 			}
	// 		}()
	// 	}
	// }

	// formattedText, _ := s.messageService.GetContent(message)

	// // Обрабатываем каждое правило
	// for _, ruleData := range forwardRules {
	// 	// Проверяем правило
	// 	rule := ruleData.rule

	// 	// Проверяем, должно ли сообщение быть переслано согласно фильтрам
	// 	shouldForward, err := s.filterService.ShouldForward(formattedText.Text, &rule)
	// 	if err != nil {
	// 		s.log.Error("Ошибка проверки фильтров", "err", err)
	// 		continue
	// 	}

	// 	if !shouldForward {
	// 		s.log.Debug("Сообщение не проходит фильтры", "ruleId", ruleData.ruleId)
	// 		continue
	// 	}

	// 	// Обрабатываем сообщение в зависимости от типа
	// 	if message.MediaAlbumId == 0 {
	// 		// Одиночное сообщение
	// 		s.queueService.Add(func() {
	// 			s.processMessage([]*client.Message{message}, ruleData.ruleId, rule)
	// 		})
	// 	} else {
	// 		// Медиа-альбом
	// 		isFirstMessage := s.mediaAlbumsService.AddMessage(ruleData.ruleId, message)
	// 		if isFirstMessage {
	// 			s.queueService.Add(func() {
	// 				s.processMediaAlbum(ruleData.ruleId, message.MediaAlbumId)
	// 			})
	// 		}
	// 	}
	// }
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

// handleUpdateDeleteMessages обрабатывает обновление об удалении сообщений
func (s *Service) handleUpdateDeleteMessages(update *client.UpdateDeleteMessages) {
	// // Обрабатываем только постоянное удаление сообщений
	// if !update.IsPermanent {
	// 	return
	// }

	// chatId := update.ChatId
	// messageIds := update.MessageIds

	// // Проверяем, является ли чат источником для какого-либо правила
	// if _, ok := isChatSource(chatId); !ok {
	// 	return
	// }

	// s.log.Debug("Обработка удаления сообщений", "chatId", chatId, "messageIds", messageIds)

	// // Отправляем задачу в очередь
	// s.queueService.Add(func() {
	// 	// Обрабатываем каждое удаленное сообщение
	// 	for _, messageId := range messageIds {
	// 		// Формируем ключ для поиска скопированных сообщений
	// 		fromChatMessageId := fmt.Sprintf("%d:%d", chatId, messageId)

	// 		// Получаем идентификаторы скопированных сообщений
	// 		toChatMessageIds, err := s.storageService.GetCopiedMessageIds(fromChatMessageId)
	// 		if err != nil {
	// 			s.log.Error("Ошибка получения скопированных сообщений", "err", err)
	// 			continue
	// 		}

	// 		if len(toChatMessageIds) == 0 {
	// 			continue
	// 		}

	// 		// Обрабатываем каждое скопированное сообщение
	// 		for _, toChatMessageId := range toChatMessageIds {
	// 			s.processSingleDeleted(fromChatMessageId, toChatMessageId)
	// 		}

	// 		// Удаляем соответствие между оригинальным и скопированными сообщениями
	// 		err = s.storageService.DeleteCopiedMessageIds(fromChatMessageId)
	// 		if err != nil {
	// 			s.log.Error("Ошибка удаления скопированных сообщений", "err", err)
	// 		}
	// 	}
	// })
}

// handleUpdateMessageSendSucceeded обрабатывает обновление об успешной отправке сообщения
func (s *Service) handleUpdateMessageSendSucceeded(update *client.UpdateMessageSendSucceeded) {
	// message := update.Message
	// chatId := message.ChatId
	// messageId := message.Id
	// oldMessageId := update.OldMessageId

	// s.log.Debug("Обработка успешной отправки сообщения",
	// 	"chatId", chatId,
	// 	"messageId", messageId,
	// 	"oldMessageId", oldMessageId)

	// // Отправляем задачу в очередь
	// s.queueService.Add(func() {
	// 	// Сохраняем соответствие между временным и постоянным Id сообщения
	// 	if err := s.storageService.SetNewMessageId(chatId, oldMessageId, messageId); err != nil {
	// 		s.log.Error("Ошибка сохранения нового Id сообщения", "err", err)
	// 	}
	// })
}

// deleteSystemMessage удаляет системное сообщение
func (s *Service) deleteSystemMessage(src *client.Message) error {
	source, ok := config.Engine.Sources[src.ChatId]
	if !ok {
		return nil
	}
	if !source.DeleteSystemMessages {
		return nil
	}
	if !s.messageService.IsSystemMessage(src) {
		return nil
	}
	_, err := s.telegramRepo.GetClient().DeleteMessages(&client.DeleteMessagesRequest{
		ChatId:     src.ChatId,
		MessageIds: []int64{src.Id},
		Revoke:     true,
	})
	return err
}

// isChatSource проверяет, является ли чат источником для какого-либо правила
func isChatSource(chatId int64) (map[string]entity.ForwardRule, bool) {
	// rules := make(map[string]entity.ForwardRule)

	// for ruleId, rule := range config.Engine.Forwards {
	// 	if rule.From == chatId && rule.Status == entity.RuleStatusActive {
	// 		rules[ruleId] = rule
	// 	}
	// }

	// return rules, len(rules) > 0
	return nil, false
}

// Обрабатывает медиа-альбом
func (s *Service) processMediaAlbum(forwardKey string, albumId client.JsonInt64) {
	// TODO: выполнить корректный перенос из budva32
	// TODO: правильно было переименовать из handleMediaAlbum?
}

// processMessage обрабатывает сообщения и выполняет пересылку согласно правилам
func (s *Service) processMessage(messages []*client.Message, forwardKey string, rule entity.ForwardRule) {
	// src := messages[0]
	// s.log.Debug("Обработка сообщения",
	// 	"chatId", src.ChatId,
	// 	"messageId", src.Id,
	// 	"albumId", src.MediaAlbumId,
	// 	"forwardKey", forwardKey)

	// // Начинаем пересылку
	// for _, dstChatId := range rule.To {
	// 	// Пересылаем сообщения
	// 	s.forwardMessages(messages, src.ChatId, dstChatId, rule.SendCopy, rule.CopyOnce, forwardKey)
	// }
}

// forwardMessages пересылает сообщения в целевой чат
func (s *Service) forwardMessages(messages []*client.Message, srcChatId, dstChatId int64, isSendCopy, isCopyOnce bool, forwardKey string) {
	// s.log.Debug("Пересылка сообщений",
	// 	"srcChatId", srcChatId,
	// 	"dstChatId", dstChatId,
	// 	"sendCopy", isSendCopy,
	// 	"copyOnce", isCopyOnce,
	// 	"messageCount", len(messages))

	// tdlibClient := s.telegramRepo.GetClient()

	// var (
	// 	result *client.Messages
	// 	err    error
	// )

	// // Метод пересылки в зависимости от флага SendCopy
	// if isSendCopy {
	// 	// Пересылка с созданием копии
	// 	if len(messages) == 1 {
	// 		// Пересылка одиночного сообщения
	// 		message, err := s.sendCopyMessage(tdlibClient, messages[0], dstChatId, forwardKey)
	// 		if err != nil {
	// 			s.log.Error("Ошибка пересылки сообщения", "err", err)
	// 			return
	// 		}

	// 		result = &client.Messages{
	// 			TotalCount: 1,
	// 			Messages:   []*client.Message{message},
	// 		}
	// 	} else {
	// 		// Пересылка медиа-альбома
	// 		result, err = s.sendCopyAlbum(tdlibClient, messages, dstChatId, forwardKey)
	// 		if err != nil {
	// 			s.log.Error("Ошибка пересылки медиа-альбома", "err", err)
	// 			return
	// 		}
	// 	}
	// } else {
	// 	// Прямая пересылка (forward)
	// 	messageIds := make([]int64, len(messages))
	// 	for i, message := range messages {
	// 		messageIds[i] = message.Id
	// 	}

	// 	result, err = tdlibClient.ForwardMessages(&client.ForwardMessagesRequest{
	// 		ChatId:     dstChatId,
	// 		FromChatId: srcChatId,
	// 		MessageIds: messageIds,
	// 	})

	// 	if err != nil {
	// 		s.log.Error("Ошибка форвардинга сообщений", "err", err)
	// 		return
	// 	}
	// }

	// // Сохраняем соответствие между оригинальными и пересланными сообщениями
	// if result != nil && len(result.Messages) > 0 {
	// 	for i, message := range result.Messages {
	// 		// В случае форвардинга медиа-альбома, сообщения могут прийти не в том порядке
	// 		srcMessage := messages[i]
	// 		if len(messages) > 1 {
	// 			// Для альбома ищем соответствующее исходное сообщение
	// 			for _, msg := range messages {
	// 				if s.matchesMediaContent(msg, message) {
	// 					srcMessage = msg
	// 					break
	// 				}
	// 			}
	// 		}

	// 		// Сохраняем связь между оригинальным и пересланным сообщениями
	// 		fromChatMessageId := fmt.Sprintf("%d:%d", srcChatId, srcMessage.Id)
	// 		toChatMessageId := fmt.Sprintf("%s:%d:%d", forwardKey, dstChatId, message.Id)

	// 		// Если это не разовое копирование (CopyOnce), сохраняем связь для последующего синхронизации при редактировании
	// 		if !isCopyOnce {
	// 			if err := s.storageService.SetCopiedMessageId(fromChatMessageId, toChatMessageId); err != nil {
	// 				s.log.Error("Ошибка сохранения связи сообщений", "err", err)
	// 			}
	// 		}

	// 		// Сохраняем временный Id сообщения для обработки асинхронных обновлений
	// 		if message.SendingState != nil {
	// 			if sendingState, ok := message.SendingState.(*client.MessageSendingStatePending); ok {
	// 				tmpMessageId := sendingState.SendingId

	// 				if err := s.storageService.SetTmpMessageId(dstChatId, message.Id, int64(tmpMessageId)); err != nil {
	// 					s.log.Error("Ошибка сохранения временного Id", "err", err)
	// 				}
	// 			}
	// 		}
	// 	}
	// }
}

// sendCopyMessage отправляет копию одиночного сообщения
func (s *Service) sendCopyMessage(tdlibClient *client.Client, message *client.Message, dstChatId int64, forwardKey string) (*client.Message, error) {
	// // Получаем текст сообщения
	// formattedText, contentType := s.messageService.GetContent(message)
	// if contentType == "" {
	// 	return nil, fmt.Errorf("неподдерживаемый тип сообщения")
	// }

	// // Применяем трансформации к тексту
	// if err := s.transformService.ReplaceMyselfLinks(formattedText, message.ChatId, dstChatId); err != nil {
	// 	s.log.Error("Ошибка при замене ссылок", "err", err)
	// }
	// if err := s.transformService.ReplaceFragments(formattedText, dstChatId); err != nil {
	// 	s.log.Error("Ошибка при замене фрагментов", "err", err)
	// }
	// if err := s.transformService.AddSources(formattedText, message, dstChatId); err != nil {
	// 	s.log.Error("Ошибка при добавлении источников", "err", err)
	// }

	// // Создаем входной контент для сообщения
	// var inputContent client.InputMessageContent

	// switch contentType {
	// case client.TypeMessageText:
	// 	inputContent = &client.InputMessageText{
	// 		Text: formattedText,
	// 	}
	// case client.TypeMessagePhoto:
	// 	content := message.Content.(*client.MessagePhoto)
	// 	inputContent = &client.InputMessagePhoto{
	// 		Photo: &client.InputFileRemote{
	// 			Id: content.Photo.Sizes[len(content.Photo.Sizes)-1].Photo.Remote.Id,
	// 		},
	// 		Caption: formattedText,
	// 	}
	// case client.TypeMessageVideo:
	// 	content := message.Content.(*client.MessageVideo)
	// 	inputContent = &client.InputMessageVideo{
	// 		Video: &client.InputFileRemote{
	// 			Id: content.Video.Video.Remote.Id,
	// 		},
	// 		Caption: formattedText,
	// 	}
	// case client.TypeMessageDocument:
	// 	content := message.Content.(*client.MessageDocument)
	// 	inputContent = &client.InputMessageDocument{
	// 		Document: &client.InputFileRemote{
	// 			Id: content.Document.Document.Remote.Id,
	// 		},
	// 		Caption: formattedText,
	// 	}
	// // TODO: перенести реализацию на остальные поддерживаемые типы
	// default:
	// 	return nil, fmt.Errorf("неподдерживаемый тип сообщения: %s", contentType)
	// }

	// // Отправляем сообщение
	// return tdlibClient.SendMessage(&client.SendMessageRequest{
	// 	ChatId:              dstChatId,
	// 	InputMessageContent: inputContent,
	// })
	return nil, nil
}

// sendCopyAlbum отправляет копию медиа-альбома
func (s *Service) sendCopyAlbum(tdlibClient *client.Client, messages []*client.Message, dstChatId int64, forwardKey string) (*client.Messages, error) {
	// contents := make([]client.InputMessageContent, 0, len(messages))

	// for i, message := range messages {
	// 	formattedText, contentType := s.messageService.GetContent(message)
	// 	if contentType == "" {
	// 		continue
	// 	}

	// 	// Применяем трансформации только к первому сообщению
	// 	if i == 0 {
	// 		if err := s.transformService.ReplaceMyselfLinks(formattedText, message.ChatId, dstChatId); err != nil {
	// 			s.log.Error("Ошибка при замене ссылок", "err", err)
	// 		}
	// 		if err := s.transformService.ReplaceFragments(formattedText, dstChatId); err != nil {
	// 			s.log.Error("Ошибка при замене фрагментов", "err", err)
	// 		}
	// 		if err := s.transformService.AddSources(formattedText, message, dstChatId); err != nil {
	// 			s.log.Error("Ошибка при добавлении источников", "err", err)
	// 		}
	// 	}

	// 	var inputContent client.InputMessageContent

	// 	switch contentType {
	// 	case client.TypeMessagePhoto:
	// 		content := message.Content.(*client.MessagePhoto)
	// 		inputContent = &client.InputMessagePhoto{
	// 			Photo: &client.InputFileRemote{
	// 				Id: content.Photo.Sizes[len(content.Photo.Sizes)-1].Photo.Remote.Id,
	// 			},
	// 			Caption: formattedText,
	// 		}
	// 	case client.TypeMessageVideo:
	// 		content := message.Content.(*client.MessageVideo)
	// 		inputContent = &client.InputMessageVideo{
	// 			Video: &client.InputFileRemote{
	// 				Id: content.Video.Video.Remote.Id,
	// 			},
	// 			Caption: formattedText,
	// 		}
	// 	default:
	// 		continue
	// 	}

	// 	contents = append(contents, inputContent)
	// }

	// if len(contents) == 0 {
	// 	return nil, fmt.Errorf("нет поддерживаемых типов контента в альбоме")
	// }

	// // Отправляем альбом
	// return tdlibClient.SendMessageAlbum(&client.SendMessageAlbumRequest{
	// 	ChatId:               dstChatId,
	// 	InputMessageContents: contents,
	// })
	return nil, nil
}

// matchesMediaContent проверяет, соответствует ли содержимое двух сообщений
func (s *Service) matchesMediaContent(src, dst *client.Message) bool {
	// srcFormattedText, srcContentType := s.messageService.GetContent(src)
	// dstFormattedText, dstContentType := s.messageService.GetContent(dst)

	// if srcContentType == "" || srcContentType != dstContentType {
	// 	return false
	// }

	// return srcFormattedText.Text == dstFormattedText.Text
	return false
}

// processSingleEdited обрабатывает редактирование сообщения
func (s *Service) processSingleEdited(message *client.Message, toChatMessageId string) {
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

// processSingleDeleted обрабатывает удаление сообщения
func (s *Service) processSingleDeleted(fromChatMessageId, toChatMessageId string) {
	// // Разбираем toChatMessageId
	// ruleId, dstChatId, dstMessageId, err := parseToChatMessageId(toChatMessageId)
	// if err != nil {
	// 	s.log.Error("Ошибка разбора toChatMessageId", "toChatMessageId", toChatMessageId, "err", err)
	// 	return
	// }

	// s.log.Debug("Обработка удаления сообщения",
	// 	"fromChatMessageId", fromChatMessageId,
	// 	"dstChatId", dstChatId,
	// 	"dstMessageId", dstMessageId,
	// 	"ruleId", ruleId)

	// // Получаем правило форвардинга
	// rule, ok := config.Engine.Forwards[ruleId]
	// if !ok {
	// 	s.log.Error("Правило форвардинга не найдено", "ruleId", ruleId)
	// 	return
	// }

	// // Если установлен флаг Indelible, не удаляем сообщение
	// if rule.Indelible {
	// 	s.log.Debug("Сообщение неудаляемое (Indelible), удаление не выполняется",
	// 		"ruleId", ruleId,
	// 		"dstChatId", dstChatId)
	// 	return
	// }

	// tdlibClient := s.telegramRepo.GetClient()

	// // Удаляем сообщение
	// _, err = tdlibClient.DeleteMessages(&client.DeleteMessagesRequest{
	// 	ChatId:     dstChatId,
	// 	MessageIds: []int64{dstMessageId},
	// 	Revoke:     true, // Удаляем для всех участников, а не только для себя
	// })
	// if err != nil {
	// 	s.log.Error("Ошибка удаления сообщения", "err", err)
	// } else {
	// 	s.log.Debug("Сообщение успешно удалено",
	// 		"dstChatId", dstChatId,
	// 		"dstMessageId", dstMessageId)

	// 	// Удаляем соответствие между временным и постоянным Id
	// 	tmpMessageId, err := s.storageService.GetTmpMessageId(dstChatId, dstMessageId)
	// 	if err == nil && tmpMessageId != 0 {
	// 		err = s.storageService.DeleteTmpMessageId(dstChatId, dstMessageId)
	// 		if err != nil {
	// 			s.log.Error("Ошибка удаления временного Id сообщения", "err", err)
	// 		}
	// 		err = s.storageService.DeleteNewMessageId(dstChatId, tmpMessageId)
	// 		if err != nil {
	// 			s.log.Error("Ошибка удаления постоянного Id сообщения", "err", err)
	// 		}
	// 	}
	// }
}

var forwardedToMu sync.Mutex

// isNotForwardedTo проверяет, было ли сообщение уже отправлено в данный чат
func isNotForwardedTo(forwardedTo map[int64]bool, dstChatId int64) bool {
	forwardedToMu.Lock()
	defer forwardedToMu.Unlock()
	if !forwardedTo[dstChatId] {
		forwardedTo[dstChatId] = true
		return true
	}
	return false
}

type filtersMode = string

const (
	filtersOK    filtersMode = "ok"
	filtersCheck filtersMode = "check"
	filtersOther filtersMode = "other"
)

// mapFiltersMode определяет, какой режим фильтрации применим
func mapFiltersMode(formattedText *client.FormattedText, rule entity.ForwardRule) filtersMode {
	if formattedText.Text == "" {
		hasInclude := false
		if rule.Include != "" {
			hasInclude = true
		}
		for _, includeSubmatch := range rule.IncludeSubmatch {
			if includeSubmatch.Regexp != "" {
				hasInclude = true
				break
			}
		}
		if hasInclude {
			return filtersOther
		}
	} else {
		if rule.Exclude != "" {
			re := regexp.MustCompile("(?i)" + rule.Exclude)
			if re.FindString(formattedText.Text) != "" {
				return filtersCheck
			}
		}
		hasInclude := false
		if rule.Include != "" {
			hasInclude = true
			re := regexp.MustCompile("(?i)" + rule.Include)
			if re.FindString(formattedText.Text) != "" {
				return filtersOK
			}
		}
		for _, includeSubmatch := range rule.IncludeSubmatch {
			if includeSubmatch.Regexp != "" {
				hasInclude = true
				re := regexp.MustCompile("(?i)" + includeSubmatch.Regexp)
				matches := re.FindAllStringSubmatch(formattedText.Text, -1)
				for _, match := range matches {
					s := match[includeSubmatch.Group]
					if slices.Contains(includeSubmatch.Match, s) {
						return filtersOK
					}
				}
			}
		}
		if hasInclude {
			return filtersOther
		}
	}
	return filtersOK
}

// parseToChatMessageId разбирает строку toChatMessageId в формате "ruleId:chatId:messageId"
func parseToChatMessageId(toChatMessageId string) (ruleId string, chatId int64, messageId int64, err error) {
	// parts := strings.Split(toChatMessageId, ":")
	// if len(parts) != 3 {
	// 	return "", 0, 0, fmt.Errorf("неверный формат toChatMessageId: %s", toChatMessageId)
	// }

	// ruleId = parts[0]

	// var chatIdInt, messageIdInt int
	// if _, err := fmt.Sscanf(parts[1], "%d", &chatIdInt); err != nil {
	// 	return "", 0, 0, fmt.Errorf("ошибка преобразования chatId: %w", err)
	// }

	// if _, err := fmt.Sscanf(parts[2], "%d", &messageIdInt); err != nil {
	// 	return "", 0, 0, fmt.Errorf("ошибка преобразования messageId: %w", err)
	// }

	// return ruleId, int64(chatIdInt), int64(messageIdInt), nil
	return "", 0, 0, nil
}
