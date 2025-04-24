package bot // TODO: under construction

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/zelenin/go-tdlib/client"

	"github.com/comerc/budva43/config"
	"github.com/comerc/budva43/entity"
)

type messageController interface {
	GetMessage(chatID, messageID int64) (*client.Message, error)
	SendMessage(chatID int64, text string) (*client.Message, error)
	// DeleteMessage(chatID, messageID int64) error
	// EditMessage(chatID, messageID int64, text string) (*client.Message, error)
	// FormatMessage(text, fromFormat, toFormat string) (string, error)
	GetMessageText(message *client.Message) string
}

type forwardController interface {
	GetForwardRule(id string) (*entity.ForwardRule, error)
	SaveForwardRule(rule *entity.ForwardRule) error
	ForwardMessage(rule *entity.ForwardRule, fromChatID, messageID int64) ([]*client.Message, error)
}

type reportController interface {
	GenerateActivityReport(startDate, endDate time.Time) (*entity.ActivityReport, error)
	GenerateForwardingReport(startDate, endDate time.Time) (*entity.ForwardingReport, error)
	GenerateErrorReport(startDate, endDate time.Time) (*entity.ErrorReport, error)
}

// Transport представляет обработчик сообщений из Telegram
type Transport struct {
	log *slog.Logger
	//
	messageController messageController
	forwardController forwardController
	reportController  reportController
	updates           chan client.Update
	stopped           bool // TODO: зачем этот флаг, если есть ctx.Done()?
}

// New создает новый экземпляр обработчика Telegram
func New(
	messageController messageController,
	forwardController forwardController,
	reportController reportController,
) *Transport {
	return &Transport{
		log: slog.With("module", "transport.telegram"),
		//
		messageController: messageController,
		forwardController: forwardController,
		reportController:  reportController,
		updates:           make(chan client.Update, 100),
		stopped:           false,
	}
}

// Start запускает обработчик сообщений
func (t *Transport) Start(ctx context.Context) error {
	// Запускаем горутину обработки входящих обновлений
	go func() {
		for update := range t.updates {
			t.processUpdate(update)
		}
	}()

	// Ожидаем сигнал остановки через контекст
	<-ctx.Done() // TODO: это неверно для метода Start()
	return nil
}

// Close останавливает обработчик сообщений
func (t *Transport) Close() error {
	if t.stopped {
		return nil
	}

	t.stopped = true
	close(t.updates)
	return nil
}

// ReceiveUpdate получает обновление от клиента Telegram
func (t *Transport) ReceiveUpdate(update client.Update) {
	if t.stopped {
		return
	}

	// Отправляем обновление в канал для асинхронной обработки
	t.updates <- update
}

// processUpdate обрабатывает полученное обновление
func (t *Transport) processUpdate(update client.Update) {
	// В текущей реализации go-tdlib обычно приходят конкретные типы обновлений
	// Поэтому для простоты используем типовое переключение для определения типа обновления
	switch updateType := update.(type) {
	case *client.UpdateNewMessage:
		t.handleNewMessage(updateType.Message)
	default:
		// Другие типы обновлений не обрабатываем
	}
}

// handleNewMessage обрабатывает новое сообщение
func (t *Transport) handleNewMessage(message *client.Message) {
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
		t.processCommand(message, text)
		return
	}

	// Если сообщение не является командой, просто логируем его
	t.log.Info("Message received", "chat_id", message.ChatId, "text", text)
}

// processCommand обрабатывает команду пользователя
func (t *Transport) processCommand(message *client.Message, text string) {
	// Разбиваем команду на части
	args := strings.Fields(text)
	if len(args) == 0 {
		return
	}

	command := args[0]
	chatID := message.ChatId

	// Проверяем, что команда от администратора, если требуется
	isAdmin := chatID == config.Bot.AdminChatId

	// Обрабатываем различные команды
	switch command {
	case "/start", "/help":
		t.sendHelpMessage(chatID)

	case "/status":
		t.sendStatusMessage(chatID)

	case "/report":
		if !isAdmin {
			t.sendMessage(chatID, "Эта команда доступна только администратору.")
			return
		}
		if len(args) < 2 {
			t.sendMessage(chatID, "Использование: /report [activity|forwarding|error]")
			return
		}
		t.generateReport(chatID, args[1])

	case "/rule":
		if !isAdmin {
			t.sendMessage(chatID, "Эта команда доступна только администратору.")
			return
		}
		if len(args) < 2 {
			t.sendMessage(chatID, "Использование: /rule [list|show|add|delete] ...")
			return
		}
		t.handleRuleCommand(chatID, args[1:])

	default:
		t.sendMessage(chatID, "Неизвестная команда. Отправьте /help для получения списка доступных команд.")
	}
}

