package update_new_message

import (
	"context"
	"time"

	"github.com/zelenin/go-tdlib/client"

	"github.com/comerc/budva43/app/config"
	"github.com/comerc/budva43/app/entity"
	"github.com/comerc/budva43/app/log"
	"github.com/comerc/budva43/app/util"
)

//go:generate mockery --name=telegramRepo --exported
type telegramRepo interface {
	// tdlibClient methods
	DeleteMessages(*client.DeleteMessagesRequest) (*client.Ok, error)
}

//go:generate mockery --name=queueRepo --exported
type queueRepo interface {
	Add(task func())
}

//go:generate mockery --name=storageService --exported
type storageService interface {
	IncrementViewedMessages(toChatId int64, date string)
	IncrementForwardedMessages(toChatId int64, date string)
}

//go:generate mockery --name=messageService --exported
type messageService interface {
	GetFormattedText(message *client.Message) *client.FormattedText
	IsSystemMessage(message *client.Message) bool
}

//go:generate mockery --name=mediaAlbumService --exported
type mediaAlbumService interface {
	AddMessage(key entity.MediaAlbumKey, message *client.Message) bool
	GetLastReceivedDiff(key entity.MediaAlbumKey) time.Duration
	PopMessages(key entity.MediaAlbumKey) []*client.Message
	GetKey(forwardRuleId entity.ForwardRuleId, MediaAlbumId client.JsonInt64) entity.MediaAlbumKey
}

//go:generate mockery --name=filtersModeService --exported
type filtersModeService interface {
	Map(formattedText *client.FormattedText, rule *entity.ForwardRule) entity.FiltersMode
}

//go:generate mockery --name=forwardedToService --exported
type forwardedToService interface {
	Init(forwardedTo map[int64]bool, dstChatIds []int64)
	Add(forwardedTo map[int64]bool, dstChatId int64) bool
}

//go:generate mockery --name=forwarderService --exported
type forwarderService interface {
	ForwardMessages(messages []*client.Message, srcChatId, dstChatId int64, isSendCopy bool, forwardRuleId string)
}

type Handler struct {
	log *log.Logger
	ctx context.Context
	//
	telegramRepo       telegramRepo
	queueRepo          queueRepo
	storageService     storageService
	messageService     messageService
	mediaAlbumsService mediaAlbumService
	filtersModeService filtersModeService
	forwardedToService forwardedToService
	forwarderService   forwarderService
}

func New(
	telegramRepo telegramRepo,
	queueRepo queueRepo,
	storageService storageService,
	messageService messageService,
	mediaAlbumsService mediaAlbumService,
	filtersModeService filtersModeService,
	forwardedToService forwardedToService,
	forwarderService forwarderService,
) *Handler {
	return &Handler{
		log: log.NewLogger("handler.update_new_message"),
		//
		telegramRepo:       telegramRepo,
		queueRepo:          queueRepo,
		storageService:     storageService,
		messageService:     messageService,
		mediaAlbumsService: mediaAlbumsService,
		filtersModeService: filtersModeService,
		forwardedToService: forwardedToService,
		forwarderService:   forwarderService,
	}
}

