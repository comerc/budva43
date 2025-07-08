package facade_grpc

import (
	"github.com/zelenin/go-tdlib/client"

	"github.com/comerc/budva43/app/dto/grpc/dto"
	"github.com/comerc/budva43/app/log"
)

//go:generate mockery --name=telegramRepo --exported
type telegramRepo interface {
	GetClientDone() <-chan any
	// tdlibClient methods
	GetMessage(*client.GetMessageRequest) (*client.Message, error)
	SendMessage(*client.SendMessageRequest) (*client.Message, error)
	SendMessageAlbum(*client.SendMessageAlbumRequest) (*client.Messages, error)
	EditMessageText(*client.EditMessageTextRequest) (*client.Message, error)
	EditMessageCaption(*client.EditMessageCaptionRequest) (*client.Message, error)
	DeleteMessages(*client.DeleteMessagesRequest) (*client.Ok, error)
	GetMessages(*client.GetMessagesRequest) (*client.Messages, error)
	GetChat(*client.GetChatRequest) (*client.Chat, error)
}

//go:generate mockery --name=messageService --exported
type messageService interface {
	GetFormattedText(*client.Message) *client.FormattedText
	GetInputMessageContent(*client.Message, *client.FormattedText) client.InputMessageContent
}

type Service struct {
	log *log.Logger
	//
	telegramRepo   telegramRepo
	messageService messageService
}

func New(telegramRepo telegramRepo, messageService messageService) *Service {
	return &Service{
		log: log.NewLogger(),
		//
		telegramRepo:   telegramRepo,
		messageService: messageService,
	}
}

func (s *Service) Start() error {
	return nil
}

func (s *Service) Close() error {
	return nil
}

func (s *Service) GetClientDone() <-chan any {
	return s.telegramRepo.GetClientDone()
}

func (s *Service) GetMessages(chatId int64, messageIds []int64) ([]*dto.Message, error) {
	var err error
	defer s.log.ErrorOrDebug(&err, "")

	var messages *client.Messages
	messages, err = s.telegramRepo.GetMessages(&client.GetMessagesRequest{
		ChatId:     chatId,
		MessageIds: messageIds,
	})
	if err != nil {
		return nil, err
	}

	var result []*dto.Message
	for _, message := range messages.Messages {
		result = append(result, &dto.Message{
			Id:   message.Id,
			Text: s.messageService.GetFormattedText(message).Text,
		})
	}

	return result, nil
}

func (s *Service) GetLastMessage(chatId int64) (*dto.Message, error) {
	var err error
	defer s.log.ErrorOrDebug(&err, "")

	var chat *client.Chat
	chat, err = s.telegramRepo.GetChat(&client.GetChatRequest{
		ChatId: chatId,
	})
	if err != nil {
		return nil, err
	}

	var result *dto.Message
	result = &dto.Message{
		Id:   chat.LastMessage.Id,
		Text: s.messageService.GetFormattedText(chat.LastMessage).Text,
	}

	return result, nil
}

func (s *Service) CreateMessage(newMessage *dto.NewMessage) (*dto.Message, error) {
	var err error
	defer s.log.ErrorOrDebug(&err, "")

	var message *client.Message
	message, err = s.telegramRepo.SendMessage(&client.SendMessageRequest{
		ChatId: newMessage.ChatId,
		InputMessageContent: &client.InputMessageText{
			Text: &client.FormattedText{
				Text: newMessage.Text,
			},
			LinkPreviewOptions: &client.LinkPreviewOptions{
				IsDisabled: true,
			},
			ClearDraft: true,
		},
		ReplyTo: &client.InputMessageReplyToMessage{
			MessageId: newMessage.ReplyToMessageId,
		},
	})
	if err != nil {
		return nil, err
	}

	var result *dto.Message
	result = &dto.Message{
		Id:   message.Id,
		Text: s.messageService.GetFormattedText(message).Text,
	}

	return result, nil
}

func (s *Service) GetMessage(chatId int64, messageId int64) (*dto.Message, error) {
	var err error
	defer s.log.ErrorOrDebug(&err, "")

	var message *client.Message
	message, err = s.telegramRepo.GetMessage(&client.GetMessageRequest{
		ChatId:    chatId,
		MessageId: messageId,
	})
	if err != nil {
		return nil, err
	}

	var result *dto.Message
	result = &dto.Message{
		Id:   message.Id,
		Text: s.messageService.GetFormattedText(message).Text,
	}

	return result, nil
}

func (s *Service) UpdateMessage(updateMessage *dto.Message) (*dto.Message, error) {
	var err error
	defer s.log.ErrorOrDebug(&err, "")

	var sourceMessage *client.Message
	sourceMessage, err = s.telegramRepo.GetMessage(&client.GetMessageRequest{
		ChatId:    updateMessage.ChatId,
		MessageId: updateMessage.Id,
	})
	if err != nil {
		return nil, err
	}

	inputMessageContent := s.messageService.GetInputMessageContent(
		sourceMessage,
		&client.FormattedText{
			Text: updateMessage.Text,
		},
	)

	var message *client.Message
	message, err = s.telegramRepo.EditMessageText(&client.EditMessageTextRequest{
		ChatId:              updateMessage.ChatId,
		MessageId:           updateMessage.Id,
		ReplyMarkup:         sourceMessage.ReplyMarkup,
		InputMessageContent: inputMessageContent,
	})
	if err != nil {
		return nil, err
	}

	var result *dto.Message
	result = &dto.Message{
		Id:   message.Id,
		Text: s.messageService.GetFormattedText(message).Text,
	}

	return result, nil
}

func (s *Service) DeleteMessages(chatId int64, messageIds []int64) (bool, error) {
	var err error
	defer s.log.ErrorOrDebug(&err, "")

	_, err = s.telegramRepo.DeleteMessages(&client.DeleteMessagesRequest{
		ChatId:     chatId,
		MessageIds: messageIds,
	})
	if err != nil {
		return false, err
	}

	return true, nil
}
