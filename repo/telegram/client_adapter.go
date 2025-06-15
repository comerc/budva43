package telegram

import "github.com/zelenin/go-tdlib/client"

// clientAdapter - tdlibClient methods (для моков в юнит-тестах)
type clientAdapter interface {
	// Message operations
	GetMessage(*client.GetMessageRequest) (*client.Message, error)
	SendMessage(*client.SendMessageRequest) (*client.Message, error)
	SendMessageAlbum(*client.SendMessageAlbumRequest) (*client.Messages, error)
	EditMessageText(*client.EditMessageTextRequest) (*client.Message, error)
	EditMessageCaption(*client.EditMessageCaptionRequest) (*client.Message, error)
	DeleteMessages(*client.DeleteMessagesRequest) (*client.Ok, error)

	// Forward operations
	ForwardMessages(*client.ForwardMessagesRequest) (*client.Messages, error)

	// Link operations
	GetMessageLink(*client.GetMessageLinkRequest) (*client.MessageLink, error)
	GetMessageLinkInfo(*client.GetMessageLinkInfoRequest) (*client.MessageLinkInfo, error)

	// Other operations
	GetListener() *client.Listener
	ParseTextEntities(*client.ParseTextEntitiesRequest) (*client.FormattedText, error)
	GetCallbackQueryAnswer(*client.GetCallbackQueryAnswerRequest) (*client.CallbackQueryAnswer, error)
	GetOption(*client.GetOptionRequest) (client.OptionValue, error)
	GetMe() (*client.User, error)
}

// GetMessage выводит информацию о сообщении
func (r *Repo) GetMessage(req *client.GetMessageRequest) (*client.Message, error) {
	return r.getClient().GetMessage(req)
}

// SendMessage отправляет сообщение
func (r *Repo) SendMessage(req *client.SendMessageRequest) (*client.Message, error) {
	return r.getClient().SendMessage(req)
}

// SendMessageAlbum отправляет альбом сообщений
func (r *Repo) SendMessageAlbum(req *client.SendMessageAlbumRequest) (*client.Messages, error) {
	return r.getClient().SendMessageAlbum(req)
}

// EditMessageText редактирует текст сообщения
func (r *Repo) EditMessageText(req *client.EditMessageTextRequest) (*client.Message, error) {
	return r.getClient().EditMessageText(req)
}

// EditMessageCaption редактирует подпись сообщения
func (r *Repo) EditMessageCaption(req *client.EditMessageCaptionRequest) (*client.Message, error) {
	return r.getClient().EditMessageCaption(req)
}

// DeleteMessages удаляет сообщения
func (r *Repo) DeleteMessages(req *client.DeleteMessagesRequest) (*client.Ok, error) {
	return r.getClient().DeleteMessages(req)
}

// ForwardMessages пересылает сообщения
func (r *Repo) ForwardMessages(req *client.ForwardMessagesRequest) (*client.Messages, error) {
	return r.getClient().ForwardMessages(req)
}

// GetMessageLink выводит ссылку на сообщение
func (r *Repo) GetMessageLink(req *client.GetMessageLinkRequest) (*client.MessageLink, error) {
	return r.getClient().GetMessageLink(req)
}

// GetMessageLinkInfo выводит информацию о ссылке на сообщение
func (r *Repo) GetMessageLinkInfo(req *client.GetMessageLinkInfoRequest) (*client.MessageLinkInfo, error) {
	return r.getClient().GetMessageLinkInfo(req)
}

// GetListener возвращает слушателя TDLib
func (r *Repo) GetListener() *client.Listener {
	return r.getClient().GetListener()
}

// ParseTextEntities парсит текст сообщения
func (r *Repo) ParseTextEntities(req *client.ParseTextEntitiesRequest) (*client.FormattedText, error) {
	return r.getClient().ParseTextEntities(req)
}

// GetCallbackQueryAnswer выводит информацию о ответе на callback-запрос
func (r *Repo) GetCallbackQueryAnswer(req *client.GetCallbackQueryAnswerRequest) (*client.CallbackQueryAnswer, error) {
	return r.getClient().GetCallbackQueryAnswer(req)
}

// GetOption выводит информацию о параметрах TDLib
func (r *Repo) GetOption(req *client.GetOptionRequest) (client.OptionValue, error) {
	return r.getClient().GetOption(req)
}

// GetMe выводит информацию о пользователе
func (r *Repo) GetMe() (*client.User, error) {
	return r.getClient().GetMe()
}
