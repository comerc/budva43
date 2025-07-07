package telegram

import (
	"github.com/comerc/budva43/app/log"
	"github.com/zelenin/go-tdlib/client"
)

// clientAdapter - tdlibClient methods (для моков в юнит-тестах)
type clientAdapter interface {
	// Message operations
	GetMessage(*client.GetMessageRequest) (*client.Message, error)
	SendMessage(*client.SendMessageRequest) (*client.Message, error)
	SendMessageAlbum(*client.SendMessageAlbumRequest) (*client.Messages, error)
	EditMessageText(*client.EditMessageTextRequest) (*client.Message, error)
	EditMessageCaption(*client.EditMessageCaptionRequest) (*client.Message, error)
	DeleteMessages(*client.DeleteMessagesRequest) (*client.Ok, error)
	// GetMessages(*client.GetMessagesRequest) (*client.Messages, error)

	// Forward operations
	ForwardMessages(*client.ForwardMessagesRequest) (*client.Messages, error)

	// Link operations
	GetMessageLink(*client.GetMessageLinkRequest) (*client.MessageLink, error)
	GetMessageLinkInfo(*client.GetMessageLinkInfoRequest) (*client.MessageLinkInfo, error)

	// Chat operations
	LoadChats(*client.LoadChatsRequest) (*client.Ok, error)
	GetChatHistory(*client.GetChatHistoryRequest) (*client.Messages, error)
	GetChat(*client.GetChatRequest) (*client.Chat, error)
	// GetChats(*client.GetChatsRequest) (*client.Chats, error)
	// GetChatMessageCount(*client.GetChatMessageCountRequest) (*client.Count, error)

	// Other operations
	GetListener() *client.Listener
	ParseTextEntities(*client.ParseTextEntitiesRequest) (*client.FormattedText, error)
	GetCallbackQueryAnswer(*client.GetCallbackQueryAnswerRequest) (*client.CallbackQueryAnswer, error)
	GetOption(*client.GetOptionRequest) (client.OptionValue, error)
	GetMe() (*client.User, error)
}

// GetMessage выводит информацию о сообщении
func (r *Repo) GetMessage(req *client.GetMessageRequest) (*client.Message, error) {
	msg, err := r.getClient().GetMessage(req)
	if err != nil {
		return nil, log.WrapError(err) // внешняя ошибка
	}
	return msg, nil
}

// SendMessage отправляет сообщение
func (r *Repo) SendMessage(req *client.SendMessageRequest) (*client.Message, error) {
	msg, err := r.getClient().SendMessage(req)
	if err != nil {
		return nil, log.WrapError(err) // внешняя ошибка
	}
	return msg, nil
}

// SendMessageAlbum отправляет альбом сообщений
func (r *Repo) SendMessageAlbum(req *client.SendMessageAlbumRequest) (*client.Messages, error) {
	messages, err := r.getClient().SendMessageAlbum(req)
	if err != nil {
		return nil, log.WrapError(err) // внешняя ошибка
	}
	return messages, nil
}

// EditMessageText редактирует текст сообщения
func (r *Repo) EditMessageText(req *client.EditMessageTextRequest) (*client.Message, error) {
	msg, err := r.getClient().EditMessageText(req)
	if err != nil {
		return nil, log.WrapError(err) // внешняя ошибка
	}
	return msg, nil
}

// EditMessageCaption редактирует подпись сообщения
func (r *Repo) EditMessageCaption(req *client.EditMessageCaptionRequest) (*client.Message, error) {
	msg, err := r.getClient().EditMessageCaption(req)
	if err != nil {
		return nil, log.WrapError(err) // внешняя ошибка
	}
	return msg, nil
}

// DeleteMessages удаляет сообщения
func (r *Repo) DeleteMessages(req *client.DeleteMessagesRequest) (*client.Ok, error) {
	ok, err := r.getClient().DeleteMessages(req)
	if err != nil {
		return nil, log.WrapError(err) // внешняя ошибка
	}
	return ok, nil
}

// ForwardMessages пересылает сообщения
func (r *Repo) ForwardMessages(req *client.ForwardMessagesRequest) (*client.Messages, error) {
	messages, err := r.getClient().ForwardMessages(req)
	if err != nil {
		return nil, log.WrapError(err) // внешняя ошибка
	}
	return messages, nil
}

// GetMessageLink выводит ссылку на сообщение
func (r *Repo) GetMessageLink(req *client.GetMessageLinkRequest) (*client.MessageLink, error) {
	link, err := r.getClient().GetMessageLink(req)
	if err != nil {
		return nil, log.WrapError(err) // внешняя ошибка
	}
	return link, nil
}

// GetMessageLinkInfo выводит информацию о ссылке на сообщение
func (r *Repo) GetMessageLinkInfo(req *client.GetMessageLinkInfoRequest) (*client.MessageLinkInfo, error) {
	info, err := r.getClient().GetMessageLinkInfo(req)
	if err != nil {
		return nil, log.WrapError(err) // внешняя ошибка
	}
	return info, nil
}

// LoadChats загружает чаты
func (r *Repo) LoadChats(req *client.LoadChatsRequest) (*client.Ok, error) {
	ok, err := r.getClient().LoadChats(req)
	if err != nil {
		return nil, log.WrapError(err) // внешняя ошибка
	}
	return ok, nil
}

// GetChatHistory выводит историю сообщений
func (r *Repo) GetChatHistory(req *client.GetChatHistoryRequest) (*client.Messages, error) {
	messages, err := r.getClient().GetChatHistory(req)
	if err != nil {
		return nil, log.WrapError(err) // внешняя ошибка
	}
	return messages, nil
}

// GetChat выводит информацию о чате
func (r *Repo) GetChat(req *client.GetChatRequest) (*client.Chat, error) {
	chat, err := r.getClient().GetChat(req)
	if err != nil {
		return nil, log.WrapError(err) // внешняя ошибка
	}
	return chat, nil
}

// GetListener возвращает слушателя TDLib
func (r *Repo) GetListener() *client.Listener {
	return r.getClient().GetListener()
}

// ParseTextEntities парсит текст сообщения
func (r *Repo) ParseTextEntities(req *client.ParseTextEntitiesRequest) (*client.FormattedText, error) {
	formattedText, err := r.getClient().ParseTextEntities(req)
	if err != nil {
		return nil, log.WrapError(err) // внешняя ошибка
	}
	return formattedText, nil
}

// GetCallbackQueryAnswer выводит информацию о ответе на callback-запрос
func (r *Repo) GetCallbackQueryAnswer(req *client.GetCallbackQueryAnswerRequest) (*client.CallbackQueryAnswer, error) {
	answer, err := r.getClient().GetCallbackQueryAnswer(req)
	if err != nil {
		return nil, log.WrapError(err) // внешняя ошибка
	}
	return answer, nil
}

// GetOption выводит информацию о параметрах TDLib
func (r *Repo) GetOption(req *client.GetOptionRequest) (client.OptionValue, error) {
	option, err := r.getClient().GetOption(req)
	if err != nil {
		return nil, log.WrapError(err) // внешняя ошибка
	}
	return option, nil
}

// GetMe выводит информацию о пользователе
func (r *Repo) GetMe() (*client.User, error) {
	user, err := r.getClient().GetMe()
	if err != nil {
		return nil, log.WrapError(err) // внешняя ошибка
	}
	return user, nil
}
