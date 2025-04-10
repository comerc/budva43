package forward

import (
	"github.com/comerc/budva43/entity"
	"github.com/zelenin/go-tdlib/client"
)

// forwardRuleService определяет интерфейс сервиса правил пересылки, необходимый контроллеру
type forwardRuleService interface {
	CompileRegexps(rule *entity.ForwardRule) error
	ShouldForward(rule *entity.ForwardRule, text string) bool
}

// messageService определяет интерфейс сервиса сообщений, необходимый контроллеру
type messageService interface {
	GetText(message *client.Message) string
	GetCaption(message *client.Message) string
}

// telegramRepo определяет интерфейс репозитория Telegram, необходимый контроллеру
type telegramRepo interface {
	GetMessage(chatID, messageID int64) (*client.Message, error)
	ForwardMessage(fromChatID, messageID int64, toChatID int64) (*client.Message, error)
	SendMessage(chatID int64, text string) (*client.Message, error)
}

// storageRepo определяет интерфейс репозитория хранилища, необходимый контроллеру
type storageRepo interface {
	Get(key []byte) ([]byte, error)
	Set(key, value []byte) error
}

// Controller представляет контроллер для работы с пересылкой сообщений
type Controller struct {
	forwardRuleService forwardRuleService
	messageService     messageService
	telegramRepo       telegramRepo
	storageRepo        storageRepo
}

// New создает новый экземпляр контроллера пересылки
func New(
	forwardRuleService forwardRuleService,
	messageService messageService,
	telegramRepo telegramRepo,
	storageRepo storageRepo,
) *Controller {
	return &Controller{
		forwardRuleService: forwardRuleService,
		messageService:     messageService,
		telegramRepo:       telegramRepo,
		storageRepo:        storageRepo,
	}
}

// ForwardMessage пересылает сообщение согласно правилу
func (c *Controller) ForwardMessage(
	rule *entity.ForwardRule,
	fromChatID, messageID int64,
) ([]*client.Message, error) {
	// Получаем исходное сообщение
	message, err := c.telegramRepo.GetMessage(fromChatID, messageID)
	if err != nil {
		return nil, err
	}

	// Получаем текст сообщения (или подпись для медиа)
	text := c.messageService.GetText(message)
	if text == "" {
		text = c.messageService.GetCaption(message)
	}

	// Проверяем, нужно ли пересылать сообщение по правилу
	if !c.forwardRuleService.ShouldForward(rule, text) {
		return nil, nil
	}

	// Пересылаем сообщение во все чаты-получатели
	result := make([]*client.Message, 0, len(rule.To))
	for _, toChatID := range rule.To {
		var forwardedMessage *client.Message
		var err error

		if rule.SendCopy {
			// Отправляем копию текста
			forwardedMessage, err = c.telegramRepo.SendMessage(toChatID, text)
		} else {
			// Пересылаем сообщение
			forwardedMessage, err = c.telegramRepo.ForwardMessage(fromChatID, messageID, toChatID)
		}

		if err != nil {
			continue
		}

		// Добавляем успешно пересланное сообщение в результат
		result = append(result, forwardedMessage)
	}

	return result, nil
}

// GetForwardRule получает правило пересылки по идентификатору
func (c *Controller) GetForwardRule(id string) (*entity.ForwardRule, error) {
	// Получаем правило из хранилища (реализация десериализации должна быть добавлена)
	// Здесь просто заглушка для примера
	return &entity.ForwardRule{
		ID:     id,
		Status: entity.RuleStatusActive,
	}, nil
}

// SaveForwardRule сохраняет правило пересылки
func (c *Controller) SaveForwardRule(rule *entity.ForwardRule) error {
	// Компилируем регулярные выражения
	if err := c.forwardRuleService.CompileRegexps(rule); err != nil {
		return err
	}

	// Сохраняем правило в хранилище (реализация сериализации должна быть добавлена)
	// Здесь просто заглушка для примера
	return nil
}
