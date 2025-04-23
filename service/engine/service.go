package engine

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/zelenin/go-tdlib/client"

	"github.com/comerc/budva43/config"
	"github.com/comerc/budva43/entity"
)

// messageService определяет интерфейс сервиса сообщений, необходимый для сервиса engine
type messageService interface {
	GetText(message *client.Message) string
	GetCaption(message *client.Message) string
	// SendMessage(chatID int64, text string) (*client.Message, error)
	// ForwardMessage(fromChatID, messageID, toChatID int64) (*client.Message, error)
	// SendMessageAlbum(chatID int64, contents []client.InputMessageContent) (*client.Messages, error)
	// ForwardMessages(fromChatID int64, messageIDs []int64, toChatID int64) (*client.Messages, error)
	// EditMessageText(chatID, messageID int64, text *client.FormattedText) (*client.Message, error)
	// EditMessageMedia(chatID, messageID int64, content client.InputMessageContent) (*client.Message, error)
	// EditMessageCaption(chatID, messageID int64, caption *client.FormattedText) (*client.Message, error)
	DeleteMessages(chatID int64, messageIDs []int64) error
	GetMessage(chatID, messageID int64) (*client.Message, error)
}

// filterService определяет интерфейс сервиса фильтрации, необходимый для сервиса engine
type filterService interface {
	ShouldForward(message *client.Message, rule *entity.ForwardRule) (bool, error)
}

// transformService определяет интерфейс сервиса трансформации, необходимый для сервиса engine
type transformService interface {
	// ReplaceMyselfLinks(text *client.FormattedText, srcChatID, dstChatID int64) error
	// ReplaceFragments(text *client.FormattedText, dstChatID int64) error
	// AddSourceSign(text *client.FormattedText, title string) error
	// AddSourceLink(message *client.Message, text *client.FormattedText, title string) error
}

// storageService определяет интерфейс сервиса хранилища, необходимый для сервиса engine
type storageService interface {
	// SetCopiedMessageID(fromChatMessageID string, toChatMessageID string) error
	GetCopiedMessageIDs(fromChatMessageID string) ([]string, error)
	// DeleteCopiedMessageIDs(fromChatMessageID string) error
	SetNewMessageID(chatID, tmpMessageID, newMessageID int64) error
	// GetNewMessageID(chatID, tmpMessageID int64) (int64, error)
	// DeleteNewMessageID(chatID, tmpMessageID int64) error
	SetTmpMessageID(chatID, newMessageID, tmpMessageID int64) error
	// GetTmpMessageID(chatID, newMessageID int64) (int64, error)
	// DeleteTmpMessageID(chatID, newMessageID int64) error
	// IncrementViewedMessages(toChatID int64) error
	// IncrementForwardedMessages(toChatID int64) error
	GetRuleByID(ruleID string) (entity.ForwardRule, bool)
}

// mediaAlbumService определяет интерфейс сервиса медиа-альбомов, необходимый для сервиса engine
type mediaAlbumService interface {
	AddMessage(forwardKey string, message *client.Message) bool
	GetLastReceivedDiff(forwardKey string, albumID client.JsonInt64) time.Duration
	GetMessages(forwardKey string, albumID client.JsonInt64) []*client.Message
}

type telegramRepo interface {
	GetClient() *client.Client
	AuthClientDone() chan any
}

// Service предоставляет функциональность движка пересылки сообщений
type Service struct {
	log *slog.Logger
	//
	message      messageService
	filter       filterService
	transform    transformService
	storage      storageService
	mediaAlbums  mediaAlbumService
	telegramRepo telegramRepo
	queue        chan func()
}

// New создает новый экземпляр сервиса engine
func New(
	message messageService,
	filter filterService,
	transform transformService,
	storage storageService,
	mediaAlbums mediaAlbumService,
	telegramRepo telegramRepo,

	// tdlibClient *client.Client,
) *Service {
	return nil

	return &Service{
		log: slog.With("module", "service.engine"),
		//
		message:      message,
		filter:       filter,
		transform:    transform,
		storage:      storage,
		mediaAlbums:  mediaAlbums,
		telegramRepo: telegramRepo,
		queue:        make(chan func(), 100),
	}
}

