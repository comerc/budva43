package chat

// Service предоставляет методы для работы с чатами
type Service struct {
	// Здесь могут быть зависимости, например, репозитории
}

// New создает новый экземпляр сервиса для работы с чатами
func New() *Service {
	return &Service{}
}

// // IsPrivate проверяет, является ли чат личным чатом
// func (s *ChatService) IsPrivate(chat *entity.Chat) bool {
// 	if chat == nil || chat.Type == nil {
// 		return false
// 	}

// 	_, ok := chat.Type.(*client.ChatTypePrivate)
// 	return ok
// }

// // IsGroup проверяет, является ли чат группой (базовой или супергруппой, но не каналом)
// func (s *ChatService) IsGroup(chat *entity.Chat) bool {
// 	if chat == nil || chat.Type == nil {
// 		return false
// 	}

// 	if _, ok := chat.Type.(*client.ChatTypeBasicGroup); ok {
// 		return true
// 	}

// 	if supergroup, ok := chat.Type.(*client.ChatTypeSupergroup); ok && !supergroup.IsChannel {
// 		return true
// 	}

// 	return false
// }

// // IsChannel проверяет, является ли чат каналом
// func (s *ChatService) IsChannel(chat *entity.Chat) bool {
// 	if chat == nil || chat.Type == nil {
// 		return false
// 	}

// 	if supergroup, ok := chat.Type.(*client.ChatTypeSupergroup); ok && supergroup.IsChannel {
// 		return true
// 	}

// 	return false
// }

// // GetTypeName возвращает строковое представление типа чата
// func (s *ChatService) GetTypeName(chat *entity.Chat) string {
// 	if chat == nil || chat.Type == nil {
// 		return "unknown"
// 	}

// 	switch chat.Type.(type) {
// 	case *client.ChatTypePrivate:
// 		return "private"
// 	case *client.ChatTypeBasicGroup:
// 		return "basic_group"
// 	case *client.ChatTypeSupergroup:
// 		if chat.Type.(*client.ChatTypeSupergroup).IsChannel {
// 			return "channel"
// 		}
// 		return "supergroup"
// 	default:
// 		return "other"
// 	}
// }

// // CanSendMessages проверяет, можно ли отправлять сообщения в чат
// func (s *ChatService) CanSendMessages(chat *entity.Chat) bool {
// 	if chat == nil {
// 		return false
// 	}

// 	// В текущей версии TDLib проверяем доступность по свойству чата
// 	// Проверка разрешений может быть добавлена при наличии соответствующего поля в API

// 	return true
// }
