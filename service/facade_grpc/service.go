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
	ForwardMessages(*client.ForwardMessagesRequest) (*client.Messages, error)
	EditMessageText(*client.EditMessageTextRequest) (*client.Message, error)
	EditMessageCaption(*client.EditMessageCaptionRequest) (*client.Message, error)
	DeleteMessages(*client.DeleteMessagesRequest) (*client.Ok, error)
	GetMessages(*client.GetMessagesRequest) (*client.Messages, error)
	GetChatHistory(*client.GetChatHistoryRequest) (*client.Messages, error)
	GetMarkdownText(*client.GetMarkdownTextRequest) (*client.FormattedText, error)
	ParseTextEntities(*client.ParseTextEntitiesRequest) (*client.FormattedText, error)
}

//go:generate mockery --name=messageService --exported
type messageService interface {
	GetFormattedText(*client.Message) *client.FormattedText
	GetInputMessageContent(*client.Message, *client.FormattedText) client.InputMessageContent
}

//go:generate mockery --name=mediaAlbumService --exported
type mediaAlbumService interface {
	// TODO: пригодится для реализации API?
}

type Service struct {
	log *log.Logger
	//
	telegramRepo      telegramRepo
	messageService    messageService
	mediaAlbumService mediaAlbumService
}

func New(
	telegramRepo telegramRepo,
	messageService messageService,
	mediaAlbumService mediaAlbumService,
) *Service {
	return &Service{
		log: log.NewLogger(),
		//
		telegramRepo:      telegramRepo,
		messageService:    messageService,
		mediaAlbumService: mediaAlbumService,
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
		dtoMessage, err := s.mapMessage(message)
		if err != nil {
			return nil, err
		}
		result = append(result, dtoMessage)
	}

	return result, nil
}

// GetLastMessage возвращает последнее сообщение в чате (только локальные)
func (s *Service) GetLastMessage(chatId int64) (*dto.Message, error) {
	var err error

	var messages *client.Messages
	messages, err = s.telegramRepo.GetChatHistory(&client.GetChatHistoryRequest{
		ChatId:    chatId,
		Limit:     1,
		OnlyLocal: true,
	})
	if err != nil {
		return nil, err
	}
	if messages.TotalCount == 0 {
		return nil, nil
	}

	return s.mapMessage(messages.Messages[0])
}

func (s *Service) SendMessage(newMessage *dto.NewMessage) (*dto.Message, error) {
	var err error

	var formattedText *client.FormattedText
	formattedText, err = s.telegramRepo.ParseTextEntities(&client.ParseTextEntitiesRequest{
		Text: newMessage.Text,
		ParseMode: &client.TextParseModeMarkdown{
			Version: 2,
		},
	})
	if err != nil {
		return nil, err
	}

	var message *client.Message
	message, err = s.telegramRepo.SendMessage(&client.SendMessageRequest{
		ChatId: newMessage.ChatId,
		InputMessageContent: &client.InputMessageText{
			Text: formattedText,
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

	// TODO: дожидаться client.UpdateMessageSendSucceeded, подставлять реальный message.Id

	return s.mapMessage(message)
}

func (s *Service) ForwardMessage(chatId int64, messageId int64) (*dto.Message, error) {
	var err error

	var messages *client.Messages
	messages, err = s.telegramRepo.ForwardMessages(&client.ForwardMessagesRequest{
		ChatId:     chatId,
		MessageIds: []int64{messageId},
	})
	if err != nil {
		return nil, err
	}
	if messages.TotalCount == 0 {
		return nil, nil
	}

	return s.mapMessage(messages.Messages[0])
}

func (s *Service) GetMessage(chatId int64, messageId int64) (*dto.Message, error) {
	var err error

	var message *client.Message
	message, err = s.telegramRepo.GetMessage(&client.GetMessageRequest{
		ChatId:    chatId,
		MessageId: messageId,
	})
	if err != nil {
		return nil, err
	}

	return s.mapMessage(message)
}

func (s *Service) UpdateMessage(updateMessage *dto.Message) (*dto.Message, error) {
	var err error

	var sourceMessage *client.Message
	sourceMessage, err = s.telegramRepo.GetMessage(&client.GetMessageRequest{
		ChatId:    updateMessage.ChatId,
		MessageId: updateMessage.Id,
	})
	if err != nil {
		return nil, err
	}

	var formattedText *client.FormattedText
	formattedText, err = s.telegramRepo.ParseTextEntities(&client.ParseTextEntitiesRequest{
		Text: updateMessage.Text,
		ParseMode: &client.TextParseModeMarkdown{
			Version: 2,
		},
	})
	if err != nil {
		return nil, err
	}

	inputMessageContent := s.messageService.GetInputMessageContent(
		sourceMessage,
		formattedText,
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

	return s.mapMessage(message)
}

func (s *Service) DeleteMessages(chatId int64, messageIds []int64) (bool, error) {
	var err error

	_, err = s.telegramRepo.DeleteMessages(&client.DeleteMessagesRequest{
		ChatId:     chatId,
		MessageIds: messageIds,
	})
	if err != nil {
		return false, err
	}

	return true, nil
}

// mapMessage преобразует сообщение из tdlib в dto.Message
func (s *Service) mapMessage(message *client.Message) (*dto.Message, error) {
	var err error

	var formattedText *client.FormattedText
	formattedText, err = s.telegramRepo.GetMarkdownText(&client.GetMarkdownTextRequest{
		Text: s.messageService.GetFormattedText(message),
	})
	if err != nil {
		return nil, err
	}

	var result *dto.Message
	result = &dto.Message{
		Id:      message.Id,
		Text:    formattedText.Text,
		ChatId:  message.ChatId,
		Forward: message.ForwardInfo != nil,
	}

	return result, nil
}
