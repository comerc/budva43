package message

import (
	"time"

	"github.com/zelenin/go-tdlib/client"
)

// Message представляет расширение структуры Message из TDLib с дополнительными полями
type Message struct {
	// Встраиваем структуру из TDLib
	*client.Message

	// Дополнительные поля, которых нет в TDLib или которые удобно иметь в преобразованном виде
	ParsedDate time.Time
}

// NewMessage создает новый экземпляр Message из client.Message
func NewMessage(tdlibMessage *client.Message) *Message {
	if tdlibMessage == nil {
		return nil
	}

	return &Message{
		Message:    tdlibMessage,
		ParsedDate: time.Unix(int64(tdlibMessage.Date), 0),
	}
}

// GetText возвращает текст сообщения, если это текстовое сообщение
func (m *Message) GetText() string {
	if m == nil || m.Content == nil {
		return ""
	}

	if content, ok := m.Content.(*client.MessageText); ok {
		return content.Text.Text
	}

	return ""
}

// GetCaption возвращает подпись медиа сообщения
func (m *Message) GetCaption() string {
	if m == nil || m.Content == nil {
		return ""
	}

	switch content := m.Content.(type) {
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
func (m *Message) IsTextMessage() bool {
	if m == nil || m.Content == nil {
		return false
	}

	_, ok := m.Content.(*client.MessageText)
	return ok
}

// IsMediaMessage проверяет, содержит ли сообщение медиа-контент
func (m *Message) IsMediaMessage() bool {
	if m == nil || m.Content == nil {
		return false
	}

	switch m.Content.(type) {
	case *client.MessagePhoto, *client.MessageVideo, *client.MessageDocument,
		*client.MessageAudio, *client.MessageAnimation, *client.MessageVoiceNote:
		return true
	}

	return false
}

// GetContentType возвращает строковое представление типа содержимого сообщения
func (m *Message) GetContentType() string {
	if m == nil || m.Content == nil {
		return "unknown"
	}

	switch m.Content.(type) {
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