// Start запускает обработчик обновлений от Telegram
func (s *Service) Start(ctx context.Context) error {
	return nil

	s.log.Info("Запуск сервиса engine")

	// Проверяем конфигурацию
	if err := s.validateConfig(); err != nil {
		return fmt.Errorf("ошибка валидации конфигурации: %w", err)
	}

	go s.processQueue(ctx)

	go func() {
		// Ждем авторизации клиента
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

// Stop останавливает сервис
func (s *Service) Stop() error {
	return nil

	s.log.Info("Остановка сервиса engine")
	close(s.queue)
	return nil
}

// Проверяет корректность конфигурации
func (s *Service) validateConfig() error {
	// Проверяем заменяемые фрагменты текста
	for chatID, settings := range config.Engine.ReplaceFragments {
		for from, to := range settings.Replacements {
			if len([]rune(from)) != len([]rune(to)) {
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

// Обрабатывает очередь отложенных задач
func (s *Service) processQueue(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case fn, ok := <-s.queue:
			if !ok {
				return
			}
			fn()
		}
	}
}

// Обрабатывает обновления от Telegram
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

// Обрабатывает обновление о новом сообщении
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
		if rule.From == chatID {
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
	if delete := config.Engine.DeleteSystemMessages[chatID]; delete {
		if s.isSystemMessage(message) {
			go func() {
				if err := s.message.DeleteMessages(chatID, []int64{message.Id}); err != nil {
					s.log.Error("Ошибка удаления системного сообщения", "err", err)
				}
			}()
		}
	}

	// Проверяем тип сообщения, пропускаем необрабатываемые типы
	text := s.message.GetText(message)
	if text == "" {
		text = s.message.GetCaption(message)
		if text == "" && !s.isMediaContent(message) {
			return
		}
	}

	// Обрабатываем медиа-альбом
	if message.MediaAlbumId != 0 {
		for _, ruleInfo := range forwardRules {
			isFirstMessage := s.mediaAlbums.AddMessage(ruleInfo.ruleID, message)
			if isFirstMessage {
				s.queue <- func() {
					s.handleMediaAlbum(ruleInfo.ruleID, message.MediaAlbumId)
				}
			}
		}
		return
	}

	// Обрабатываем обычное сообщение
	for _, ruleInfo := range forwardRules {
		rule := ruleInfo.rule
		if rule.Status != entity.RuleStatusActive {
			continue
		}

		// Копируем сообщение для каждого правила в очередь, чтобы не блокировать обработку
		messageCopy := message
		ruleIDCopy := ruleInfo.ruleID
		ruleCopy := rule

		s.queue <- func() {
			s.processMessage([]*client.Message{messageCopy}, ruleIDCopy, ruleCopy)
		}
	}
}

// Обрабатывает обновление о редактировании сообщения
func (s *Service) handleUpdateMessageEdited(update *client.UpdateMessageEdited) {
	chatID := update.ChatId
	messageID := update.MessageId

	// Проверяем, является ли чат источником для какого-либо правила
	isSourceChat := false
	for _, rule := range config.Engine.Forwards {
		if rule.From == chatID {
			isSourceChat = true
			break
		}
	}

	if !isSourceChat {
		return
	}

	s.queue <- func() {
		// Получаем информацию о скопированных сообщениях
		fromChatMessageID := fmt.Sprintf("%d:%d", chatID, messageID)
		toChatMessageIDs, err := s.storage.GetCopiedMessageIDs(fromChatMessageID)
		if err != nil {
			s.log.Error("Ошибка получения копий сообщения", "err", err)
			return
		}

		if len(toChatMessageIDs) == 0 {
			return
		}

		// Получаем обновленное сообщение
		message, err := s.message.GetMessage(chatID, messageID)
		if err != nil {
			s.log.Error("Ошибка получения сообщения", "err", err)
			return
		}

		// Обрабатываем редактирование для каждой копии
		for _, toChatMessageID := range toChatMessageIDs {
			s.processSingleEdited(message, toChatMessageID)
		}
	}
}

// Обрабатывает обновление об удалении сообщений
func (s *Service) handleUpdateDeleteMessages(update *client.UpdateDeleteMessages) {
	if !update.IsPermanent {
		return
	}

	chatID := update.ChatId
	messageIDs := update.MessageIds

	// Проверяем, является ли чат источником для какого-либо правила
	isSourceChat := false
	for _, rule := range config.Engine.Forwards {
		if rule.From == chatID {
			isSourceChat = true
			break
		}
	}

	if !isSourceChat {
		return
	}

	s.queue <- func() {
		for _, messageID := range messageIDs {
			fromChatMessageID := fmt.Sprintf("%d:%d", chatID, messageID)
			toChatMessageIDs, err := s.storage.GetCopiedMessageIDs(fromChatMessageID)
			if err != nil {
				s.log.Error("Ошибка получения копий сообщения", "err", err)
				continue
			}

			if len(toChatMessageIDs) == 0 {
				continue
			}

			// Удаляем копии, если не установлен флаг Indelible
			for _, toChatMessageID := range toChatMessageIDs {
				s.processSingleDeleted(fromChatMessageID, toChatMessageID)
			}
		}
	}
}

// Обрабатывает обновление об успешной отправке сообщения
func (s *Service) handleUpdateMessageSendSucceeded(update *client.UpdateMessageSendSucceeded) {
	message := update.Message
	tmpMessageID := update.OldMessageId

	s.queue <- func() {
		if err := s.storage.SetNewMessageID(message.ChatId, tmpMessageID, message.Id); err != nil {
			s.log.Error("Ошибка сохранения нового ID сообщения", "err", err)
		}

		if err := s.storage.SetTmpMessageID(message.ChatId, message.Id, tmpMessageID); err != nil {
			s.log.Error("Ошибка сохранения временного ID сообщения", "err", err)
		}
	}
}

// Проверяет, является ли сообщение системным
func (s *Service) isSystemMessage(message *client.Message) bool {
	switch message.Content.(type) {
	case *client.MessageChatChangeTitle:
		return true
	case *client.MessageChatChangePhoto:
		return true
	case *client.MessageChatDeletePhoto:
		return true
	case *client.MessageChatAddMembers:
		return true
	case *client.MessageChatDeleteMember:
		return true
	case *client.MessageChatJoinByLink:
		return true
	case *client.MessagePinMessage:
		return true
	default:
		return false
	}
}

// Проверяет, содержит ли сообщение медиа-контент
func (s *Service) isMediaContent(message *client.Message) bool {
	switch message.Content.(type) {
	case *client.MessagePhoto:
		return true
	case *client.MessageVideo:
		return true
	case *client.MessageAnimation:
		return true
	case *client.MessageAudio:
		return true
	case *client.MessageDocument:
		return true
	case *client.MessageVoiceNote:
		return true
	case *client.MessageVideoNote:
		return true
	default:
		return false
	}
}

// Обрабатывает медиа-альбом
func (s *Service) handleMediaAlbum(forwardKey string, albumID client.JsonInt64) {
	// Ждем, пока соберутся все сообщения альбома
	const waitForMediaAlbum = 1 * time.Second
	diff := s.mediaAlbums.GetLastReceivedDiff(forwardKey, albumID)
	if diff < waitForMediaAlbum {
		time.Sleep(waitForMediaAlbum - diff)
		// Повторная проверка после ожидания, т.к. могли поступить новые сообщения
		s.handleMediaAlbum(forwardKey, albumID)
		return
	}

	// Получаем все сообщения альбома
	messages := s.mediaAlbums.GetMessages(forwardKey, albumID)
	if len(messages) == 0 {
		return
	}

	// Находим правило пересылки
	rule, ok := s.storage.GetRuleByID(forwardKey)
	if !ok {
		s.log.Error("Правило пересылки не найдено", "forwardKey", forwardKey)
		return
	}

	if rule.Status != entity.RuleStatusActive {
		return
	}

	// Обрабатываем альбом
	s.processMessage(messages, forwardKey, rule)
}

// Обрабатывает сообщение или группу сообщений
func (s *Service) processMessage(messages []*client.Message, forwardKey string, rule entity.ForwardRule) {
	// Реализация будет добавлена позже
}

// Обрабатывает редактирование одиночного сообщения
func (s *Service) processSingleEdited(message *client.Message, toChatMessageID string) {
	// Реализация будет добавлена позже
}

// Обрабатывает удаление одиночного сообщения
func (s *Service) processSingleDeleted(fromChatMessageID, toChatMessageID string) {
	// Реализация будет добавлена позже
}
