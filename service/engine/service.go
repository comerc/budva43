package engine

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/zelenin/go-tdlib/client"

	"github.com/comerc/budva43/config"
	"github.com/comerc/budva43/entity"
	"github.com/comerc/budva43/util"
)

// TODO: выполнить корректный перенос из budva32 (нужно вернуть функционал старой версии):
// - Функции пересылки сообщений (`forwardMessages`, `sendCopyMessage`, `sendCopyAlbum`) имеют более формализованный интерфейс, но могут не полностью воспроизводить логику старой версии
// - Обработка редактированных сообщений (`processSingleEdited`) отличается от старой реализации, что может приводить к различиям в поведении
// - Функция `matchesMediaContent` для проверки соответствия медиа-контента отсутствовала в старой версии и может быть избыточной

type queueService interface {
	Add(task func())
}

type messageService interface {
	GetContent(message *client.Message) (*client.FormattedText, string)
	IsSystemMessage(message *client.Message) bool
	// GetContentType(message *client.Message) string
	// SendMessage(chatID int64, text string) (*client.Message, error)
	// ForwardMessage(fromChatID, messageID, toChatID int64) (*client.Message, error)
	// SendMessageAlbum(chatID int64, contents []client.InputMessageContent) (*client.Messages, error)
	// ForwardMessages(fromChatID int64, messageIDs []int64, toChatID int64) (*client.Messages, error)
	// EditMessageText(chatID, messageID int64, text *client.FormattedText) (*client.Message, error)
	// EditMessageMedia(chatID, messageID int64, content client.InputMessageContent) (*client.Message, error)
	// EditMessageCaption(chatID, messageID int64, caption *client.FormattedText) (*client.Message, error)
	// DeleteMessages(chatID int64, messageIDs []int64) error
	// GetMessage(chatID, messageID int64) (*client.Message, error)
}

type filterService interface {
	ShouldForward(text string, rule *entity.ForwardRule) (bool, error)
}

type transformService interface {
	ReplaceMyselfLinks(formattedText *client.FormattedText, srcChatID, dstChatID int64) error
	ReplaceFragments(formattedText *client.FormattedText, dstChatID int64) error
	AddSources(formattedText *client.FormattedText, message *client.Message, dstChatId int64) error
}

type storageService interface {
	SetCopiedMessageID(fromChatMessageID string, toChatMessageID string) error
	GetCopiedMessageIDs(fromChatMessageID string) ([]string, error)
	DeleteCopiedMessageIDs(fromChatMessageID string) error
	SetNewMessageID(chatID, tmpMessageID, newMessageID int64) error
	// GetNewMessageID(chatID, tmpMessageID int64) (int64, error)
	DeleteNewMessageID(chatID, tmpMessageID int64) error
	SetTmpMessageID(chatID, newMessageID, tmpMessageID int64) error
	GetTmpMessageID(chatID, newMessageID int64) (int64, error)
	DeleteTmpMessageID(chatID, newMessageID int64) error
	// IncrementViewedMessages(toChatID int64) error
	// IncrementForwardedMessages(toChatID int64) error
}

type mediaAlbumService interface {
	AddMessage(forwardKey string, message *client.Message) bool
	GetLastReceivedDiff(key string) time.Duration
	GetMessages(key string) []*client.Message
}

type telegramRepo interface {
	GetClient() *client.Client
	AuthClientDone() chan any
}

// Service предоставляет функциональность движка пересылки сообщений
type Service struct {
	log *slog.Logger
	//
	queueService       queueService
	messageService     messageService
	filterService      filterService
	transformService   transformService
	storageService     storageService
	mediaAlbumsService mediaAlbumService
	telegramRepo       telegramRepo
}

// New создает новый экземпляр сервиса engine
func New(
	queueService queueService,
	messageService messageService,
	filterService filterService,
	transformService transformService,
	storageService storageService,
	mediaAlbumsService mediaAlbumService,
	telegramRepo telegramRepo,
) *Service {
	return &Service{
		log: slog.With("module", "service.engine"),
		//
		queueService:       queueService,
		messageService:     messageService,
		filterService:      filterService,
		transformService:   transformService,
		storageService:     storageService,
		mediaAlbumsService: mediaAlbumsService,
		telegramRepo:       telegramRepo,
	}
}

