package telegram

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/comerc/budva43/entity"
	"github.com/zelenin/go-tdlib/client"
)

// messageController определяет интерфейс контроллера сообщений, необходимый для Telegram транспорта
type messageController interface {
	GetMessage(chatID, messageID int64) (*client.Message, error)
	SendMessage(chatID int64, text string) (*client.Message, error)
	DeleteMessage(chatID, messageID int64) error
	EditMessage(chatID, messageID int64, text string) (*client.Message, error)
	FormatMessage(text, fromFormat, toFormat string) (string, error)
	GetMessageText(message *client.Message) string
}

// forwardController определяет интерфейс контроллера пересылок, необходимый для Telegram транспорта
type forwardController interface {
	GetForwardRule(id string) (*entity.ForwardRule, error)
	SaveForwardRule(rule *entity.ForwardRule) error
	ForwardMessage(rule interface{}, fromChatID, messageID int64) ([]*client.Message, error)
}

// reportController определяет интерфейс контроллера отчетов, необходимый для Telegram транспорта
type reportController interface {
	GenerateActivityReport(startDate, endDate time.Time) (interface{}, error)
	GenerateForwardingReport(startDate, endDate time.Time) (interface{}, error)
	GenerateErrorReport(startDate, endDate time.Time) (interface{}, error)
}

// telegramClient определяет интерфейс клиента Telegram, необходимый для обработчика
type telegramClient interface {
	GetMessage(chatID, messageID int64) (*client.Message, error)
	SendMessage(chatID int64, text string) (*client.Message, error)
	DeleteMessage(chatID, messageID int64) error
	EditMessage(chatID, messageID int64, text string) (*client.Message, error)
}

// Handler представляет обработчик сообщений из Telegram
type Handler struct {
	messageController messageController
	forwardController forwardController
	reportController  reportController
	telegramClient    telegramClient
	adminChatID       int64
	updates           chan client.Update
	stopped           bool
}

// NewHandler создает новый экземпляр обработчика Telegram
func NewHandler(
	messageController messageController,
	forwardController forwardController,
	reportController reportController,
	telegramClient telegramClient,
	adminChatID int64,
) *Handler {
	return &Handler{
		messageController: messageController,
		forwardController: forwardController,
		reportController:  reportController,
		telegramClient:    telegramClient,
		adminChatID:       adminChatID,
		updates:           make(chan client.Update, 100),
		stopped:           false,
	}
}

// Start запускает обработчик сообщений
func (h *Handler) Start(ctx context.Context) error {
	// Запускаем горутину обработки входящих обновлений
	go func() {
		for update := range h.updates {
			h.processUpdate(update)
		}
	}()

	// Ожидаем сигнал остановки через контекст
	<-ctx.Done()
	return h.Stop()
}

// Stop останавливает обработчик сообщений
func (h *Handler) Stop() error {
	if h.stopped {
		return nil
	}

	h.stopped = true
	close(h.updates)
	return nil
}

// ReceiveUpdate получает обновление от клиента Telegram
func (h *Handler) ReceiveUpdate(update client.Update) {
	if h.stopped {
		return
	}

	// Отправляем обновление в канал для асинхронной обработки
	h.updates <- update
}

// processUpdate обрабатывает полученное обновление
func (h *Handler) processUpdate(update client.Update) {
	// В текущей реализации go-tdlib обычно приходят конкретные типы обновлений
	// Поэтому для простоты используем типовое переключение для определения типа обновления
	switch updateType := update.(type) {
	case *client.UpdateNewMessage:
		h.handleNewMessage(updateType.Message)
	default:
		// Другие типы обновлений не обрабатываем
	}
}

// handleNewMessage обрабатывает новое сообщение
func (h *Handler) handleNewMessage(message *client.Message) {
	// Проверяем, что это сообщение от пользователя
	// Заглушка для проверки, в реальном приложении нужно получить отправителя
	isOutgoing := false // Заглушка

	if isOutgoing {
		return // Игнорируем исходящие сообщения
	}

	// Получаем текст сообщения
	var text string
	if content, ok := message.Content.(*client.MessageText); ok {
		text = content.Text.Text
	}

	// Если текст начинается с "/", обрабатываем как команду
	if strings.HasPrefix(text, "/") {
		h.processCommand(message, text)
		return
	}

	// Если сообщение не является командой, просто логируем его
	fmt.Printf("Received message from chat %d: %s\n", message.ChatId, text)
}

