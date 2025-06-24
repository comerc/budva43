package forwarder

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/zelenin/go-tdlib/client"

	"github.com/comerc/budva43/app/entity"
	"github.com/comerc/budva43/app/log"
	"github.com/comerc/budva43/app/util"
)

//go:generate mockery --name=telegramRepo --exported
type telegramRepo interface {
	// tdlibClient methods
	ForwardMessages(*client.ForwardMessagesRequest) (*client.Messages, error)
	GetMessage(*client.GetMessageRequest) (*client.Message, error)
	SendMessage(*client.SendMessageRequest) (*client.Message, error)
	SendMessageAlbum(*client.SendMessageAlbumRequest) (*client.Messages, error)
}

//go:generate mockery --name=storageService --exported
type storageService interface {
	SetCopiedMessageId(chatId, messageId int64, toChatMessageId string)
	GetCopiedMessageIds(chatId, messageId int64) []string
	GetNewMessageId(chatId, tmpMessageId int64) int64
	SetAnswerMessageId(dstChatId, tmpMessageId, chatId, messageId int64)
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

//go:generate mockery --name=rateLimiterService --exported
type rateLimiterService interface {
	WaitForForward(ctx context.Context, dstChatId int64)
}

type Service struct {
	log *log.Logger
	ctx context.Context
	//
	telegramRepo       telegramRepo
	storageService     storageService
	messageService     messageService
	transformService   transformService
	rateLimiterService rateLimiterService
}

func New(
	telegramRepo telegramRepo,
	storageService storageService,
	messageService messageService,
	transformService transformService,
	rateLimiterService rateLimiterService,
) *Service {
	return &Service{
		log: log.NewLogger("service.forwarder"),
		//
		telegramRepo:       telegramRepo,
		storageService:     storageService,
		messageService:     messageService,
		transformService:   transformService,
		rateLimiterService: rateLimiterService,
	}
}

// ForwardMessages пересылает сообщения в целевой чат
func (s *Service) ForwardMessages(
	messages []*client.Message, filtersMode entity.FiltersMode,
	srcChatId, dstChatId int64, isSendCopy bool, forwardRuleId string,
	engineConfig *entity.EngineConfig,
) {
	var err error
	defer s.log.ErrorOrDebug(&err, "ForwardMessages",
		"filtersMode", filtersMode,
		"srcChatId", srcChatId,
		"dstChatId", dstChatId,
		"isSendCopy", isSendCopy,
		"forwardRuleId", forwardRuleId,
		"len(messages)", len(messages),
	)

	s.rateLimiterService.WaitForForward(s.ctx, dstChatId)

	var result *client.Messages

	if isSendCopy {
		contents := s.prepareMessageContents(messages, dstChatId, engineConfig)
		replyToMessageId := s.getReplyToMessageId(messages[0], dstChatId)
		result, err = s.sendMessages(dstChatId, contents, replyToMessageId)
		err = log.WrapError(err)
	} else {
		result, err = s.telegramRepo.ForwardMessages(&client.ForwardMessagesRequest{
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
					SendDate: int32(time.Now().Unix()), // nolint:gosec
				},
			},
			SendCopy:      false,
			RemoveCaption: false,
		})
		err = log.WrapError(err)
	}

	if err != nil {
		return
	}

	if len(result.Messages) != int(result.TotalCount) || result.TotalCount == 0 {
		err = log.NewError("invalid value", "result.TotalCount", result.TotalCount)
		return
	}

	if len(result.Messages) != len(messages) {
		err = log.NewError("invalid value", "len(result.Messages)", len(result.Messages))
		return
	}

	if isSendCopy {
		for i, dst := range result.Messages {
			tmpMessageId := dst.Id
			src := messages[i] // !! for origin message (in prepareMessageContents)
			toChatMessageId := fmt.Sprintf("%s:%d:%d", forwardRuleId, dstChatId, tmpMessageId)
			s.storageService.SetCopiedMessageId(src.ChatId, src.Id, toChatMessageId)
			// TODO: isAnswer
			if replyMarkupData := s.messageService.GetReplyMarkupData(src); len(replyMarkupData) > 0 {
				s.storageService.SetAnswerMessageId(dstChatId, tmpMessageId, src.ChatId, src.Id)
			}
		}
	}
}