// Start запускает обработчик обновлений от Telegram
func (s *Service) Start(ctx context.Context) error {
	// return nil

	s.log.Info("Запуск сервиса engine")

	// Проверяем конфигурацию
	if err := s.validateConfig(); err != nil {
		return fmt.Errorf("ошибка валидации конфигурации: %w", err)
	}

	go func() {
		// Ждём авторизации клиента
		select {
		case <-s.telegramRepo.AuthClientDone():
		case <-ctx.Done():
			return
		}
		// Получаем канал обновлений от Telegram
		listener := s.telegramRepo.GetClient().GetListener()
		s.handleUpdates(ctx, listener)
	}()

	return nil
}

// Close останавливает сервис
func (s *Service) Close() error {
	return nil
}

// Проверяет корректность конфигурации
func (s *Service) validateConfig() error {
	// Проверяем заменяемые фрагменты текста
	for chatID, settings := range config.Engine.ReplaceFragments {
		for from, to := range settings.Replacements {
			if util.RuneCountForUTF16(from) != util.RuneCountForUTF16(to) {
				return fmt.Errorf("длина исходного и заменяемого текста должна быть одинаковой: %s -> %s", from, to)
			}
		}
		s.log.Info("Валидированы настройки замены фрагментов", "chatID", chatID, "replacements", len(settings.Replacements))
	}

	// Проверяем правила пересылки
	for ruleID, rule := range config.Engine.Forwards {
		for _, dstChatID := range rule.To {
			if rule.From == dstChatID {
				return fmt.Errorf("идентификатор получателя не может совпадать с идентификатором источника: %d", dstChatID)
			}
		}
		s.log.Info("Валидировано правило пересылки", "ruleID", ruleID, "from", rule.From, "to", rule.To)
	}

	return nil
}

