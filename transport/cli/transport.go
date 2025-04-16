package cli

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/zelenin/go-tdlib/client"
	"golang.org/x/term"

	"github.com/comerc/budva43/config"
	"github.com/comerc/budva43/entity"
	"github.com/comerc/budva43/util"
	// "golang.org/x/term"
)

// TODO: не нравится, что нужно вводить auth для каждого последующего шага

// messageController определяет интерфейс контроллера сообщений для CLI
type messageController interface {
	GetMessage(chatID, messageID int64) (*client.Message, error)
	SendMessage(chatID int64, text string) (*client.Message, error)
	ListMessages(limit, offset int) ([]*client.Message, error)
	GetMessageText(message *client.Message) string
	IsTextMessage(message *client.Message) bool
	GetContentType(message *client.Message) string
}

// forwardController определяет интерфейс контроллера пересылок для CLI
type forwardController interface {
	GetForwardRule(id string) (*entity.ForwardRule, error)
	SaveForwardRule(rule *entity.ForwardRule) error
	// ListForwardRules() ([]*entity.ForwardRule, error)
	// DeleteForwardRule(id string) error
}

// reportController определяет интерфейс контроллера отчетов для CLI
type reportController interface {
	GenerateActivityReport(startDate, endDate time.Time) (*entity.ActivityReport, error)
	GenerateForwardingReport(startDate, endDate time.Time) (*entity.ForwardingReport, error)
	GenerateErrorReport(startDate, endDate time.Time) (*entity.ErrorReport, error)
}

// authTelegramController определяет интерфейс контроллера авторизации Telegram для CLI
type authTelegramController interface {
	SubmitPhoneNumber(phone string)
	SubmitCode(code string)
	SubmitPassword(password string)
	InitClientDone() chan any
	GetAuthorizationState() (client.AuthorizationState, error)
}

// Transport представляет интерфейс командной строки
type Transport struct {
	messageController messageController
	forwardController forwardController
	reportController  reportController
	authController    authTelegramController
	scanner           *bufio.Scanner
	commands          []command
	commandMap        map[string]*command
}

// command представляет команду CLI
type command struct {
	name        string
	description string
	handler     func(args []string) error
}

// New создает новый экземпляр CLI
func New(
	messageController messageController,
	forwardController forwardController,
	reportController reportController,
	authController authTelegramController,
) *Transport {
	cli := &Transport{
		messageController: messageController,
		forwardController: forwardController,
		reportController:  reportController,
		authController:    authController,
		scanner:           bufio.NewScanner(os.Stdin),
		commands:          []command{},
	}

	// Регистрация команд
	cli.registerCommands()

	return cli
}

// registerCommands регистрирует доступные команды
func (t *Transport) registerCommands() {
	t.commands = []command{
		{
			name:        "help",
			description: "Показать список доступных команд",
			handler:     t.handleHelp,
		},
		{
			name:        "exit",
			description: "Выйти из программы",
			handler:     t.handleExit,
		},
		{
			name:        "messages",
			description: "Управление сообщениями: list, get, send",
			handler:     t.handleMessages,
		},
		{
			name:        "rules",
			description: "Управление правилами пересылки: list, get, add, delete",
			handler:     t.handleRules,
		},
		{
			name:        "report",
			description: "Генерация отчетов: activity, forwarding, error",
			handler:     t.handleReport,
		},
		{
			name:        "auth",
			description: "Запустить процесс авторизации в Telegram",
			handler:     t.handleAuth,
		},
	}

	t.commandMap = make(map[string]*command)
	for _, cmd := range t.commands {
		t.commandMap[cmd.name] = &cmd
	}
}

// Start запускает CLI интерфейс
func (t *Transport) Start(ctx context.Context, shutdown func()) error {
	// Запускаем обработку ввода в отдельной горутине
	go func() {

		select {
		case <-ctx.Done():
			return
		case <-t.authController.InitClientDone():
			// Если пришло какое-либо состояние, то TDLib клиент готов
			slog.Info("TDLib клиент готов")
		}

		fmt.Println("Запуск CLI интерфейса. Введите 'help' для просмотра доступных команд.")

		for {
			select {
			case <-ctx.Done():
				return
			default:
				fmt.Println("> ")
				if !t.scanner.Scan() {
					return
				}

				input := t.scanner.Text()

				if err := t.processCommand(input); err != nil {
					if err.Error() == "exit" {
						shutdown()
						slog.Info("Exit command processed")
						return
					}
					slog.Error("Command execution failed", "err", err)
					fmt.Printf("Ошибка: %v\n", err)
				}
			}
		}
	}()

	return nil
}

func (t *Transport) Stop() error {
	return nil
}

// processCommand обрабатывает введенную команду
func (t *Transport) processCommand(input string) error {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return nil
	}

	cmd := parts[0]
	var args []string
	if len(parts) > 1 {
		args = parts[1:]
	}

	if command, ok := t.commandMap[cmd]; ok {
		return command.handler(args)
	}

	slog.Info("Unknown command", "command", cmd)
	fmt.Printf("Неизвестная команда: %s. Введите 'help' для просмотра доступных команд.\n", cmd)
	return nil
}

