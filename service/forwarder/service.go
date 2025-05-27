package forwarder

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/zelenin/go-tdlib/client"

	"github.com/comerc/budva43/util"
)

type telegramRepo interface {
	GetClient() *client.Client
}

type storageService interface {
	SetCopiedMessageId(fromChatMessageId string, toChatMessageId string)
	GetCopiedMessageIds(fromChatMessageId string) []string
	GetNewMessageId(chatId, tmpMessageId int64) int64
	SetAnswerMessageId(dstChatId, tmpMessageId int64, fromChatMessageId string)
}

type messageService interface {
	GetFormattedText(message *client.Message) *client.FormattedText
	GetInputMessageContent(message *client.Message, formattedText *client.FormattedText) client.InputMessageContent
	GetReplyMarkupData(message *client.Message) []byte
}

type transformService interface {
	Transform(formattedText *client.FormattedText, withSources bool, src *client.Message, dstChatId int64)
}

type rateLimiterService interface {
	WaitForForward(ctx context.Context, dstChatId int64)
}

type Service struct {
	log *slog.Logger
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
		log: slog.With("module", "service.forwarder"),
		//
		telegramRepo:       telegramRepo,
		storageService:     storageService,
		messageService:     messageService,
		transformService:   transformService,
		rateLimiterService: rateLimiterService,
	}
}

// ForwardMessages пересылает сообщения в целевой чат
func (s *Service) ForwardMessages(messages []*client.Message, srcChatId, dstChatId int64, isSendCopy bool, forwardRuleId string) error {
	// s.log.Debug("ForwardMessages",
	// 	"srcChatId", srcChatId,
	// 	"dstChatId", dstChatId,
	// 	"sendCopy", isSendCopy,
	// 	"forwardRuleId", forwardRuleId,
	// 	"messageCount", len(messages))

	s.rateLimiterService.WaitForForward(s.ctx, dstChatId)

	var (
		result *client.Messages
		err    error
	)
	defer func() {
		if err != nil {
			// s.log.Error("ForwardMessages", "err", err)
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
					SendDate: int32(time.Now().Unix()), // nolint:gosec
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
				// s.log.Error("ForwardMessages - dst == nil !!", "result", result, "messages", messages)
				continue
			}
			tmpMessageId := dst.Id
			src := messages[i] // !! for origin message (in prepareMessageContents)
			toChatMessageId := fmt.Sprintf("%s:%d:%d", forwardRuleId, dstChatId, tmpMessageId)
			fromChatMessageId := fmt.Sprintf("%d:%d", src.ChatId, src.Id)
			s.storageService.SetCopiedMessageId(fromChatMessageId, toChatMessageId)
			// TODO: isAnswer
			if replyMarkupData := s.messageService.GetReplyMarkupData(src); len(replyMarkupData) > 0 {
				s.storageService.SetAnswerMessageId(dstChatId, tmpMessageId, fromChatMessageId)
			}
		}
	}

	return nil
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
		// s.log.Error("getOriginMessage", "err", err)
		return nil
	}

	targetMessage := message
	targetFormattedText := s.messageService.GetFormattedText(targetMessage)
	originFormattedText := s.messageService.GetFormattedText(originMessage)
	// workaround for https://github.com/tdlib/td/issues/1572
	if targetFormattedText.Text != originFormattedText.Text {
		// s.log.Debug("targetMessage != originMessage")
		return nil
	}

	return originMessage
}

// prepareMessageContents подготавливает сообщения для отправки
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

		withSources := i == 0
		s.transformService.Transform(formattedText, withSources, src, dstChatId)

		content := s.messageService.GetInputMessageContent(src, formattedText)
		if content != nil {
			contents = append(contents, content)
		}
	}

	return contents
}

// getReplyToMessageId получает ID сообщения для ответа
func (s *Service) getReplyToMessageId(src *client.Message, dstChatId int64) int64 {
	var replyToMessageId int64

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
	toChatMessageIds := s.storageService.GetCopiedMessageIds(fromChatMessageId)

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

	replyToMessageId = s.storageService.GetNewMessageId(dstChatId, tmpMessageId)

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
