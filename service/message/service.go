package message_processing

import (
	"github.com/comerc/budva43/entity"
	"github.com/zelenin/go-tdlib/client"
)

// MessageProcessingService предоставляет методы для обработки и преобразования сообщений
type MessageProcessingService struct {
	// Здесь могут быть зависимости, например, репозитории
}

// NewMessageProcessingService создает новый экземпляр сервиса для работы с сообщениями
func NewMessageProcessingService() *MessageProcessingService {
	return &MessageProcessingService{}
}

// GetText возвращает текст сообщения, если это текстовое сообщение
func (s *MessageProcessingService) GetText(message *entity.Message) string {
	if message == nil || message.Content == nil {
		return ""
	}

	if content, ok := message.Content.(*client.MessageText); ok {
		return content.Text.Text
	}

	return ""
}

// GetCaption возвращает подпись медиа сообщения
func (s *MessageProcessingService) GetCaption(message *entity.Message) string {
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
func (s *MessageProcessingService) IsTextMessage(message *entity.Message) bool {
	if message == nil || message.Content == nil {
		return false
	}

	_, ok := message.Content.(*client.MessageText)
	return ok
}

// IsMediaMessage проверяет, содержит ли сообщение медиа-контент
func (s *MessageProcessingService) IsMediaMessage(message *entity.Message) bool {
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

// GetContentType возвращает тип содержимого сообщения
func (s *MessageProcessingService) GetContentType(message *entity.Message) string {
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
		return "voice"
	default:
		return "other"
	}
}

// FormatMessageContent преобразует формат сообщения (например, из markdown в HTML)
func (s *MessageProcessingService) FormatMessageContent(text string, fromFormat, toFormat string) (string, error) {
	// Здесь будет реализация преобразования между форматами
	// Например, конвертация Markdown в HTML или обратно

	// Пока просто заглушка
	return text, nil
}

// ExtractMessageMetadata извлекает метаданные из сообщения
func (s *MessageProcessingService) ExtractMessageMetadata(message *entity.Message) map[string]interface{} {
	metadata := make(map[string]interface{})

	if message == nil {
		return metadata
	}

	metadata["messageId"] = message.Id
	metadata["chatId"] = message.ChatId
	metadata["date"] = message.Date

	// Извлечение специфичных метаданных в зависимости от типа сообщения
	switch content := message.Content.(type) {
	case *client.MessageText:
		metadata["contentType"] = "text"
		metadata["textLength"] = len(content.Text.Text)
	case *client.MessagePhoto:
		metadata["contentType"] = "photo"
		metadata["captionLength"] = len(content.Caption.Text)
		// Другие типы сообщений...
	}

	return metadata
}
