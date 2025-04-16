package message

import (
	"github.com/zelenin/go-tdlib/client"
)

// Service предоставляет методы для обработки и преобразования сообщений
type Service struct {
	// Здесь могут быть зависимости, например, репозитории
}

// New создает новый экземпляр сервиса для работы с сообщениями
func New() *Service {
	return &Service{}
}

// GetText возвращает текст сообщения, если это текстовое сообщение
func (s *Service) GetText(message *client.Message) string {
	if message == nil || message.Content == nil {
		return ""
	}

	if content, ok := message.Content.(*client.MessageText); ok {
		return content.Text.Text
	}

	return ""
}

// GetCaption возвращает подпись медиа сообщения
func (s *Service) GetCaption(message *client.Message) string {
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
func (s *Service) IsTextMessage(message *client.Message) bool {
	if message == nil || message.Content == nil {
		return false
	}

	_, ok := message.Content.(*client.MessageText)
	return ok
}

// IsMediaMessage проверяет, содержит ли сообщение медиа-контент
func (s *Service) IsMediaMessage(message *client.Message) bool {
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
func (s *Service) GetContentType(message *client.Message) string {
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
func (s *Service) FormatMessageContent(text string, fromFormat, toFormat string) (string, error) {
	// Здесь будет реализация преобразования между форматами
	// Например, конвертация Markdown в HTML или обратно

	// Пока просто заглушка
	return text, nil
}

// ExtractMessageMetadata извлекает метаданные из сообщения
func (s *Service) ExtractMessageMetadata(message *client.Message) map[string]any {
	metadata := make(map[string]any)

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