// handleHelp обрабатывает команду help
func (t *Transport) handleHelp(args []string) error {
	fmt.Println("Доступные команды:")
	for _, cmd := range t.commands {
		fmt.Printf("  %-15s - %s\n", cmd.name, cmd.description)
	}
	return nil
}

// handleExit обрабатывает команду exit
func (t *Transport) handleExit(args []string) error {
	fmt.Println("Выход из программы...")
	return fmt.Errorf("exit")
}

// handleMessages обрабатывает команду messages
func (t *Transport) handleMessages(args []string) error {
	if len(args) == 0 {
		fmt.Println("Использование: messages [list|get|send] ...")
		return nil
	}

	switch args[0] {
	case "list":
		return t.handleMessageList()
	case "get":
		if len(args) < 3 {
			fmt.Println("Использование: messages get [chat_id] [message_id]")
			return nil
		}
		return t.handleMessageGet(args[1], args[2])
	case "send":
		if len(args) < 3 {
			fmt.Println("Использование: messages send [chat_id] [текст сообщения]")
			return nil
		}
		return t.handleMessageSend(args[1], strings.Join(args[2:], " "))
	default:
		fmt.Printf("Неизвестная подкоманда: %s. Доступные: list, get, send\n", args[0])
		return nil
	}
}

// handleMessageList обрабатывает команду messages list
func (t *Transport) handleMessageList() error {
	messages, err := t.messageController.ListMessages(10, 0)
	if err != nil {
		return fmt.Errorf("ошибка при получении списка сообщений: %w", err)
	}

	if len(messages) == 0 {
		fmt.Println("Список сообщений пуст")
		return nil
	}

	fmt.Println("Список последних сообщений:")
	for i, msg := range messages {
		text := t.messageController.GetMessageText(msg)
		fmt.Printf("%d. Чат: %d, ID: %d, Текст: %s\n", i+1, msg.ChatId, msg.Id, text)
	}

	return nil
}

// handleMessageGet обрабатывает команду messages get
func (t *Transport) handleMessageGet(chatIDStr, messageIDStr string) error {
	var chatID, messageID int64
	if _, err := fmt.Sscanf(chatIDStr, "%d", &chatID); err != nil {
		return fmt.Errorf("неверный формат chat_id: %w", err)
	}
	if _, err := fmt.Sscanf(messageIDStr, "%d", &messageID); err != nil {
		return fmt.Errorf("неверный формат message_id: %w", err)
	}

	message, err := t.messageController.GetMessage(chatID, messageID)
	if err != nil {
		return fmt.Errorf("ошибка при получении сообщения: %w", err)
	}

	text := t.messageController.GetMessageText(message)
	fmt.Printf("Сообщение:\nID: %d\nЧат: %d\nТекст: %s\n",
		message.Id, message.ChatId, text)
	return nil
}

// handleMessageSend обрабатывает команду messages send
func (t *Transport) handleMessageSend(chatIDStr, text string) error {
	var chatID int64
	if _, err := fmt.Sscanf(chatIDStr, "%d", &chatID); err != nil {
		return fmt.Errorf("неверный формат chat_id: %w", err)
	}

	message, err := t.messageController.SendMessage(chatID, text)
	if err != nil {
		return fmt.Errorf("ошибка при отправке сообщения: %w", err)
	}

	fmt.Printf("Сообщение отправлено:\nID: %d\nЧат: %d\nТекст: %s\n",
		message.Id, message.ChatId, t.messageController.GetMessageText(message))
	return nil
}

// handleRules обрабатывает команду rules
func (t *Transport) handleRules(args []string) error {
	if len(args) == 0 {
		fmt.Println("Использование: rules [list|get|add|delete] ...")
		return nil
	}

	switch args[0] {
	case "list":
		return t.handleRulesList()
	case "get":
		if len(args) < 2 {
			fmt.Println("Использование: rules get [id]")
			return nil
		}
		return t.handleRuleGet(args[1])
	case "add":
		if len(args) < 4 {
			fmt.Println("Использование: rules add [from_chat_id] [to_chat_id] [active=true|false]")
			return nil
		}
		return t.handleRuleAdd(args[1], args[2], args[3])
	case "delete":
		if len(args) < 2 {
			fmt.Println("Использование: rules delete [id]")
			return nil
		}
		return t.handleRuleDelete(args[1])
	default:
		fmt.Printf("Неизвестная подкоманда: %s. Доступные: list, get, add, delete\n", args[0])
		return nil
	}
}

// handleRulesList обрабатывает команду rules list
func (t *Transport) handleRulesList() error {
	// rules, err := t.forwardController.ListForwardRules()
	// if err != nil {
	// 	return fmt.Errorf("ошибка при получении списка правил: %w", err)
	// }

	// if len(rules) == 0 {
	// 	fmt.Println("Список правил пересылки пуст")
	// 	return nil
	// }

	// fmt.Println("Список правил пересылки:")
	// for i, rule := range rules {
	// 	fmt.Printf("%d. ID: %s, От: %d, К: %v, Активно: %t\n",
	// 		i+1, rule.ID, rule.From, rule.To, rule.Status == entity.RuleStatusActive)
	// }

	return nil
}