// Run выполняет обрабатку обновления о новом сообщении
func (h *Handler) Run(ctx context.Context, update *client.UpdateNewMessage) {
	h.ctx = ctx
	src := update.Message
	defer h.log.Debug("Run",
		"chatId", src.ChatId,
		"messageId", src.Id,
	)

	if _, ok := config.Engine.UniqueSources[src.ChatId]; !ok {
		return
	}
	if h.messageService.IsSystemMessage(src) {
		fn := func() {
			h.deleteSystemMessage(src)
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
		h.forwardedToService.Init(forwardedTo, forwardRule.To)
		if src.MediaAlbumId == 0 {
			fn := func() {
				h.processMessage([]*client.Message{src}, forwardRule, forwardedTo, checkFns, otherFns)
			}
			h.queueRepo.Add(fn)
		} else {
			key := h.mediaAlbumsService.GetKey(forwardRule.Id, src.MediaAlbumId)
			isFirstMessage := h.mediaAlbumsService.AddMessage(key, src)
			if !isFirstMessage {
				continue
			}
			cb := func(messages []*client.Message) {
				h.processMessage(messages, forwardRule, forwardedTo, checkFns, otherFns)
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
			func() {
				var err error
				defer h.log.ErrorOrDebug(&err, "has fn",
					"check", check,
				)
				if fn == nil {
					err = log.NewError("fn is nil")
					return
				}
				fn()
			}()
		}
		for other, fn := range otherFns {
			func() {
				var err error
				defer h.log.ErrorOrDebug(&err, "has fn",
					"other", other,
				)
				if fn == nil {
					err = log.NewError("fn is nil")
					return
				}
				fn()
			}()
		}
	}
	h.queueRepo.Add(fn)
}

// deleteSystemMessage удаляет системное сообщение
func (h *Handler) deleteSystemMessage(src *client.Message) {
	var err error
	defer h.log.ErrorOrDebug(&err, "deleteSystemMessage",
		"chatId", src.ChatId,
		"messageId", src.Id,
	)

	source, ok := config.Engine.Sources[src.ChatId]
	if !ok {
		return
	}
	if !source.DeleteSystemMessages {
		return
	}
	_, err = h.telegramRepo.DeleteMessages(&client.DeleteMessagesRequest{
		ChatId:     src.ChatId,
		MessageIds: []int64{src.Id},
		Revoke:     true,
	})
}

// processMessage обрабатывает сообщения и выполняет пересылку согласно правилам
func (h *Handler) processMessage(messages []*client.Message,
	forwardRule *entity.ForwardRule, forwardedTo map[int64]bool,
	checkFns map[int64]func(), otherFns map[int64]func()) {
	var err error
	src := messages[0]
	filtersMode := ""
	result := []int64{}
	defer func() {
		h.log.ErrorOrDebug(&err, "processMessage",
			"chatId", src.ChatId,
			"messageId", src.Id,
			"mediaAlbumId", src.MediaAlbumId,
			"filtersMode", filtersMode,
			"result", result,
		)
	}()

	formattedText := h.messageService.GetFormattedText(src)
	if formattedText == nil {
		err = log.NewError("GetFormattedText return nil")
		return
	}

	filtersMode = h.filtersModeService.Map(formattedText, forwardRule)
	switch filtersMode {
	case entity.FiltersOK:
		// checkFns[rule.Check] = nil // !! не надо сбрасывать - хочу проверить сообщение, даже если где-то прошли фильтры
		otherFns[forwardRule.Other] = nil
		for _, dstChatId := range forwardRule.To {
			if h.forwardedToService.Add(forwardedTo, dstChatId) {
				h.forwarderService.ForwardMessages(messages, src.ChatId, dstChatId, forwardRule.SendCopy, forwardRule.Id)
				result = append(result, dstChatId)
			}
		}
	case entity.FiltersCheck:
		if forwardRule.Check != 0 {
			_, ok := checkFns[forwardRule.Check]
			if !ok {
				checkFns[forwardRule.Check] = func() {
					const isSendCopy = false // обязательно надо форвардить, иначе не видно текущего сообщения
					h.forwarderService.ForwardMessages(messages, src.ChatId, forwardRule.Check, isSendCopy, forwardRule.Id)
				}
			}
		}
	case entity.FiltersOther:
		if forwardRule.Other != 0 {
			_, ok := otherFns[forwardRule.Other]
			if !ok {
				otherFns[forwardRule.Other] = func() {
					const isSendCopy = true // обязательно надо копировать, иначе не видно редактирование исходного сообщения
					h.forwarderService.ForwardMessages(messages, src.ChatId, forwardRule.Other, isSendCopy, forwardRule.Id)
				}
			}
		}
	}
}

const waitForMediaAlbum = 3 * time.Second

// processMediaAlbum обрабатывает медиа-альбом
func (h *Handler) processMediaAlbum(key entity.MediaAlbumKey, cb func([]*client.Message)) {
	diff := h.mediaAlbumsService.GetLastReceivedDiff(key)
	if diff < waitForMediaAlbum {
		timer := time.NewTimer(waitForMediaAlbum - diff)
		defer timer.Stop()

		select {
		case <-h.ctx.Done():
			return
		case <-timer.C:
			h.processMediaAlbum(key, cb)
		}
		return
	}
	messages := h.mediaAlbumsService.PopMessages(key)
	cb(messages)
}

// addStatistics добавляет статистику пересылаемых и просмотренных сообщений
func (h *Handler) addStatistics(forwardedTo map[int64]bool) {
	date := util.GetCurrentDate()
	for dstChatId, ok := range forwardedTo {
		if ok {
			h.storageService.IncrementForwardedMessages(dstChatId, date)
		}
		h.storageService.IncrementViewedMessages(dstChatId, date)
	}
}
