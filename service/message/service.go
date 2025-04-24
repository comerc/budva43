package message

import (
	"log/slog"

	"github.com/zelenin/go-tdlib/client"
)

// Service предоставляет методы для обработки и преобразования сообщений
type Service struct {
	log *slog.Logger
	//
}

// New создает новый экземпляр сервиса для работы с сообщениями
func New() *Service {
	return &Service{
		log: slog.With("module", "service.message"),
		//
	}
}

// GetText возвращает текст сообщения, если это текстовое сообщение
func (s *Service) GetText(message *client.Message) string {
	if content, ok := message.Content.(*client.MessageText); ok {
		return content.Text.Text
	}

	return ""
}

// GetCaption возвращает подпись медиа сообщения
func (s *Service) GetCaption(message *client.Message) string {
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
	_, ok := message.Content.(*client.MessageText)
	return ok
}

// IsSystemMessage проверяет, является ли сообщение системным
func (s *Service) IsSystemMessage(message *client.Message) bool {
	switch message.Content.(type) {
	case *client.MessageChatChangeTitle,
		*client.MessageChatChangePhoto,
		*client.MessageChatDeletePhoto,
		*client.MessageChatAddMembers,
		*client.MessageChatDeleteMember,
		*client.MessageChatJoinByLink,
		*client.MessagePinMessage:
		return true
	default:
		return false
	}
}

// IsMediaMessage проверяет, содержит ли сообщение медиа-контент
func (s *Service) IsMediaMessage(message *client.Message) bool {
	switch message.Content.(type) {
	case *client.MessagePhoto,
		*client.MessageVideo,
		*client.MessageDocument,
		*client.MessageAudio,
		*client.MessageAnimation,
		*client.MessageVoiceNote:
		return true
	default:
		return false
	}
}

// // GetContentType возвращает тип содержимого сообщения
// func (s *Service) GetContentType(message *client.Message) ContentType {
// 	switch message.Content.(type) {
// 	case *client.MessageText:
// 		return ContentTypeText
// 	case *client.MessagePhoto:
// 		return ContentTypePhoto
// 	case *client.MessageVideo:
// 		return ContentTypeVideo
// 	case *client.MessageDocument:
// 		return ContentTypeDocument
// 	case *client.MessageAudio:
// 		return ContentTypeAudio
// 	case *client.MessageAnimation:
// 		return ContentTypeAnimation
// 	case *client.MessageVoiceNote:
// 		return ContentTypeVoice
// 	default:
// 		return ""
// 	}
// }

// // FormatMessageContent преобразует формат сообщения (например, из markdown в HTML)
// func (s *Service) FormatMessageContent(text string, fromFormat, toFormat string) (string, error) {
// 	// Здесь будет реализация преобразования между форматами
// 	// Например, конвертация Markdown в HTML или обратно

// 	// Пока просто заглушка
// 	return text, nil
// }

// // ExtractMessageMetadata извлекает метаданные из сообщения
// func (s *Service) ExtractMessageMetadata(message *client.Message) map[string]any {
// 	metadata := make(map[string]any)

// 	metadata["messageId"] = message.Id
// 	metadata["chatId"] = message.ChatId
// 	metadata["date"] = message.Date

// 	// Извлечение специфичных метаданных в зависимости от типа сообщения
// 	switch content := message.Content.(type) {
// 	case *client.MessageText:
// 		metadata["contentType"] = "text"
// 		metadata["textLength"] = len(content.Text.Text)
// 	case *client.MessagePhoto:
// 		metadata["contentType"] = "photo"
// 		metadata["captionLength"] = len(content.Caption.Text)
// 		// TODO: реализовать другие типы сообщений...
// 	}

// 	return metadata
// }