// handleUpdates обрабатывает обновления от Telegram
func (s *Service) handleUpdates(ctx context.Context, listener *client.Listener) {
	for {
		select {
		case <-ctx.Done():
			listener.Close()
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
	message := update.Message
	chatID := message.ChatId

	// Проверяем, является ли чат источником для какого-либо правила
	isSourceChat := false
	var forwardRules []struct {
		ruleID string
		rule   entity.ForwardRule
	}

	for ruleID, rule := range config.Engine.Forwards {
		if rule.From == chatID && rule.Status == entity.RuleStatusActive {
			isSourceChat = true
			forwardRules = append(forwardRules, struct {
				ruleID string
				rule   entity.ForwardRule
			}{ruleID, rule})
		}
	}

	if !isSourceChat {
		return
	}

	// Проверяем удаление системных сообщений
	if shouldDelete := config.Engine.DeleteSystemMessages[chatID]; shouldDelete {
		if s.messageService.IsSystemMessage(message) {
			go func() {
				tdlibClient := s.telegramRepo.GetClient()
				_, err := tdlibClient.DeleteMessages(&client.DeleteMessagesRequest{
					ChatId:     chatID,
					MessageIds: []int64{message.Id},
					Revoke:     true, // Удаляем для всех участников, а не только для себя
				})
				if err != nil {
					s.log.Error("Ошибка удаления системного сообщения", "err", err)
				}
			}()
		}
	}

	formattedText, _ := s.messageService.GetContent(message)

	// Обрабатываем каждое правило
	for _, ruleData := range forwardRules {
		// Проверяем правило
		rule := ruleData.rule

		// Проверяем, должно ли сообщение быть переслано согласно фильтрам
		shouldForward, err := s.filterService.ShouldForward(formattedText.Text, &rule)
		if err != nil {
			s.log.Error("Ошибка проверки фильтров", "err", err)
			continue
		}

		if !shouldForward {
			s.log.Debug("Сообщение не проходит фильтры", "ruleID", ruleData.ruleID)
			continue
		}

		// Обрабатываем сообщение в зависимости от типа
		if message.MediaAlbumId == 0 {
			// Одиночное сообщение
			s.queueService.Add(func() {
				s.processMessage([]*client.Message{message}, ruleData.ruleID, rule)
			})
		} else {
			// Медиа-альбом
			isFirstMessage := s.mediaAlbumsService.AddMessage(ruleData.ruleID, message)
			if isFirstMessage {
				s.queueService.Add(func() {
					s.processMediaAlbum(ruleData.ruleID, message.MediaAlbumId)
				})
			}
		}
	}
}

// handleUpdateMessageEdited обрабатывает обновление о редактировании сообщения
func (s *Service) handleUpdateMessageEdited(update *client.UpdateMessageEdited) {
	chatID := update.ChatId
	messageID := update.MessageId

	// Проверяем, является ли чат источником для какого-либо правила
	if _, ok := isChatSource(chatID); !ok {
		return
	}

	s.log.Debug("Обработка редактирования сообщения", "chatID", chatID, "messageID", messageID)

	// Отправляем задачу в очередь
	s.queueService.Add(func() {
		// Формируем ключ для поиска скопированных сообщений
		fromChatMessageID := fmt.Sprintf("%d:%d", chatID, messageID)

		// Получаем идентификаторы скопированных сообщений
		toChatMessageIDs, err := s.storageService.GetCopiedMessageIDs(fromChatMessageID)
		if err != nil {
			s.log.Error("Ошибка получения скопированных сообщений", "err", err)
			return
		}

		if len(toChatMessageIDs) == 0 {
			s.log.Debug("Скопированные сообщения не найдены", "fromChatMessageID", fromChatMessageID)
			return
		}

		tdlibClient := s.telegramRepo.GetClient()

		// Получаем исходное сообщение
		src, err := tdlibClient.GetMessage(&client.GetMessageRequest{
			ChatId:    chatID,
			MessageId: messageID,
		})
		if err != nil {
			s.log.Error("Ошибка получения исходного сообщения", "err", err)
			return
		}

		// Обрабатываем каждое скопированное сообщение
		for _, toChatMessageID := range toChatMessageIDs {
			s.processSingleEdited(src, toChatMessageID)
		}
	})
}

// handleUpdateDeleteMessages обрабатывает обновление об удалении сообщений
func (s *Service) handleUpdateDeleteMessages(update *client.UpdateDeleteMessages) {
	// Обрабатываем только постоянное удаление сообщений
	if !update.IsPermanent {
		return
	}

	chatID := update.ChatId
	messageIDs := update.MessageIds

	// Проверяем, является ли чат источником для какого-либо правила
	if _, ok := isChatSource(chatID); !ok {
		return
	}

	s.log.Debug("Обработка удаления сообщений", "chatID", chatID, "messageIDs", messageIDs)

	// Отправляем задачу в очередь
	s.queueService.Add(func() {
		// Обрабатываем каждое удаленное сообщение
		for _, messageID := range messageIDs {
			// Формируем ключ для поиска скопированных сообщений
			fromChatMessageID := fmt.Sprintf("%d:%d", chatID, messageID)

			// Получаем идентификаторы скопированных сообщений
			toChatMessageIDs, err := s.storageService.GetCopiedMessageIDs(fromChatMessageID)
			if err != nil {
				s.log.Error("Ошибка получения скопированных сообщений", "err", err)
				continue
			}

			if len(toChatMessageIDs) == 0 {
				continue
			}

			// Обрабатываем каждое скопированное сообщение
			for _, toChatMessageID := range toChatMessageIDs {
				s.processSingleDeleted(fromChatMessageID, toChatMessageID)
			}

			// Удаляем соответствие между оригинальным и скопированными сообщениями
			err = s.storageService.DeleteCopiedMessageIDs(fromChatMessageID)
			if err != nil {
				s.log.Error("Ошибка удаления скопированных сообщений", "err", err)
			}
		}
	})
}

// handleUpdateMessageSendSucceeded обрабатывает обновление об успешной отправке сообщения
func (s *Service) handleUpdateMessageSendSucceeded(update *client.UpdateMessageSendSucceeded) {
	message := update.Message
	chatID := message.ChatId
	messageID := message.Id
	oldMessageID := update.OldMessageId

	s.log.Debug("Обработка успешной отправки сообщения",
		"chatID", chatID,
		"messageID", messageID,
		"oldMessageID", oldMessageID)

	// Отправляем задачу в очередь
	s.queueService.Add(func() {
		// Сохраняем соответствие между временным и постоянным ID сообщения
		if err := s.storageService.SetNewMessageID(chatID, oldMessageID, messageID); err != nil {
			s.log.Error("Ошибка сохранения нового ID сообщения", "err", err)
		}
	})
}

// isChatSource проверяет, является ли чат источником для какого-либо правила
func isChatSource(chatID int64) (map[string]entity.ForwardRule, bool) {
	rules := make(map[string]entity.ForwardRule)

	for ruleID, rule := range config.Engine.Forwards {
		if rule.From == chatID && rule.Status == entity.RuleStatusActive {
			rules[ruleID] = rule
		}
	}

	return rules, len(rules) > 0
}

// Обрабатывает медиа-альбом
func (s *Service) processMediaAlbum(forwardKey string, albumID client.JsonInt64) {
	// TODO: выполнить корректный перенос из budva32
	// TODO: правильно было переименовать из handleMediaAlbum?
}

// processMessage обрабатывает сообщения и выполняет пересылку согласно правилам
func (s *Service) processMessage(messages []*client.Message, forwardKey string, rule entity.ForwardRule) {
	src := messages[0]
	s.log.Debug("Обработка сообщения",
		"chatID", src.ChatId,
		"messageID", src.Id,
		"albumID", src.MediaAlbumId,
		"forwardKey", forwardKey)

	// Начинаем пересылку
	for _, dstChatID := range rule.To {
		// Пересылаем сообщения
		s.forwardMessages(messages, src.ChatId, dstChatID, rule.SendCopy, rule.CopyOnce, forwardKey)
	}
}

// forwardMessages пересылает сообщения в целевой чат
func (s *Service) forwardMessages(messages []*client.Message, srcChatID, dstChatID int64, isSendCopy, isCopyOnce bool, forwardKey string) {
	s.log.Debug("Пересылка сообщений",
		"srcChatID", srcChatID,
		"dstChatID", dstChatID,
		"sendCopy", isSendCopy,
		"copyOnce", isCopyOnce,
		"messageCount", len(messages))

	tdlibClient := s.telegramRepo.GetClient()

	var (
		result *client.Messages
		err    error
	)

	// Метод пересылки в зависимости от флага SendCopy
	if isSendCopy {
		// Пересылка с созданием копии
		if len(messages) == 1 {
			// Пересылка одиночного сообщения
			message, err := s.sendCopyMessage(tdlibClient, messages[0], dstChatID, forwardKey)
			if err != nil {
				s.log.Error("Ошибка пересылки сообщения", "err", err)
				return
			}

			result = &client.Messages{
				TotalCount: 1,
				Messages:   []*client.Message{message},
			}
		} else {
			// Пересылка медиа-альбома
			result, err = s.sendCopyAlbum(tdlibClient, messages, dstChatID, forwardKey)
			if err != nil {
				s.log.Error("Ошибка пересылки медиа-альбома", "err", err)
				return
			}
		}
	} else {
		// Прямая пересылка (forward)
		messageIDs := make([]int64, len(messages))
		for i, message := range messages {
			messageIDs[i] = message.Id
		}

		result, err = tdlibClient.ForwardMessages(&client.ForwardMessagesRequest{
			ChatId:     dstChatID,
			FromChatId: srcChatID,
			MessageIds: messageIDs,
		})

		if err != nil {
			s.log.Error("Ошибка форвардинга сообщений", "err", err)
			return
		}
	}

	// Сохраняем соответствие между оригинальными и пересланными сообщениями
	if result != nil && len(result.Messages) > 0 {
		for i, message := range result.Messages {
			// В случае форвардинга медиа-альбома, сообщения могут прийти не в том порядке
			srcMessage := messages[i]
			if len(messages) > 1 {
				// Для альбома ищем соответствующее исходное сообщение
				for _, msg := range messages {
					if s.matchesMediaContent(msg, message) {
						srcMessage = msg
						break
					}
				}
			}

			// Сохраняем связь между оригинальным и пересланным сообщениями
			fromChatMessageID := fmt.Sprintf("%d:%d", srcChatID, srcMessage.Id)
			toChatMessageID := fmt.Sprintf("%s:%d:%d", forwardKey, dstChatID, message.Id)

			// Если это не разовое копирование (CopyOnce), сохраняем связь для последующего синхронизации при редактировании
			if !isCopyOnce {
				if err := s.storageService.SetCopiedMessageID(fromChatMessageID, toChatMessageID); err != nil {
					s.log.Error("Ошибка сохранения связи сообщений", "err", err)
				}
			}

			// Сохраняем временный ID сообщения для обработки асинхронных обновлений
			if message.SendingState != nil {
				if sendingState, ok := message.SendingState.(*client.MessageSendingStatePending); ok {
					tmpMessageID := sendingState.SendingId

					if err := s.storageService.SetTmpMessageID(dstChatID, message.Id, int64(tmpMessageID)); err != nil {
						s.log.Error("Ошибка сохранения временного ID", "err", err)
					}
				}
			}
		}
	}
}

// sendCopyMessage отправляет копию одиночного сообщения
func (s *Service) sendCopyMessage(tdlibClient *client.Client, message *client.Message, dstChatID int64, forwardKey string) (*client.Message, error) {
	// Получаем текст сообщения
	formattedText, contentType := s.messageService.GetContent(message)
	if contentType == "" {
		return nil, fmt.Errorf("неподдерживаемый тип сообщения")
	}

	// Применяем трансформации к тексту
	if err := s.transformService.ReplaceMyselfLinks(formattedText, message.ChatId, dstChatID); err != nil {
		s.log.Error("Ошибка при замене ссылок", "err", err)
	}
	if err := s.transformService.ReplaceFragments(formattedText, dstChatID); err != nil {
		s.log.Error("Ошибка при замене фрагментов", "err", err)
	}
	if err := s.transformService.AddSources(formattedText, message, dstChatID); err != nil {
		s.log.Error("Ошибка при добавлении источников", "err", err)
	}

	// Создаем входной контент для сообщения
	var inputContent client.InputMessageContent

	switch contentType {
	case client.TypeMessageText:
		inputContent = &client.InputMessageText{
			Text: formattedText,
		}
	case client.TypeMessagePhoto:
		content := message.Content.(*client.MessagePhoto)
		inputContent = &client.InputMessagePhoto{
			Photo: &client.InputFileRemote{
				Id: content.Photo.Sizes[len(content.Photo.Sizes)-1].Photo.Remote.Id,
			},
			Caption: formattedText,
		}
	case client.TypeMessageVideo:
		content := message.Content.(*client.MessageVideo)
		inputContent = &client.InputMessageVideo{
			Video: &client.InputFileRemote{
				Id: content.Video.Video.Remote.Id,
			},
			Caption: formattedText,
		}
	case client.TypeMessageDocument:
		content := message.Content.(*client.MessageDocument)
		inputContent = &client.InputMessageDocument{
			Document: &client.InputFileRemote{
				Id: content.Document.Document.Remote.Id,
			},
			Caption: formattedText,
		}
	// TODO: перенести реализацию на остальные поддерживаемые типы
	default:
		return nil, fmt.Errorf("неподдерживаемый тип сообщения: %s", contentType)
	}

	// Отправляем сообщение
	return tdlibClient.SendMessage(&client.SendMessageRequest{
		ChatId:              dstChatID,
		InputMessageContent: inputContent,
	})
}

// sendCopyAlbum отправляет копию медиа-альбома
func (s *Service) sendCopyAlbum(tdlibClient *client.Client, messages []*client.Message, dstChatID int64, forwardKey string) (*client.Messages, error) {
	contents := make([]client.InputMessageContent, 0, len(messages))

	for i, message := range messages {
		formattedText, contentType := s.messageService.GetContent(message)
		if contentType == "" {
			continue
		}

		// Применяем трансформации только к первому сообщению
		if i == 0 {
			if err := s.transformService.ReplaceMyselfLinks(formattedText, message.ChatId, dstChatID); err != nil {
				s.log.Error("Ошибка при замене ссылок", "err", err)
			}
			if err := s.transformService.ReplaceFragments(formattedText, dstChatID); err != nil {
				s.log.Error("Ошибка при замене фрагментов", "err", err)
			}
			if err := s.transformService.AddSources(formattedText, message, dstChatID); err != nil {
				s.log.Error("Ошибка при добавлении источников", "err", err)
			}
		}

		var inputContent client.InputMessageContent

		switch contentType {
		case client.TypeMessagePhoto:
			content := message.Content.(*client.MessagePhoto)
			inputContent = &client.InputMessagePhoto{
				Photo: &client.InputFileRemote{
					Id: content.Photo.Sizes[len(content.Photo.Sizes)-1].Photo.Remote.Id,
				},
				Caption: formattedText,
			}
		case client.TypeMessageVideo:
			content := message.Content.(*client.MessageVideo)
			inputContent = &client.InputMessageVideo{
				Video: &client.InputFileRemote{
					Id: content.Video.Video.Remote.Id,
				},
				Caption: formattedText,
			}
		default:
			continue
		}

		contents = append(contents, inputContent)
	}

	if len(contents) == 0 {
		return nil, fmt.Errorf("нет поддерживаемых типов контента в альбоме")
	}

	// Отправляем альбом
	return tdlibClient.SendMessageAlbum(&client.SendMessageAlbumRequest{
		ChatId:               dstChatID,
		InputMessageContents: contents,
	})
}

// matchesMediaContent проверяет, соответствует ли содержимое двух сообщений
func (s *Service) matchesMediaContent(src, dst *client.Message) bool {
	srcFormattedText, srcContentType := s.messageService.GetContent(src)
	dstFormattedText, dstContentType := s.messageService.GetContent(dst)

	if srcContentType == "" || srcContentType != dstContentType {
		return false
	}

	return srcFormattedText.Text == dstFormattedText.Text
}

// processSingleEdited обрабатывает редактирование сообщения
func (s *Service) processSingleEdited(message *client.Message, toChatMessageID string) {
	// Разбираем toChatMessageID
	ruleID, dstChatID, dstMessageID, err := parseToChatMessageID(toChatMessageID)
	if err != nil {
		s.log.Error("Ошибка разбора toChatMessageID", "toChatMessageID", toChatMessageID, "err", err)
		return
	}

	s.log.Debug("Обработка редактирования сообщения",
		"srcChatID", message.ChatId,
		"srcMessageID", message.Id,
		"dstChatID", dstChatID,
		"dstMessageID", dstMessageID,
		"ruleID", ruleID)

	// Получаем правило форвардинга
	rule, ok := config.Engine.Forwards[ruleID]
	if !ok {
		s.log.Error("Правило форвардинга не найдено", "ruleID", ruleID)
		return
	}

	// Если установлен флаг CopyOnce, не обрабатываем редактирование
	if rule.CopyOnce {
		s.log.Debug("Сообщение скопировано однократно, редактирование не применяется",
			"ruleID", ruleID,
			"dstChatID", dstChatID)
		return
	}

	tdlibClient := s.telegramRepo.GetClient()

	// Получаем оригинальное сообщение
	srcMessage, err := tdlibClient.GetMessage(&client.GetMessageRequest{
		ChatId:    message.ChatId,
		MessageId: message.Id,
	})
	if err != nil {
		s.log.Error("Ошибка получения исходного сообщения", "err", err)
		return
	}

	// Получаем контент сообщения
	formattedText, contentType := s.messageService.GetContent(srcMessage)
	if contentType == "" {
		s.log.Error("Неподдерживаемый тип сообщения при редактировании")
		return
	}

	// Применяем трансформации к тексту
	if err := s.transformService.ReplaceMyselfLinks(formattedText, srcMessage.ChatId, dstChatID); err != nil {
		s.log.Error("Ошибка при замене ссылок", "err", err)
	}
	if err := s.transformService.ReplaceFragments(formattedText, dstChatID); err != nil {
		s.log.Error("Ошибка при замене фрагментов", "err", err)
	}

	// В зависимости от типа контента, применяем соответствующее редактирование
	switch contentType {
	case client.TypeMessageText:
		// Редактирование текста
		_, err = tdlibClient.EditMessageText(&client.EditMessageTextRequest{
			ChatId:    dstChatID,
			MessageId: dstMessageID,
			InputMessageContent: &client.InputMessageText{
				Text: formattedText,
			},
		})
	case client.TypeMessagePhoto:
		// TODO: почему только тут применяется getInputMessageContent() ?
		content := getInputMessageContent(srcMessage.Content, formattedText, contentType)
		_, err = tdlibClient.EditMessageMedia(&client.EditMessageMediaRequest{
			ChatId:              dstChatID,
			MessageId:           dstMessageID,
			InputMessageContent: content,
		})
	case client.TypeMessageVideo, client.TypeMessageDocument, client.TypeMessageAudio, client.TypeMessageAnimation:
		// TODO: реализовать?
	case client.TypeMessageVoiceNote:
		// Редактирование подписи медиа
		_, err = tdlibClient.EditMessageCaption(&client.EditMessageCaptionRequest{
			ChatId:    dstChatID,
			MessageId: dstMessageID,
			Caption:   formattedText,
		})
	default:
		err = fmt.Errorf("неподдерживаемый тип контента: %s", contentType)
	}

	if err != nil {
		s.log.Error("Ошибка редактирования сообщения", "err", err)
	} else {
		s.log.Debug("Сообщение успешно отредактировано",
			"dstChatID", dstChatID,
			"dstMessageID", dstMessageID)
	}
}

// processSingleDeleted обрабатывает удаление сообщения
func (s *Service) processSingleDeleted(fromChatMessageID, toChatMessageID string) {
	// Разбираем toChatMessageID
	ruleID, dstChatID, dstMessageID, err := parseToChatMessageID(toChatMessageID)
	if err != nil {
		s.log.Error("Ошибка разбора toChatMessageID", "toChatMessageID", toChatMessageID, "err", err)
		return
	}

	s.log.Debug("Обработка удаления сообщения",
		"fromChatMessageID", fromChatMessageID,
		"dstChatID", dstChatID,
		"dstMessageID", dstMessageID,
		"ruleID", ruleID)

	// Получаем правило форвардинга
	rule, ok := config.Engine.Forwards[ruleID]
	if !ok {
		s.log.Error("Правило форвардинга не найдено", "ruleID", ruleID)
		return
	}

	// Если установлен флаг Indelible, не удаляем сообщение
	if rule.Indelible {
		s.log.Debug("Сообщение неудаляемое (Indelible), удаление не выполняется",
			"ruleID", ruleID,
			"dstChatID", dstChatID)
		return
	}

	tdlibClient := s.telegramRepo.GetClient()

	// Удаляем сообщение
	_, err = tdlibClient.DeleteMessages(&client.DeleteMessagesRequest{
		ChatId:     dstChatID,
		MessageIds: []int64{dstMessageID},
		Revoke:     true, // Удаляем для всех участников, а не только для себя
	})
	if err != nil {
		s.log.Error("Ошибка удаления сообщения", "err", err)
	} else {
		s.log.Debug("Сообщение успешно удалено",
			"dstChatID", dstChatID,
			"dstMessageID", dstMessageID)

		// Удаляем соответствие между временным и постоянным ID
		tmpMessageID, err := s.storageService.GetTmpMessageID(dstChatID, dstMessageID)
		if err == nil && tmpMessageID != 0 {
			err = s.storageService.DeleteTmpMessageID(dstChatID, dstMessageID)
			if err != nil {
				s.log.Error("Ошибка удаления временного ID сообщения", "err", err)
			}
			err = s.storageService.DeleteNewMessageID(dstChatID, tmpMessageID)
			if err != nil {
				s.log.Error("Ошибка удаления постоянного ID сообщения", "err", err)
			}
		}
	}
}

// parseToChatMessageID разбирает строку toChatMessageID в формате "ruleID:chatID:messageID"
func parseToChatMessageID(toChatMessageID string) (ruleID string, chatID int64, messageID int64, err error) {
	parts := strings.Split(toChatMessageID, ":")
	if len(parts) != 3 {
		return "", 0, 0, fmt.Errorf("неверный формат toChatMessageID: %s", toChatMessageID)
	}

	ruleID = parts[0]

	var chatIDInt, messageIDInt int
	if _, err := fmt.Sscanf(parts[1], "%d", &chatIDInt); err != nil {
		return "", 0, 0, fmt.Errorf("ошибка преобразования chatID: %w", err)
	}

	if _, err := fmt.Sscanf(parts[2], "%d", &messageIDInt); err != nil {
		return "", 0, 0, fmt.Errorf("ошибка преобразования messageID: %w", err)
	}

	return ruleID, int64(chatIDInt), int64(messageIDInt), nil
}

// TODO: ?? перенести в service/message/service.go ??
func getInputMessageContent(messageContent client.MessageContent, formattedText *client.FormattedText, contentType string) client.InputMessageContent {
	switch contentType {
	case client.TypeMessageText:
		// TODO: messageText := messageContent.(*client.MessageText) // при переносе budva32 не работает?
		return &client.InputMessageText{
			Text: formattedText,
			// TODO: DisableWebPagePreview: messageText.WebPage == nil || messageText.WebPage.Url == "", // при переносе budva32 теперь не работает?
			ClearDraft: true,
		}
	case client.TypeMessageAnimation:
		messageAnimation := messageContent.(*client.MessageAnimation)
		return &client.InputMessageAnimation{
			Animation: &client.InputFileRemote{
				Id: messageAnimation.Animation.Animation.Remote.Id,
			},
			// TODO: AddedStickerFileIds , // if applicable?
			Duration: messageAnimation.Animation.Duration,
			Width:    messageAnimation.Animation.Width,
			Height:   messageAnimation.Animation.Height,
			Caption:  formattedText,
		}
	case client.TypeMessageAudio:
		messageAudio := messageContent.(*client.MessageAudio)
		return &client.InputMessageAudio{
			Audio: &client.InputFileRemote{
				Id: messageAudio.Audio.Audio.Remote.Id,
			},
			AlbumCoverThumbnail: getInputThumbnail(messageAudio.Audio.AlbumCoverThumbnail),
			Title:               messageAudio.Audio.Title,
			Duration:            messageAudio.Audio.Duration,
			Performer:           messageAudio.Audio.Performer,
			Caption:             formattedText,
		}
	case client.TypeMessageDocument:
		messageDocument := messageContent.(*client.MessageDocument)
		return &client.InputMessageDocument{
			Document: &client.InputFileRemote{
				Id: messageDocument.Document.Document.Remote.Id,
			},
			Thumbnail: getInputThumbnail(messageDocument.Document.Thumbnail),
			Caption:   formattedText,
		}
	case client.TypeMessagePhoto:
		messagePhoto := messageContent.(*client.MessagePhoto)
		return &client.InputMessagePhoto{
			Photo: &client.InputFileRemote{
				Id: messagePhoto.Photo.Sizes[0].Photo.Remote.Id,
			},
			// Thumbnail: , // https://github.com/tdlib/td/issues/1505
			// A: if you use InputFileRemote, then there is no way to change the thumbnail, so there are no reasons to specify it.
			// TODO: AddedStickerFileIds: ,
			Width:   messagePhoto.Photo.Sizes[0].Width,
			Height:  messagePhoto.Photo.Sizes[0].Height,
			Caption: formattedText,
			// Ttl: ,
		}
	case client.TypeMessageVideo:
		messageVideo := messageContent.(*client.MessageVideo)
		// TODO: https://github.com/tdlib/td/issues/1504
		// var stickerSets *client.StickerSets
		// var AddedStickerFileIds []int32 // ????
		// if messageVideo.Video.HasStickers {
		// 	var err error
		// 	stickerSets, err = tdlibClient.GetAttachedStickerSets(&client.GetAttachedStickerSetsRequest{
		// 		FileId: messageVideo.Video.Video.Id,
		// 	})
		// 	if err != nil {
		// 		log.Print("GetAttachedStickerSets > ", err)
		// 	}
		// }
		return &client.InputMessageVideo{
			Video: &client.InputFileRemote{
				Id: messageVideo.Video.Video.Remote.Id,
			},
			Thumbnail: getInputThumbnail(messageVideo.Video.Thumbnail),
			// TODO: AddedStickerFileIds: ,
			Duration:          messageVideo.Video.Duration,
			Width:             messageVideo.Video.Width,
			Height:            messageVideo.Video.Height,
			SupportsStreaming: messageVideo.Video.SupportsStreaming,
			Caption:           formattedText,
			// Ttl: ,
		}
	case client.TypeMessageVoiceNote:
		return &client.InputMessageVoiceNote{
			// TODO: support ContentModeVoiceNote
			// VoiceNote: ,
			// Duration: ,
			// Waveform: ,
			Caption: formattedText,
		}
	}
	return nil
}

// TODO: ?? перенести в service/message/service.go ??
func getInputThumbnail(thumbnail *client.Thumbnail) *client.InputThumbnail {
	if thumbnail == nil || thumbnail.File == nil && thumbnail.File.Remote == nil {
		return nil
	}
	return &client.InputThumbnail{
		Thumbnail: &client.InputFileRemote{
			Id: thumbnail.File.Remote.Id,
		},
		Width:  thumbnail.Width,
		Height: thumbnail.Height,
	}
}
