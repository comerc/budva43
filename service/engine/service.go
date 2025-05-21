package engine

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"slices"
	"strings"
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
				s.handleUpdateNewMessage(updateByType)
			case *client.UpdateMessageEdited:
				s.handleUpdateMessageEdited(updateByType)
			case *client.UpdateDeleteMessages:
				s.handleUpdateDeleteMessages(updateByType)
			case *client.UpdateMessageSendSucceeded:
				s.handleUpdateMessageSendSucceeded(updateByType)
			}
		}
	}
}

// handleUpdateNewMessage обрабатывает обновление о новом сообщении
func (s *Service) handleUpdateNewMessage(update *client.UpdateNewMessage) {
	src := update.Message
	if _, ok := config.Engine.UniqueSources[src.ChatId]; !ok {
		return
	}
	if s.messageService.IsSystemMessage(src) {
		fn := func() {
			_ = s.deleteSystemMessage(src)
		}
		s.queueRepo.Add(fn)
		return
	}
	formattedText := s.messageService.GetFormattedText(src)
	if formattedText == nil {
		return
	}
	isExist := false
	forwardedTo := make(map[int64]bool)
	checkFns := make(map[int64]func())
	otherFns := make(map[int64]func())
	for _, forwardRuleId := range config.Engine.OrderedForwardRules {
		forwardRule := config.Engine.ForwardRules[forwardRuleId]
		if src.ChatId != forwardRule.From {
			continue
		}
		if !forwardRule.SendCopy && !src.CanBeSaved {
			continue
		}
		isExist = true // как минимум, собираем статистику просмотренных сообщений
		initForwardedTo(forwardedTo, forwardRule.To)
		if src.MediaAlbumId == 0 {
			fn := func() {
				_ = s.processNewMessage([]*client.Message{src}, forwardRule, forwardedTo, checkFns, otherFns)
			}
			s.queueRepo.Add(fn)
		} else {
			key := s.mediaAlbumsService.GetKey(forwardRule.Id, src.MediaAlbumId)
			isFirstMessage := s.mediaAlbumsService.AddMessage(key, src)
			if !isFirstMessage {
				continue
			}
			cb := func(messages []*client.Message) {
				_ = s.processNewMessage(messages, forwardRule, forwardedTo, checkFns, otherFns)
			}
			fn := func() {
				s.processMediaAlbum(key, cb)
			}
			s.queueRepo.Add(fn)
		}
	}
	if !isExist {
		return
	}
	fn := func() {
		s.addStatistics(forwardedTo)
		for check, fn := range checkFns {
			if fn == nil {
				s.log.Error("check is nil", "check", check)
				continue
			}
			s.log.Info("check is fn()", "check", check)
			fn()
		}
		for other, fn := range otherFns {
			if fn == nil {
				s.log.Error("other is nil", "other", other)
				continue
			}
			s.log.Info("other is fn()", "other", other)
			fn()
		}
	}
	s.queueRepo.Add(fn)
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
	message := update.Message
	tmpMessageId := update.OldMessageId
	fn := func() {
		_ = s.storageService.SetNewMessageId(message.ChatId, tmpMessageId, message.Id)
		_ = s.storageService.SetTmpMessageId(message.ChatId, message.Id, tmpMessageId)
		s.log.Info("handleUpdateMessageSendSucceeded")
	}
	s.queueRepo.Add(fn)
}

