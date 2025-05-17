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

// GetContent извлекает содержимое сообщения для поддерживаемых типов
func (s *Service) GetContent(message *client.Message) (*client.FormattedText, string) {
	if message == nil || message.Content == nil {
		return &client.FormattedText{}, ""
	}
	switch content := message.Content.(type) {
	case *client.MessageText:
		return content.Text, content.Type
	case *client.MessagePhoto:
		return content.Caption, content.Type
	case *client.MessageVideo:
		return content.Caption, content.Type
	case *client.MessageDocument:
		return content.Caption, content.Type
	case *client.MessageAudio:
		return content.Caption, content.Type
	case *client.MessageAnimation:
		return content.Caption, content.Type
	case *client.MessageVoiceNote:
		return content.Caption, content.Type
	default:
		return &client.FormattedText{}, ""
	}
}

// // IsSystemMessage проверяет, является ли сообщение системным
// func (s *Service) IsSystemMessage(message *client.Message) bool {
// 	switch message.Content.(type) {
// 	case *client.MessageChatChangeTitle,
// 		*client.MessageChatChangePhoto,
// 		*client.MessageChatDeletePhoto,
// 		*client.MessageChatAddMembers,
// 		*client.MessageChatDeleteMember,
// 		*client.MessageChatJoinByLink,
// 		*client.MessagePinMessage:
// 		return true
// 	default:
// 		return false
// 	}
// }

// // GetReplyMarkupData извлекает данные из replyMarkup
// func (s *Service) GetReplyMarkupData(message *client.Message) ([]byte, bool) {
// 	if message.ReplyMarkup != nil {
// 		if a, ok := message.ReplyMarkup.(*client.ReplyMarkupInlineKeyboard); ok {
// 			row := a.Rows[0]
// 			btn := row[0]
// 			if callback, ok := btn.Type.(*client.InlineKeyboardButtonTypeCallback); ok {
// 				return callback.Data, true
// 			}
// 		}
// 	}
// 	return nil, false
// }