// getOriginMessage получает оригинальное сообщение для пересланного сообщения
func (s *Service) getOriginMessage(message *client.Message) *client.Message {
	var err error
	defer s.log.ErrorOrDebug(&err, "getOriginMessage")

	if message.ForwardInfo == nil {
		s.log.Warn("message.ForwardInfo is nil")
		return nil
	}

	origin, ok := message.ForwardInfo.Origin.(*client.MessageOriginChannel)
	if !ok {
		err = log.NewError("invalid message.ForwardInfo.Origin")
		return nil
	}

	var originMessage *client.Message
	originMessage, err = s.telegramRepo.GetMessage(&client.GetMessageRequest{
		ChatId:    origin.ChatId,
		MessageId: origin.MessageId,
	})
	if err != nil {
		err = log.WrapError(err)
		return nil
	}

	targetMessage := message
	targetFormattedText := s.messageService.GetFormattedText(targetMessage)
	originFormattedText := s.messageService.GetFormattedText(originMessage)
	// workaround for https://github.com/tdlib/td/issues/1572
	if targetFormattedText.Text != originFormattedText.Text {
		err = log.NewError("targetMessage != originMessage")
		return nil
	}

	return originMessage
}

// prepareMessageContents подготавливает сообщения для отправки
func (s *Service) prepareMessageContents(messages []*client.Message, dstChatId int64, engineConfig *entity.EngineConfig) []client.InputMessageContent {
	contents := make([]client.InputMessageContent, 0)

	for i, message := range messages {
		func() {
			var err error
			defer s.log.ErrorOrDebug(&err, "prepareMessageContents",
				"i", i,
				"chatId", message.ChatId,
				"messageId", message.Id,
			)

			originMessage := s.getOriginMessage(message)
			if originMessage != nil {
				messages[i] = originMessage
			}
			src := messages[i] // !! for origin message

			srcFormattedText := s.messageService.GetFormattedText(src)
			var formattedText *client.FormattedText
			formattedText, err = util.DeepCopy(srcFormattedText)
			if err != nil {
				err = log.WrapError(err)
				return
			}

			withSources := i == 0
			s.transformService.Transform(formattedText, withSources, src, dstChatId, engineConfig)

			content := s.messageService.GetInputMessageContent(src, formattedText)
			if content != nil {
				contents = append(contents, content)
			}
		}()
	}

	return contents
}

// getReplyToMessageId получает ID сообщения для ответа
func (s *Service) getReplyToMessageId(src *client.Message, dstChatId int64) int64 {
	var err error
	defer s.log.ErrorOrDebug(&err, "getReplyToMessageId")

	var replyToMessageId int64
	replyTo, ok := src.ReplyTo.(*client.MessageReplyToMessage)
	if !ok {
		err = log.NewError("invalid src.ReplyTo")
		return 0
	}

	replyToMessageId = replyTo.MessageId
	if replyToMessageId == 0 {
		err = log.NewError("invalid replyToMessageId")
		return 0
	}

	replyInChatId := replyTo.ChatId
	if replyInChatId != src.ChatId {
		err = log.NewError("replyInChatId != src.ChatId")
		return 0
	}

	toChatMessageIds := s.storageService.GetCopiedMessageIds(replyInChatId, replyToMessageId)

	if len(toChatMessageIds) == 0 {
		err = log.NewError("toChatMessageIds is empty")
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
		err = log.NewError("tmpMessageId is 0")
		return 0
	}

	replyToMessageId = s.storageService.GetNewMessageId(dstChatId, tmpMessageId)

	return replyToMessageId
}

// sendMessages отправляет сообщения в чат
func (s *Service) sendMessages(dstChatId int64, contents []client.InputMessageContent, replyToMessageId int64) (*client.Messages, error) {
	var err error

	if len(contents) == 1 {
		var message *client.Message
		message, err = s.telegramRepo.SendMessage(&client.SendMessageRequest{
			ChatId:              dstChatId,
			InputMessageContent: contents[0],
			ReplyTo: &client.InputMessageReplyToMessage{
				MessageId: replyToMessageId,
			},
		})
		if err != nil {
			return nil, log.WrapError(err)
		}
		return &client.Messages{
			TotalCount: 1,
			Messages:   []*client.Message{message},
		}, nil
	}
	var messages *client.Messages
	messages, err = s.telegramRepo.SendMessageAlbum(&client.SendMessageAlbumRequest{
		ChatId:               dstChatId,
		InputMessageContents: contents,
		ReplyTo: &client.InputMessageReplyToMessage{
			MessageId: replyToMessageId,
		},
	})
	if err != nil {
		return nil, log.WrapError(err)
	}
	return messages, nil
}