// handleRuleGet обрабатывает команду rules get
func (t *Transport) handleRuleGet(id string) error {
	rule, err := t.forwardController.GetForwardRule(id)
	if err != nil {
		return fmt.Errorf("ошибка при получении правила: %w", err)
	}

	fmt.Printf("Правило:\nID: %s\nОт: %d\nК: %v\nАктивно: %t\n",
		rule.ID, rule.From, rule.To, rule.Status == entity.RuleStatusActive)
	return nil
}

// handleRuleAdd обрабатывает команду rules add
func (t *Transport) handleRuleAdd(fromStr, toStr, activeStr string) error {
	var from int64
	if _, err := fmt.Sscanf(fromStr, "%d", &from); err != nil {
		return fmt.Errorf("неверный формат from_chat_id: %w", err)
	}

	var to int64
	if _, err := fmt.Sscanf(toStr, "%d", &to); err != nil {
		return fmt.Errorf("неверный формат to_chat_id: %w", err)
	}

	active := activeStr == "true"
	status := entity.RuleStatusInactive
	if active {
		status = entity.RuleStatusActive
	}

	rule := &entity.ForwardRule{
		From:   from,
		To:     []int64{to},
		Status: status,
	}

	if err := t.forwardController.SaveForwardRule(rule); err != nil {
		return fmt.Errorf("ошибка при добавлении правила: %w", err)
	}

	fmt.Println("Правило пересылки успешно добавлено")
	return nil
}

// handleRuleDelete обрабатывает команду rules delete
func (t *Transport) handleRuleDelete(id string) error {
	// if err := t.forwardController.DeleteForwardRule(id); err != nil {
	// 	return fmt.Errorf("ошибка при удалении правила: %w", err)
	// }

	// fmt.Println("Правило пересылки успешно удалено")
	return nil
}

// handleReport обрабатывает команду report
func (t *Transport) handleReport(args []string) error {
	if len(args) == 0 {
		fmt.Println("Использование: report [activity|forwarding|error] [days=7]")
		return nil
	}

	reportType := args[0]
	days := 7
	if len(args) > 1 {
		if _, err := fmt.Sscanf(args[1], "%d", &days); err != nil {
			fmt.Println("Используется период по умолчанию (7 дней)")
		}
	}

	// Получаем даты для отчета
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)

	fmt.Printf("Генерация отчета '%s' за период %s - %s...\n",
		reportType, startDate.Format("02.01.2006"), endDate.Format("02.01.2006"))

	var err error
	switch reportType {
	case "activity":
		_, err = t.reportController.GenerateActivityReport(startDate, endDate)
	case "forwarding":
		_, err = t.reportController.GenerateForwardingReport(startDate, endDate)
	case "error":
		_, err = t.reportController.GenerateErrorReport(startDate, endDate)
	default:
		return fmt.Errorf("неизвестный тип отчета: %s. Доступные: activity, forwarding, error", reportType)
	}

	if err != nil {
		return fmt.Errorf("ошибка при генерации отчета: %w", err)
	}

	fmt.Printf("Отчет '%s' успешно сгенерирован\n", reportType)
	return nil
}

// handleAuth обрабатывает команду auth
func (t *Transport) handleAuth(args []string) error {
	var err error

	state, err := t.authController.GetAuthorizationState()
	if err != nil {
		return fmt.Errorf("ошибка при получении состояния авторизации: %w", err)
	}

	slog.Debug("GetAuthorizationState()", "state", state.AuthorizationStateType())

	switch state.AuthorizationStateType() {
	case client.TypeAuthorizationStateWaitPhoneNumber:

		var phoneNumber string
		if config.Telegram.PhoneNumber != "" {
			// TODO: перенести в authController
			phoneNumber = config.Telegram.PhoneNumber
			fmt.Println("Используется номер телефона из конфигурации:", util.MaskPhoneNumber(phoneNumber))
			time.Sleep(3 * time.Second)
		} else {
			fmt.Println("Введите номер телефона: ")
			phoneNumber, err = t.hiddenReadLine()
			if err != nil {
				return fmt.Errorf("ошибка при чтении телефона: %w", err)
			}
		}
		t.authController.SubmitPhoneNumber(string(phoneNumber))

	case client.TypeAuthorizationStateWaitCode:
		fmt.Println("Введите код подтверждения: ")
		code, err := t.hiddenReadLine()
		if err != nil {
			return fmt.Errorf("ошибка при чтении кода: %w", err)
		}
		t.authController.SubmitCode(string(code))

	case client.TypeAuthorizationStateWaitPassword:
		fmt.Println("Введите пароль: ")
		password, err := t.hiddenReadLine()
		if err != nil {
			return fmt.Errorf("ошибка при чтении пароля: %w", err)
		}
		t.authController.SubmitPassword(string(password))

	case client.TypeAuthorizationStateReady:
		fmt.Println("Авторизация в Telegram успешно завершена!")
	}

	return nil
}

func (t *Transport) hiddenReadLine() (string, error) {
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	return string(password), err
}
