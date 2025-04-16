package message

import (
	"github.com/zelenin/go-tdlib/client"
)

// messageService определяет интерфейс сервиса сообщений, необходимый контроллеру
type messageService interface {
	GetText(message *client.Message) string
	GetCaption(message *client.Message) string
	IsTextMessage(message *client.Message) bool
	IsMediaMessage(message *client.Message) bool
	GetContentType(message *client.Message) string
	FormatMessageContent(text string, fromFormat, toFormat string) (string, error)
	ExtractMessageMetadata(message *client.Message) map[string]any
}

// Controller представляет контроллер для работы с сообщениями
type Controller struct {
	messageService messageService
}

// New создает новый экземпляр контроллера сообщений
func New(messageService messageService) *Controller {
	return &Controller{
		messageService: messageService,
	}
}

// GetMessage получает сообщение по идентификатору
func (c *Controller) GetMessage(chatID, messageID int64) (*client.Message, error) {
	// Получаем сообщение из репозитория и возвращаем его напрямую
	// return c.messageService.GetMessage(chatID, messageID)
	return nil, nil
}

// SendMessage отправляет новое сообщение
func (c *Controller) SendMessage(chatID int64, text string) (*client.Message, error) {
	// Отправляем сообщение через репозиторий
	// return c.messageService.SendMessage(chatID, text)
	return nil, nil
}

// DeleteMessage удаляет сообщение
func (c *Controller) DeleteMessage(chatID, messageID int64) error {
	// return c.messageService.DeleteMessage(chatID, messageID)
	return nil
}

// EditMessage редактирует сообщение
func (c *Controller) EditMessage(chatID, messageID int64, text string) (*client.Message, error) {
	// Редактируем сообщение через репозиторий
	// return c.messageService.EditMessage(chatID, messageID, text)
	return nil, nil
}

// FormatMessage форматирует текст сообщения
func (c *Controller) FormatMessage(text, fromFormat, toFormat string) (string, error) {
	return c.messageService.FormatMessageContent(text, fromFormat, toFormat)
}

// GetMessageText возвращает текст сообщения
func (c *Controller) GetMessageText(message *client.Message) string {
	return c.messageService.GetText(message)
}

// GetMessageCaption возвращает подпись медиа сообщения
func (c *Controller) GetMessageCaption(message *client.Message) string {
	return c.messageService.GetCaption(message)
}

// GetContentType возвращает тип содержимого сообщения
func (c *Controller) GetContentType(message *client.Message) string {
	return c.messageService.GetContentType(message)
}

// IsTextMessage проверяет, является ли сообщение текстовым
func (c *Controller) IsTextMessage(message *client.Message) bool {
	return c.messageService.IsTextMessage(message)
}

// IsMediaMessage проверяет, содержит ли сообщение медиа-контент
func (c *Controller) IsMediaMessage(message *client.Message) bool {
	return c.messageService.IsMediaMessage(message)
}

// ListMessages возвращает список сообщений
func (c *Controller) ListMessages(limit, offset int) ([]*client.Message, error) {
	// Заглушка для метода ListMessages
	// В реальной реализации здесь был бы код для получения списка сообщений
	// из репозитория
	return []*client.Message{}, nil
}