// processCommand обрабатывает команду пользователя
func (h *Handler) processCommand(message *client.Message, text string) {
	// Разбиваем команду на части
	args := strings.Fields(text)
	if len(args) == 0 {
		return
	}

	command := args[0]
	chatID := message.ChatId

	// Проверяем, что команда от администратора, если требуется
	isAdmin := chatID == h.adminChatID

	// Обрабатываем различные команды
	switch command {
	case "/start", "/help":
		h.sendHelpMessage(chatID)

	case "/status":
		h.sendStatusMessage(chatID)

	case "/report":
		if !isAdmin {
			h.sendMessage(chatID, "Эта команда доступна только администратору.")
			return
		}
		if len(args) < 2 {
			h.sendMessage(chatID, "Использование: /report [activity|forwarding|error]")
			return
		}
		h.generateReport(chatID, args[1])

	case "/rule":
		if !isAdmin {
			h.sendMessage(chatID, "Эта команда доступна только администратору.")
			return
		}
		if len(args) < 2 {
			h.sendMessage(chatID, "Использование: /rule [list|show|add|delete] ...")
			return
		}
		h.handleRuleCommand(chatID, args[1:])

	default:
		h.sendMessage(chatID, "Неизвестная команда. Отправьте /help для получения списка доступных команд.")
	}
}

// sendHelpMessage отправляет справочное сообщение
func (h *Handler) sendHelpMessage(chatID int64) {
	helpText := `Доступные команды:
/help - показать это сообщение
/status - показать статус бота
/report [тип] - сгенерировать отчет (только для администратора)
/rule ... - управление правилами пересылки (только для администратора)`

	h.sendMessage(chatID, helpText)
}

// sendStatusMessage отправляет сообщение о текущем статусе бота
func (h *Handler) sendStatusMessage(chatID int64) {
	statusText := "Бот работает в штатном режиме."
	h.sendMessage(chatID, statusText)
}

// generateReport генерирует отчет заданного типа
func (h *Handler) generateReport(chatID int64, reportType string) {
	// Получаем даты для отчета (последняя неделя)
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -7)

	// Генерируем отчет в зависимости от типа
	switch reportType {
	case "activity":
		_, err := h.reportController.GenerateActivityReport(startDate, endDate)
		if err != nil {
			h.sendMessage(chatID, fmt.Sprintf("Ошибка при генерации отчета об активности: %v", err))
			return
		}
		h.sendMessage(chatID, fmt.Sprintf("Отчет об активности сгенерирован. Период: %s - %s",
			startDate.Format("02.01.2006"), endDate.Format("02.01.2006")))

	case "forwarding":
		_, err := h.reportController.GenerateForwardingReport(startDate, endDate)
		if err != nil {
			h.sendMessage(chatID, fmt.Sprintf("Ошибка при генерации отчета о пересылке: %v", err))
			return
		}
		h.sendMessage(chatID, fmt.Sprintf("Отчет о пересылке сгенерирован. Период: %s - %s",
			startDate.Format("02.01.2006"), endDate.Format("02.01.2006")))

	case "error":
		_, err := h.reportController.GenerateErrorReport(startDate, endDate)
		if err != nil {
			h.sendMessage(chatID, fmt.Sprintf("Ошибка при генерации отчета об ошибках: %v", err))
			return
		}
		h.sendMessage(chatID, fmt.Sprintf("Отчет об ошибках сгенерирован. Период: %s - %s",
			startDate.Format("02.01.2006"), endDate.Format("02.01.2006")))

	default:
		h.sendMessage(chatID, fmt.Sprintf("Неизвестный тип отчета: %s. Доступные типы: activity, forwarding, error", reportType))
		return
	}

	// Логируем успешную генерацию отчета
	fmt.Printf("Generated %s report for period %s - %s\n",
		reportType, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
}

// handleRuleCommand обрабатывает команды для управления правилами пересылки
func (h *Handler) handleRuleCommand(chatID int64, args []string) {
	if len(args) == 0 {
		h.sendMessage(chatID, "Недостаточно аргументов для команды /rule")
		return
	}

	switch args[0] {
	case "list":
		h.sendMessage(chatID, "Список правил пересылки (не реализовано)")

	case "show":
		if len(args) < 2 {
			h.sendMessage(chatID, "Использование: /rule show [id правила]")
			return
		}
		ruleID := args[1]
		rule, err := h.forwardController.GetForwardRule(ruleID)
		if err != nil {
			h.sendMessage(chatID, fmt.Sprintf("Ошибка при получении правила: %v", err))
			return
		}

		ruleInfo := fmt.Sprintf("Правило #%s:\nОт: %d\nК: %v\nАктивно: %t",
			rule.ID, rule.From, rule.To, rule.Status == entity.RuleStatusActive)
		h.sendMessage(chatID, ruleInfo)

	default:
		h.sendMessage(chatID, "Неизвестная подкоманда для /rule. Доступные: list, show")
	}
}

// sendMessage отправляет текстовое сообщение в чат
func (h *Handler) sendMessage(chatID int64, text string) {
	// Отправляем сообщение через контроллер
	_, err := h.messageController.SendMessage(chatID, text)
	if err != nil {
		fmt.Printf("Error sending message to chat %d: %v\n", chatID, err)
	}
}