// deleteSystemMessage удаляет системное сообщение
func (s *Service) deleteSystemMessage(src *client.Message) error {
	var err error
	defer func() {
		if err != nil {
			s.log.Error("deleteSystemMessage", "err", err)
		}
	}()
	source, ok := config.Engine.Sources[src.ChatId]
	if !ok {
		return nil
	}
	if !source.DeleteSystemMessages {
		return nil
	}
	_, err = s.telegramRepo.GetClient().DeleteMessages(&client.DeleteMessagesRequest{
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

// processNewMessage обрабатывает сообщения и выполняет пересылку согласно правилам
func (s *Service) processNewMessage(messages []*client.Message,
	forwardRule *entity.ForwardRule, forwardedTo map[int64]bool,
	checkFns map[int64]func(), otherFns map[int64]func()) error {
	var (
		src         = messages[0]
		filtersMode = ""
		result      []int64
		err         error
	)
	defer func() {
		level := slog.LevelInfo
		fields := []any{
			"ChatId", src.ChatId,
			"Id", src.Id,
			"MediaAlbumId", src.MediaAlbumId,
			"filtersMode", filtersMode,
			"result", result,
		}
		if err != nil {
			level = slog.LevelError
			fields = append(fields, "err", err)
		}
		s.log.Log(context.Background(), level, "processNewMessage", fields...)
	}()

	formattedText := s.messageService.GetFormattedText(src)
	if formattedText == nil {
		err = fmt.Errorf("GetFormattedText return nil")
		return err
	}

	filtersMode = mapFiltersMode(formattedText, forwardRule)
	switch filtersMode {
	case filtersOK:
		// checkFns[rule.Check] = nil // !! не надо сбрасывать - хочу проверить сообщение, даже если где-то прошли фильтры
		otherFns[forwardRule.Other] = nil
		for _, dstChatId := range forwardRule.To {
			if isNotForwardedTo(forwardedTo, dstChatId) {
				err = s.forwardMessages(messages, src.ChatId, dstChatId, forwardRule.SendCopy, forwardRule.Id)
				result = append(result, dstChatId)
			}
		}
	case filtersCheck:
		if forwardRule.Check != 0 {
			_, ok := checkFns[forwardRule.Check]
			if !ok {
				checkFns[forwardRule.Check] = func() {
					const isSendCopy = false // обязательно надо форвардить, иначе не видно текущего сообщения
					err = s.forwardMessages(messages, src.ChatId, forwardRule.Check, isSendCopy, forwardRule.Id)
				}
			}
		}
	case filtersOther:
		if forwardRule.Other != 0 {
			_, ok := otherFns[forwardRule.Other]
			if !ok {
				otherFns[forwardRule.Other] = func() {
					const isSendCopy = true // обязательно надо копировать, иначе не видно редактирование исходного сообщения
					err = s.forwardMessages(messages, src.ChatId, forwardRule.Other, isSendCopy, forwardRule.Id)
				}
			}
		}
	}

	return err
}

// getOriginMessage получает оригинальное сообщение для пересланного сообщения
func (s *Service) getOriginMessage(message *client.Message) *client.Message {
	if message.ForwardInfo == nil {
		return nil
	}

	origin, ok := message.ForwardInfo.Origin.(*client.MessageOriginChannel)
	if !ok {
		return nil
	}

	originMessage, err := s.telegramRepo.GetClient().GetMessage(&client.GetMessageRequest{
		ChatId:    origin.ChatId,
		MessageId: origin.MessageId,
	})

	if err != nil {
		s.log.Error("getOriginMessage", "err", err)
		return nil
	}

	targetMessage := message
	targetFormattedText := s.messageService.GetFormattedText(targetMessage)
	originFormattedText := s.messageService.GetFormattedText(originMessage)
	// workaround for https://github.com/tdlib/td/issues/1572
	if targetFormattedText.Text != originFormattedText.Text {
		s.log.Debug("targetMessage != originMessage")
		return nil
	}

	return originMessage
}

func (s *Service) prepareMessageContents(messages []*client.Message, dstChatId int64) []client.InputMessageContent {
	contents := make([]client.InputMessageContent, 0)

	for i, message := range messages {
		originMessage := s.getOriginMessage(message)
		if originMessage != nil {
			messages[i] = originMessage
		}
		src := messages[i] // !! for origin message
		srcFormattedText := s.messageService.GetFormattedText(src)
		formattedText := util.Copy(srcFormattedText)

		isFirstMessageInAlbum := i == 0
		if err := s.transformService.Transform(formattedText, isFirstMessageInAlbum, src, dstChatId); err != nil {
			s.log.Error("Transform", "err", err)
		}

		content := s.messageService.GetInputMessageContent(src, formattedText)
		if content != nil {
			contents = append(contents, content)
		}
	}

	return contents
}

// getReplyToMessageId получает ID сообщения для ответа
func (s *Service) getReplyToMessageId(src *client.Message, dstChatId int64) int64 {
	var replyToMessageId int64 = 0
	var err error

	replyTo, ok := src.ReplyTo.(*client.MessageReplyToMessage)
	if !ok {
		return 0
	}

	replyToMessageId = replyTo.MessageId
	replyInChatId := replyTo.ChatId

	if replyToMessageId <= 0 || replyInChatId != src.ChatId {
		return 0
	}

	fromChatMessageId := fmt.Sprintf("%d:%d", replyInChatId, replyToMessageId)
	toChatMessageIds, err := s.storageService.GetCopiedMessageIds(fromChatMessageId)
	if err != nil {
		s.log.Error("GetCopiedMessageIds", "err", err)
		return 0
	}

	if len(toChatMessageIds) == 0 {
		return 0
	}

	var tmpMessageId int64 = 0
	for _, toChatMessageId := range toChatMessageIds {
		a := strings.Split(toChatMessageId, ":")
		if util.ConvertToInt[int64](a[1]) == dstChatId {
			tmpMessageId = util.ConvertToInt[int64](a[2])
			break
		}
	}

	if tmpMessageId == 0 {
		return 0
	}

	replyToMessageId, err = s.storageService.GetNewMessageId(dstChatId, tmpMessageId)
	if err != nil {
		s.log.Error("GetNewMessageId", "err", err)
		return 0
	}

	return replyToMessageId
}

// sendMessages отправляет сообщения в чат
func (s *Service) sendMessages(dstChatId int64, contents []client.InputMessageContent, replyToMessageId int64) (*client.Messages, error) {
	if len(contents) == 1 {
		var message *client.Message
		message, err := s.telegramRepo.GetClient().SendMessage(&client.SendMessageRequest{
			ChatId:              dstChatId,
			InputMessageContent: contents[0],
			ReplyTo: &client.InputMessageReplyToMessage{
				MessageId: replyToMessageId,
			},
		})
		if err != nil {
			return nil, err
		}
		return &client.Messages{
			TotalCount: 1,
			Messages:   []*client.Message{message},
		}, nil
	}
	return s.telegramRepo.GetClient().SendMessageAlbum(&client.SendMessageAlbumRequest{
		ChatId:               dstChatId,
		InputMessageContents: contents,
		ReplyTo: &client.InputMessageReplyToMessage{
			MessageId: replyToMessageId,
		},
	})
}

// forwardMessages пересылает сообщения в целевой чат
func (s *Service) forwardMessages(messages []*client.Message, srcChatId, dstChatId int64, isSendCopy bool, forwardRuleId string) error {
	// TODO: не возвращается ошибка - это нормально?
	s.log.Debug("forwardMessages",
		"srcChatId", srcChatId,
		"dstChatId", dstChatId,
		"sendCopy", isSendCopy,
		"forwardRuleId", forwardRuleId,
		"messageCount", len(messages))

	s.rateLimiterService.WaitForForward(s.ctx, dstChatId)

	var (
		result *client.Messages
		err    error
	)
	defer func() {
		if err != nil {
			s.log.Error("forwardMessages", "err", err)
		}
	}()

	if isSendCopy {
		contents := s.prepareMessageContents(messages, dstChatId)
		replyToMessageId := s.getReplyToMessageId(messages[0], dstChatId)
		result, err = s.sendMessages(dstChatId, contents, replyToMessageId)
	} else {
		result, err = s.telegramRepo.GetClient().ForwardMessages(&client.ForwardMessagesRequest{
			ChatId:     dstChatId,
			FromChatId: srcChatId,
			MessageIds: func() []int64 {
				var messageIds []int64
				for _, message := range messages {
					messageIds = append(messageIds, message.Id)
				}
				return messageIds
			}(),
			Options: &client.MessageSendOptions{
				DisableNotification: false,
				FromBackground:      false,
				SchedulingState: &client.MessageSchedulingStateSendAtDate{
					SendDate: int32(time.Now().Unix()),
				},
			},
			SendCopy:      false,
			RemoveCaption: false,
		})
	}

	if err != nil {
		return err
	}

	if len(result.Messages) != int(result.TotalCount) || result.TotalCount == 0 {
		return fmt.Errorf("invalid TotalCount")
	}

	if len(result.Messages) != len(messages) {
		return fmt.Errorf("invalid len(messages)")
	}

	if isSendCopy {
		for i, dst := range result.Messages {
			if dst == nil {
				s.log.Error("forwardMessages - dst == nil !!", "result", result, "messages", messages)
				continue
			}
			tmpMessageId := dst.Id
			src := messages[i] // !! for origin message (in prepareMessageContents)
			toChatMessageId := fmt.Sprintf("%s:%d:%d", forwardRuleId, dstChatId, tmpMessageId)
			fromChatMessageId := fmt.Sprintf("%d:%d", src.ChatId, src.Id)
			s.storageService.SetCopiedMessageId(fromChatMessageId, toChatMessageId)
			// TODO: isAnswer
			if _, ok := s.messageService.GetReplyMarkupData(src); ok {
				s.storageService.SetAnswerMessageId(dstChatId, tmpMessageId, fromChatMessageId)
			}
		}
	}

	return nil
}

const waitForMediaAlbum = 3 * time.Second

// processMediaAlbum обрабатывает медиа-альбом
func (s *Service) processMediaAlbum(key entity.MediaAlbumKey, cb func([]*client.Message)) {
	// TODO: не возвращается error ?
	diff := s.mediaAlbumsService.GetLastReceivedDiff(key)
	if diff < waitForMediaAlbum {
		time.Sleep(waitForMediaAlbum - diff)
		s.processMediaAlbum(key, cb)
		return
	}
	messages := s.mediaAlbumsService.PopMessages(key)
	cb(messages)
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

// addStatistics добавляет статистику пересылаемых и просмотренных сообщений
func (s *Service) addStatistics(forwardedTo map[int64]bool) {
	date := util.GetCurrentDate()
	for dstChatId, isForwarded := range forwardedTo {
		if isForwarded {
			s.storageService.IncrementForwardedMessages(dstChatId, date)
		}
		s.storageService.IncrementViewedMessages(dstChatId, date)
	}
}

var forwardedToMu sync.Mutex

// initForwardedTo инициализирует forwardedTo для новых чатов
func initForwardedTo(forwardedTo map[int64]bool, dstChatIds []int64) {
	forwardedToMu.Lock()
	defer forwardedToMu.Unlock()
	for _, dstChatId := range dstChatIds {
		_, ok := forwardedTo[dstChatId]
		if !ok {
			forwardedTo[dstChatId] = false
		}
	}
}

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
func mapFiltersMode(formattedText *client.FormattedText, rule *entity.ForwardRule) filtersMode {
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
