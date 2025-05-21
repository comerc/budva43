package update_new_message

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

type telegramRepo interface {
	GetClient() *client.Client
}

type queueRepo interface {
	Add(task func())
}

type storageService interface {
	SetCopiedMessageId(fromChatMessageId string, toChatMessageId string) error
	GetCopiedMessageIds(fromChatMessageId string) ([]string, error)
	GetNewMessageId(chatId, tmpMessageId int64) (int64, error)
	IncrementViewedMessages(toChatId int64, date string) error
	IncrementForwardedMessages(toChatId int64, date string) error
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

type Handler struct {
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

func New(
	telegramRepo telegramRepo,
	queueRepo queueRepo,
	storageService storageService,
	messageService messageService,
	mediaAlbumsService mediaAlbumService,
	transformService transformService,
	rateLimiterService rateLimiterService,
) *Handler {
	return &Handler{
		log: slog.With("module", "handler.update_new_message"),
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

// Run выполняет обрабатку обновления о новом сообщении
func (h *Handler) Run(ctx context.Context, update *client.UpdateNewMessage) {
	h.ctx = ctx
	src := update.Message
	if _, ok := config.Engine.UniqueSources[src.ChatId]; !ok {
		return
	}
	if h.messageService.IsSystemMessage(src) {
		fn := func() {
			_ = h.deleteSystemMessage(src)
		}
		h.queueRepo.Add(fn)
		return
	}
	formattedText := h.messageService.GetFormattedText(src)
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
				_ = h.processNewMessage([]*client.Message{src}, forwardRule, forwardedTo, checkFns, otherFns)
			}
			h.queueRepo.Add(fn)
		} else {
			key := h.mediaAlbumsService.GetKey(forwardRule.Id, src.MediaAlbumId)
			isFirstMessage := h.mediaAlbumsService.AddMessage(key, src)
			if !isFirstMessage {
				continue
			}
			cb := func(messages []*client.Message) {
				_ = h.processNewMessage(messages, forwardRule, forwardedTo, checkFns, otherFns)
			}
			fn := func() {
				h.processMediaAlbum(key, cb)
			}
			h.queueRepo.Add(fn)
		}
	}
	if !isExist {
		return
	}
	fn := func() {
		h.addStatistics(forwardedTo)
		for check, fn := range checkFns {
			if fn == nil {
				h.log.Error("check is nil", "check", check)
				continue
			}
			h.log.Info("check is fn()", "check", check)
			fn()
		}
		for other, fn := range otherFns {
			if fn == nil {
				h.log.Error("other is nil", "other", other)
				continue
			}
			h.log.Info("other is fn()", "other", other)
			fn()
		}
	}
	h.queueRepo.Add(fn)
}

// deleteSystemMessage удаляет системное сообщение
func (h *Handler) deleteSystemMessage(src *client.Message) error {
	var err error
	defer func() {
		if err != nil {
			h.log.Error("deleteSystemMessage", "err", err)
		}
	}()
	source, ok := config.Engine.Sources[src.ChatId]
	if !ok {
		return nil
	}
	if !source.DeleteSystemMessages {
		return nil
	}
	_, err = h.telegramRepo.GetClient().DeleteMessages(&client.DeleteMessagesRequest{
		ChatId:     src.ChatId,
		MessageIds: []int64{src.Id},
		Revoke:     true,
	})
	return err
}

// processNewMessage обрабатывает сообщения и выполняет пересылку согласно правилам
func (h *Handler) processNewMessage(messages []*client.Message,
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
		h.log.Log(context.Background(), level, "processNewMessage", fields...)
	}()

	formattedText := h.messageService.GetFormattedText(src)
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
				err = h.forwardMessages(messages, src.ChatId, dstChatId, forwardRule.SendCopy, forwardRule.Id)
				result = append(result, dstChatId)
			}
		}
	case filtersCheck:
		if forwardRule.Check != 0 {
			_, ok := checkFns[forwardRule.Check]
			if !ok {
				checkFns[forwardRule.Check] = func() {
					const isSendCopy = false // обязательно надо форвардить, иначе не видно текущего сообщения
					err = h.forwardMessages(messages, src.ChatId, forwardRule.Check, isSendCopy, forwardRule.Id)
				}
			}
		}
	case filtersOther:
		if forwardRule.Other != 0 {
			_, ok := otherFns[forwardRule.Other]
			if !ok {
				otherFns[forwardRule.Other] = func() {
					const isSendCopy = true // обязательно надо копировать, иначе не видно редактирование исходного сообщения
					err = h.forwardMessages(messages, src.ChatId, forwardRule.Other, isSendCopy, forwardRule.Id)
				}
			}
		}
	}

	return err
}

// getOriginMessage получает оригинальное сообщение для пересланного сообщения
func (h *Handler) getOriginMessage(message *client.Message) *client.Message {
	if message.ForwardInfo == nil {
		return nil
	}

	origin, ok := message.ForwardInfo.Origin.(*client.MessageOriginChannel)
	if !ok {
		return nil
	}

	originMessage, err := h.telegramRepo.GetClient().GetMessage(&client.GetMessageRequest{
		ChatId:    origin.ChatId,
		MessageId: origin.MessageId,
	})

	if err != nil {
		h.log.Error("getOriginMessage", "err", err)
		return nil
	}

	targetMessage := message
	targetFormattedText := h.messageService.GetFormattedText(targetMessage)
	originFormattedText := h.messageService.GetFormattedText(originMessage)
	// workaround for https://github.com/tdlib/td/issues/1572
	if targetFormattedText.Text != originFormattedText.Text {
		h.log.Debug("targetMessage != originMessage")
		return nil
	}

	return originMessage
}

