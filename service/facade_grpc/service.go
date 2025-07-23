package facade_grpc

import (
	"path/filepath"
	"slices"

	"github.com/zelenin/go-tdlib/client"

	"github.com/comerc/budva43/app/dto/grpc/dto"
	"github.com/comerc/budva43/app/log"
)

//go:generate mockery --name=telegramRepo --exported
type telegramRepo interface {
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
	GetMessageLink(*client.GetMessageLinkRequest) (*client.MessageLink, error)
	GetMessageLinkInfo(*client.GetMessageLinkInfoRequest) (*client.MessageLinkInfo, error)
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

	result := make([]*dto.Message, len(messages.Messages))
	for i, message := range messages.Messages {
		dtoMessage, err := s.mapMessage(message)
		if err != nil {
			return nil, err
		}
		result[i] = dtoMessage
	}

	return result, nil
}

func (s *Service) GetChatHistory(
	chatId int64,
	fromMessageId int64,
	offset int32,
	limit int32,
) ([]*dto.Message, error) {
	var err error

	var messages *client.Messages
	messages, err = s.telegramRepo.GetChatHistory(&client.GetChatHistoryRequest{
		ChatId:        chatId,
		FromMessageId: fromMessageId,
		Offset:        offset,
		Limit:         limit,
		OnlyLocal:     false,
	})
	if err != nil {
		return nil, err
	}

	result := make([]*dto.Message, len(messages.Messages))
	for i, message := range messages.Messages {
		dtoMessage, err := s.mapMessage(message)
		if err != nil {
			return nil, err
		}
		result[i] = dtoMessage
	}

	return result, nil
}

func (s *Service) SendMessage(newMessage *dto.NewMessage) error {
	var err error

	var formattedText *client.FormattedText
	formattedText, err = s.telegramRepo.ParseTextEntities(&client.ParseTextEntitiesRequest{
		Text: newMessage.Text,
		ParseMode: &client.TextParseModeMarkdown{
			Version: 2,
		},
	})
	if err != nil {
		return err
	}

	_, err = s.telegramRepo.SendMessage(&client.SendMessageRequest{
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
		return err
	}

	return nil
}

func (s *Service) SendMessageAlbum(newMessages []*dto.NewMessage) error {
	var err error

	var inputMessageContents []client.InputMessageContent
	for _, newMessage := range newMessages {
		var formattedText *client.FormattedText
		formattedText, err = s.telegramRepo.ParseTextEntities(&client.ParseTextEntitiesRequest{
			Text: newMessage.Text,
			ParseMode: &client.TextParseModeMarkdown{
				Version: 2,
			},
		})
		if err != nil {
			return err
		}
		fileExt := filepath.Ext(newMessage.FilePath)
		if slices.Contains([]string{".png", ".jpg", ".jpeg", ".gif", ".webp"}, fileExt) {
			inputMessageContents = append(inputMessageContents, &client.InputMessagePhoto{
				Photo: &client.InputFileLocal{
					Path: newMessage.FilePath,
				},
				Caption: formattedText,
			})
		} else if slices.Contains([]string{".mp4", ".mov", ".avi", ".mkv", ".webm"}, fileExt) {
			inputMessageContents = append(inputMessageContents, &client.InputMessageVideo{
				Video: &client.InputFileLocal{
					Path: newMessage.FilePath,
				},
				Caption: formattedText,
			})
		} else if slices.Contains([]string{".mp3", ".wav", ".ogg", ".m4a", ".aac", ".flac", ".wma", ".opus"}, fileExt) {
			inputMessageContents = append(inputMessageContents, &client.InputMessageAudio{
				Audio: &client.InputFileLocal{
					Path: newMessage.FilePath,
				},
				Caption: formattedText,
			})
		} else {
			inputMessageContents = append(inputMessageContents, &client.InputMessageDocument{
				Document: &client.InputFileLocal{
					Path: newMessage.FilePath,
				},
				Caption: formattedText,
			})
		}
	}

	var messages *client.Messages
	messages, err = s.telegramRepo.SendMessageAlbum(&client.SendMessageAlbumRequest{
		ChatId:               newMessages[0].ChatId,
		InputMessageContents: inputMessageContents,
		ReplyTo: &client.InputMessageReplyToMessage{
			MessageId: newMessages[0].ReplyToMessageId,
		},
	})
	if err != nil {
		return err
	}
	if messages.TotalCount == 0 {
		return log.NewError("no messages sent")
	}

	return nil
}

func (s *Service) ForwardMessage(chatId int64, messageId int64) error {
	var err error

	var messages *client.Messages
	messages, err = s.telegramRepo.ForwardMessages(&client.ForwardMessagesRequest{
		ChatId:     chatId,
		MessageIds: []int64{messageId},
	})
	if err != nil {
		return err
	}
	if messages.TotalCount == 0 {
		return log.NewError("no messages forwarded")
	}

	return nil
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

func (s *Service) UpdateMessage(message *dto.Message) error {
	var err error

	var sourceMessage *client.Message
	sourceMessage, err = s.telegramRepo.GetMessage(&client.GetMessageRequest{
		ChatId:    message.ChatId,
		MessageId: message.Id,
	})
	if err != nil {
		return err
	}

	var formattedText *client.FormattedText
	formattedText, err = s.telegramRepo.ParseTextEntities(&client.ParseTextEntitiesRequest{
		Text: message.Text,
		ParseMode: &client.TextParseModeMarkdown{
			Version: 2,
		},
	})
	if err != nil {
		return err
	}

	inputMessageContent := s.messageService.GetInputMessageContent(
		sourceMessage,
		formattedText,
	)

	_, err = s.telegramRepo.EditMessageText(&client.EditMessageTextRequest{
		ChatId:              message.ChatId,
		MessageId:           message.Id,
		ReplyMarkup:         sourceMessage.ReplyMarkup,
		InputMessageContent: inputMessageContent,
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) DeleteMessages(chatId int64, messageIds []int64) error {
	var err error

	_, err = s.telegramRepo.DeleteMessages(&client.DeleteMessagesRequest{
		ChatId:     chatId,
		MessageIds: messageIds,
	})
	if err != nil {
		return err
	}

	return nil
}

// GetMessageLink возвращает ссылку на сообщение
func (s *Service) GetMessageLink(chatId int64, messageId int64) (string, error) {
	var err error

	var messageLink *client.MessageLink
	messageLink, err = s.telegramRepo.GetMessageLink(&client.GetMessageLinkRequest{
		ChatId:    chatId,
		MessageId: messageId,
	})
	if err != nil {
		return "", err
	}

	return messageLink.Link, nil
}

func (s *Service) GetMessageLinkInfo(link string) (*dto.Message, error) {
	var err error

	var messageLinkInfo *client.MessageLinkInfo
	messageLinkInfo, err = s.telegramRepo.GetMessageLinkInfo(&client.GetMessageLinkInfoRequest{
		Url: link,
	})
	if err != nil {
		return nil, err
	}

	message := messageLinkInfo.Message

	var result *dto.Message
	result = &dto.Message{
		Id:      message.Id,
		Text:    "", // !! не используется
		ChatId:  message.ChatId,
		Forward: message.ForwardInfo != nil,
	}
	return result, nil
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
