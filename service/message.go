package service

import (
	"github.com/comerc/budva43/entity"
	"github.com/zelenin/go-tdlib/client"
)

// MessageService предоставляет методы для работы с сообщениями
type MessageService struct {
	// Здесь могут быть зависимости, например, репозитории
}

// NewMessageService создает новый экземпляр сервиса для работы с сообщениями
func NewMessageService() *MessageService {
	return &MessageService{}
}

// GetText возвращает текст сообщения, если это текстовое сообщение
func (s *MessageService) GetText(message *entity.Message) string {
	if message == nil || message.Content == nil {
		return ""
	}

	if content, ok := message.Content.(*client.MessageText); ok {
		return content.Text.Text
	}

	return ""
}

// GetCaption возвращает подпись медиа сообщения
func (s *MessageService) GetCaption(message *entity.Message) string {
	if message == nil || message.Content == nil {
		return ""
	}

	switch content := message.Content.(type) {
	case *client.MessagePhoto:
		return content.Caption.Text
	case *client.MessageVideo:
		return content.Caption.Text
	case *client.MessageDocument:
		return content.Caption.Text
	case *client.MessageAudio:
		return content.Caption.Text
	case *client.MessageAnimation:
		return content.Caption.Text
	case *client.MessageVoiceNote:
		return content.Caption.Text
	}

	return ""
}

// IsTextMessage проверяет, является ли сообщение текстовым
func (s *MessageService) IsTextMessage(message *entity.Message) bool {
	if message == nil || message.Content == nil {
		return false
	}

	_, ok := message.Content.(*client.MessageText)
	return ok
}

// IsMediaMessage проверяет, содержит ли сообщение медиа-контент
func (s *MessageService) IsMediaMessage(message *entity.Message) bool {
	if message == nil || message.Content == nil {
		return false
	}

	switch message.Content.(type) {
	case *client.MessagePhoto, *client.MessageVideo, *client.MessageDocument,
		*client.MessageAudio, *client.MessageAnimation, *client.MessageVoiceNote:
		return true
	}

	return false
}

// GetContentType возвращает строковое представление типа содержимого сообщения
func (s *MessageService) GetContentType(message *entity.Message) string {
	if message == nil || message.Content == nil {
		return "unknown"
	}

	switch message.Content.(type) {
	case *client.MessageText:
		return "text"
	case *client.MessagePhoto:
		return "photo"
	case *client.MessageVideo:
		return "video"
	case *client.MessageDocument:
		return "document"
	case *client.MessageAudio:
		return "audio"
	case *client.MessageAnimation:
		return "animation"
	case *client.MessageVoiceNote:
		return "voice_note"
	case *client.MessageVideoNote:
		return "video_note"
	case *client.MessageSticker:
		return "sticker"
	case *client.MessageLocation:
		return "location"
	case *client.MessageContact:
		return "contact"
	case *client.MessagePoll:
		return "poll"
	default:
		return "other"
	}
}