// prepareMessageContents подготавливает сообщения для отправки
func (h *Handler) prepareMessageContents(messages []*client.Message, dstChatId int64) []client.InputMessageContent {
	contents := make([]client.InputMessageContent, 0)

	for i, message := range messages {
		originMessage := h.getOriginMessage(message)
		if originMessage != nil {
			messages[i] = originMessage
		}
		src := messages[i] // !! for origin message
		srcFormattedText := h.messageService.GetFormattedText(src)
		formattedText := util.Copy(srcFormattedText)

		isFirstMessageInAlbum := i == 0
		if err := h.transformService.Transform(formattedText, isFirstMessageInAlbum, src, dstChatId); err != nil {
			h.log.Error("Transform", "err", err)
		}

		content := h.messageService.GetInputMessageContent(src, formattedText)
		if content != nil {
			contents = append(contents, content)
		}
	}

	return contents
}

// getReplyToMessageId получает ID сообщения для ответа
func (h *Handler) getReplyToMessageId(src *client.Message, dstChatId int64) int64 {
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
	toChatMessageIds, err := h.storageService.GetCopiedMessageIds(fromChatMessageId)
	if err != nil {
		h.log.Error("GetCopiedMessageIds", "err", err)
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

	replyToMessageId, err = h.storageService.GetNewMessageId(dstChatId, tmpMessageId)
	if err != nil {
		h.log.Error("GetNewMessageId", "err", err)
		return 0
	}

	return replyToMessageId
}

// sendMessages отправляет сообщения в чат
func (h *Handler) sendMessages(dstChatId int64, contents []client.InputMessageContent, replyToMessageId int64) (*client.Messages, error) {
	if len(contents) == 1 {
		var message *client.Message
		message, err := h.telegramRepo.GetClient().SendMessage(&client.SendMessageRequest{
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
	return h.telegramRepo.GetClient().SendMessageAlbum(&client.SendMessageAlbumRequest{
		ChatId:               dstChatId,
		InputMessageContents: contents,
		ReplyTo: &client.InputMessageReplyToMessage{
			MessageId: replyToMessageId,
		},
	})
}

// forwardMessages пересылает сообщения в целевой чат
func (h *Handler) forwardMessages(messages []*client.Message, srcChatId, dstChatId int64, isSendCopy bool, forwardRuleId string) error {
	// TODO: не возвращается ошибка - это нормально?
	h.log.Debug("forwardMessages",
		"srcChatId", srcChatId,
		"dstChatId", dstChatId,
		"sendCopy", isSendCopy,
		"forwardRuleId", forwardRuleId,
		"messageCount", len(messages))

	h.rateLimiterService.WaitForForward(h.ctx, dstChatId)

	var (
		result *client.Messages
		err    error
	)
	defer func() {
		if err != nil {
			h.log.Error("forwardMessages", "err", err)
		}
	}()

	if isSendCopy {
		contents := h.prepareMessageContents(messages, dstChatId)
		replyToMessageId := h.getReplyToMessageId(messages[0], dstChatId)
		result, err = h.sendMessages(dstChatId, contents, replyToMessageId)
	} else {
		result, err = h.telegramRepo.GetClient().ForwardMessages(&client.ForwardMessagesRequest{
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
				h.log.Error("forwardMessages - dst == nil !!", "result", result, "messages", messages)
				continue
			}
			tmpMessageId := dst.Id
			src := messages[i] // !! for origin message (in prepareMessageContents)
			toChatMessageId := fmt.Sprintf("%s:%d:%d", forwardRuleId, dstChatId, tmpMessageId)
			fromChatMessageId := fmt.Sprintf("%d:%d", src.ChatId, src.Id)
			h.storageService.SetCopiedMessageId(fromChatMessageId, toChatMessageId)
			// TODO: isAnswer
			if _, ok := h.messageService.GetReplyMarkupData(src); ok {
				h.storageService.SetAnswerMessageId(dstChatId, tmpMessageId, fromChatMessageId)
			}
		}
	}

	return nil
}

const waitForMediaAlbum = 3 * time.Second

// processMediaAlbum обрабатывает медиа-альбом
func (h *Handler) processMediaAlbum(key entity.MediaAlbumKey, cb func([]*client.Message)) {
	// TODO: не возвращается error ?
	diff := h.mediaAlbumsService.GetLastReceivedDiff(key)
	if diff < waitForMediaAlbum {
		time.Sleep(waitForMediaAlbum - diff)
		h.processMediaAlbum(key, cb)
		return
	}
	messages := h.mediaAlbumsService.PopMessages(key)
	cb(messages)
}

// addStatistics добавляет статистику пересылаемых и просмотренных сообщений
func (h *Handler) addStatistics(forwardedTo map[int64]bool) {
	date := util.GetCurrentDate()
	for dstChatId, isForwarded := range forwardedTo {
		if isForwarded {
			h.storageService.IncrementForwardedMessages(dstChatId, date)
		}
		h.storageService.IncrementViewedMessages(dstChatId, date)
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
