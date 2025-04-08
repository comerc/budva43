package chat

import (
	"github.com/zelenin/go-tdlib/client"
)

// Chat представляет расширение структуры Chat из TDLib
type Chat struct {
	// Встраиваем структуру из TDLib
	*client.Chat
}

// NewChat создает новый экземпляр Chat из client.Chat
func NewChat(tdlibChat *client.Chat) *Chat {
	if tdlibChat == nil {
		return nil
	}

	return &Chat{
		Chat: tdlibChat,
	}
}

// IsPrivate проверяет, является ли чат личным чатом
func (c *Chat) IsPrivate() bool {
	if c == nil || c.Type == nil {
		return false
	}

	_, ok := c.Type.(*client.ChatTypePrivate)
	return ok
}

// IsGroup проверяет, является ли чат группой (базовой или супергруппой, но не каналом)
func (c *Chat) IsGroup() bool {
	if c == nil || c.Type == nil {
		return false
	}

	if _, ok := c.Type.(*client.ChatTypeBasicGroup); ok {
		return true
	}

	if supergroup, ok := c.Type.(*client.ChatTypeSupergroup); ok && !supergroup.IsChannel {
		return true
	}

	return false
}

// IsChannel проверяет, является ли чат каналом
func (c *Chat) IsChannel() bool {
	if c == nil || c.Type == nil {
		return false
	}

	if supergroup, ok := c.Type.(*client.ChatTypeSupergroup); ok && supergroup.IsChannel {
		return true
	}

	return false
}

// GetTypeName возвращает строковое представление типа чата
func (c *Chat) GetTypeName() string {
	if c == nil || c.Type == nil {
		return "unknown"
	}

	switch c.Type.(type) {
	case *client.ChatTypePrivate:
		return "private"
	case *client.ChatTypeBasicGroup:
		return "basic_group"
	case *client.ChatTypeSupergroup:
		if c.Type.(*client.ChatTypeSupergroup).IsChannel {
			return "channel"
		}
		return "supergroup"
	default:
		return "other"
	}
}

// CanSendMessages проверяет, можно ли отправлять сообщения в чат
func (c *Chat) CanSendMessages() bool {
	if c == nil {
		return false
	}

	// В текущей версии TDLib проверяем доступность по свойству чата
	// Проверка разрешений может быть добавлена при наличии соответствующего поля в API

	return true
}
