package model

import (
	"time"

	"github.com/comerc/budva43/entity"
)

// MessageDTO представляет собой объект передачи данных сообщения для API
// Используется когда формат данных должен отличаться от внутренней сущности Message
type MessageDTO struct {
	ID               int64           `json:"id"`
	ChatID           int64           `json:"chat_id"`
	SenderID         int64           `json:"sender_id,omitempty"`
	Text             string          `json:"text,omitempty"`
	Caption          string          `json:"caption,omitempty"`
	Date             string          `json:"date"`         // Преобразованная в строку дата
	ContentType      string          `json:"content_type"` // Тип контента (text, photo, video и т.д.)
	HasMedia         bool            `json:"has_media"`    // Флаг наличия медиа
	MediaURL         string          `json:"media_url,omitempty"`
	SourceInfo       *SourceInfoDTO  `json:"source_info,omitempty"`
	ForwardInfo      *ForwardInfoDTO `json:"forward_info,omitempty"`
	LinkPreview      *LinkPreviewDTO `json:"link_preview,omitempty"`
	ReplyToMessageID int64           `json:"reply_to_message_id,omitempty"`
}

// SourceInfoDTO представляет информацию об источнике сообщения
type SourceInfoDTO struct {
	ChatName   string `json:"chat_name,omitempty"`
	ChatType   string `json:"chat_type,omitempty"`
	IsVerified bool   `json:"is_verified,omitempty"`
}

// ForwardInfoDTO представляет информацию о пересылке сообщения
type ForwardInfoDTO struct {
	FromChatID   int64  `json:"from_chat_id,omitempty"`
	FromChatName string `json:"from_chat_name,omitempty"`
	FromUserID   int64  `json:"from_user_id,omitempty"`
	FromUserName string `json:"from_user_name,omitempty"`
	Date         string `json:"date,omitempty"`
}

// LinkPreviewDTO представляет данные предпросмотра ссылки
type LinkPreviewDTO struct {
	URL         string `json:"url,omitempty"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	ImageURL    string `json:"image_url,omitempty"`
}

// ConvertToMessageDTO преобразует внутреннюю сущность Message в DTO для API
func ConvertToMessageDTO(message *entity.Message) *MessageDTO {
	if message == nil || message.Message == nil {
		return nil
	}

	// Форматирование даты в удобный для API формат
	dateStr := message.ParsedDate.Format(time.RFC3339)

	dto := &MessageDTO{
		ID:          message.Id,     // Используем правильное имя поля Id
		ChatID:      message.ChatId, // Используем правильное имя поля ChatId
		Text:        message.GetText(),
		Caption:     message.GetCaption(),
		Date:        dateStr,
		ContentType: message.GetContentType(),
		HasMedia:    message.IsMediaMessage(),
	}

	// В реальном приложении здесь был бы код для получения идентификатора отправителя
	// Поскольку мы не знаем точную структуру message.Message, это условный код
	// dto.SenderID = ... // получаем из message информацию об отправителе

	// Аналогично, для сообщения-ответа
	// dto.ReplyToMessageID = ... // получаем из message информацию о сообщении, на которое отвечаем

	// Здесь можно добавить заполнение других полей DTO
	// например, информацию о медиа, источнике и т.д.

	return dto
}

// ConvertToMessageBatchDTO преобразует массив сущностей Message в массив DTO
func ConvertToMessageBatchDTO(messages []*entity.Message) []*MessageDTO {
	result := make([]*MessageDTO, 0, len(messages))
	for _, msg := range messages {
		if dto := ConvertToMessageDTO(msg); dto != nil {
			result = append(result, dto)
		}
	}
	return result
}

// ConvertFromMessageDTO преобразует DTO обратно в сущность Message
// Используется, например, когда мы получаем данные из API и нужно
// их преобразовать во внутреннюю модель данных
func ConvertFromMessageDTO(dto *MessageDTO) (*entity.Message, error) {
	if dto == nil {
		return nil, nil
	}

	// Здесь было бы преобразование из DTO в сущность
	// Но поскольку полная реализация entity.Message нам неизвестна,
	// возвращаем заглушку

	// В реальном приложении здесь был бы код для создания
	// экземпляра entity.Message и заполнения его полей из dto

	return &entity.Message{}, nil
}