// sendHelpMessage отправляет справочное сообщение
func (t *Transport) sendHelpMessage(chatID int64) {
	helpText := `Доступные команды:
/help - показать это сообщение
/status - показать статус бота
/report [тип] - сгенерировать отчет (только для администратора)
/rule ... - управление правилами пересылки (только для администратора)`

	t.sendMessage(chatID, helpText)
}

// sendStatusMessage отправляет сообщение о текущем статусе бота
func (t *Transport) sendStatusMessage(chatID int64) {
	statusText := "Бот работает в штатном режиме."
	t.sendMessage(chatID, statusText)
}

// generateReport генерирует отчет заданного типа
func (t *Transport) generateReport(chatID int64, reportType string) {
	// Получаем даты для отчета (последняя неделя)
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -7)

	// Генерируем отчет в зависимости от типа
	switch reportType {
	case "activity":
		_, err := t.reportController.GenerateActivityReport(startDate, endDate)
		if err != nil {
			t.sendMessage(chatID, fmt.Sprintf("Ошибка при генерации отчета об активности: %v", err))
			return
		}
		t.sendMessage(chatID, fmt.Sprintf("Отчет об активности сгенерирован. Период: %s - %s",
			startDate.Format("02.01.2006"), endDate.Format("02.01.2006")))

	case "forwarding":
		_, err := t.reportController.GenerateForwardingReport(startDate, endDate)
		if err != nil {
			t.sendMessage(chatID, fmt.Sprintf("Ошибка при генерации отчета о пересылке: %v", err))
			return
		}
		t.sendMessage(chatID, fmt.Sprintf("Отчет о пересылке сгенерирован. Период: %s - %s",
			startDate.Format("02.01.2006"), endDate.Format("02.01.2006")))

	case "error":
		_, err := t.reportController.GenerateErrorReport(startDate, endDate)
		if err != nil {
			t.sendMessage(chatID, fmt.Sprintf("Ошибка при генерации отчета об ошибках: %v", err))
			return
		}
		t.sendMessage(chatID, fmt.Sprintf("Отчет об ошибках сгенерирован. Период: %s - %s",
			startDate.Format("02.01.2006"), endDate.Format("02.01.2006")))

	default:
		t.sendMessage(chatID, fmt.Sprintf("Неизвестный тип отчета: %s. Доступные типы: activity, forwarding, error", reportType))
		return
	}

	// Логируем успешную генерацию отчета
	t.log.Info("Report generated",
		"type", reportType,
		"start_date", startDate.Format("2006-01-02"),
		"end_date", endDate.Format("2006-01-02"))
}

// handleRuleCommand обрабатывает команды для управления правилами пересылки
func (t *Transport) handleRuleCommand(chatID int64, args []string) {
	if len(args) == 0 {
		t.sendMessage(chatID, "Недостаточно аргументов для команды /rule")
		return
	}

	switch args[0] {
	case "list":
		t.sendMessage(chatID, "Список правил пересылки (не реализовано)")

	case "show":
		if len(args) < 2 {
			t.sendMessage(chatID, "Использование: /rule show [id правила]")
			return
		}
		ruleID := args[1]
		rule, err := t.forwardController.GetForwardRule(ruleID)
		if err != nil {
			t.sendMessage(chatID, fmt.Sprintf("Ошибка при получении правила: %v", err))
			return
		}

		ruleInfo := fmt.Sprintf("Правило #%s:\nОт: %d\nК: %v\nАктивно: %t",
			rule.ID, rule.From, rule.To, rule.Status == entity.RuleStatusActive)
		t.sendMessage(chatID, ruleInfo)

	default:
		t.sendMessage(chatID, "Неизвестная подкоманда для /rule. Доступные: list, show")
	}
}

// sendMessage отправляет текстовое сообщение в чат
func (t *Transport) sendMessage(chatID int64, text string) {
	// Отправляем сообщение через контроллер
	_, err := t.messageController.SendMessage(chatID, text)
	if err != nil {
		t.log.Error("Failed to send message", "chat_id", chatID, "err", err)
	}
}
