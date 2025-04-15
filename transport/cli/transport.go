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

	"github.com/comerc/budva43/entity"
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
	cancel            context.CancelFunc
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
func (c *Transport) registerCommands() {
	c.commands = []command{
		{
			name:        "help",
			description: "Показать список доступных команд",
			handler:     c.handleHelp,
		},
		{
			name:        "exit",
			description: "Выйти из программы",
			handler:     c.handleExit,
		},
		{
			name:        "messages",
			description: "Управление сообщениями: list, get, send",
			handler:     c.handleMessages,
		},
		{
			name:        "rules",
			description: "Управление правилами пересылки: list, get, add, delete",
			handler:     c.handleRules,
		},
		{
			name:        "report",
			description: "Генерация отчетов: activity, forwarding, error",
			handler:     c.handleReport,
		},
		{
			name:        "auth",
			description: "Запустить процесс авторизации в Telegram",
			handler:     c.handleAuth,
		},
	}

	c.commandMap = make(map[string]*command)
	for _, cmd := range c.commands {
		c.commandMap[cmd.name] = &cmd
	}
}

// Start запускает CLI интерфейс
func (c *Transport) Start(ctx context.Context, cancel context.CancelFunc) error {
	c.cancel = cancel

	// Запускаем обработку ввода в отдельной горутине
	go func() {

		select {
		case <-ctx.Done():
			return
		case <-c.authController.InitClientDone():
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
				if !c.scanner.Scan() {
					return
				}

				input := c.scanner.Text()

				if err := c.processCommand(input); err != nil {
					if err.Error() == "exit" {
						cancel()
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

func (c *Transport) Stop() error {
	c.cancel() // TODO: может лучше вызывать interrupt, чтобы не тащить cancel?
	return nil
}

// processCommand обрабатывает введенную команду
func (c *Transport) processCommand(input string) error {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return nil
	}

	cmd := parts[0]
	var args []string
	if len(parts) > 1 {
		args = parts[1:]
	}

	if command, ok := c.commandMap[cmd]; ok {
		return command.handler(args)
	}

	slog.Info("Unknown command", "command", cmd)
	fmt.Printf("Неизвестная команда: %s. Введите 'help' для просмотра доступных команд.\n", cmd)
	return nil
}

// handleHelp обрабатывает команду help
func (c *Transport) handleHelp(args []string) error {
	fmt.Println("Доступные команды:")
	for _, cmd := range c.commands {
		fmt.Printf("  %-15s - %s\n", cmd.name, cmd.description)
	}
	return nil
}

// handleExit обрабатывает команду exit
func (c *Transport) handleExit(args []string) error {
	fmt.Println("Выход из программы...")
	return fmt.Errorf("exit")
}

// handleMessages обрабатывает команду messages
func (c *Transport) handleMessages(args []string) error {
	if len(args) == 0 {
		fmt.Println("Использование: messages [list|get|send] ...")
		return nil
	}

	switch args[0] {
	case "list":
		return c.handleMessageList()
	case "get":
		if len(args) < 3 {
			fmt.Println("Использование: messages get [chat_id] [message_id]")
			return nil
		}
		return c.handleMessageGet(args[1], args[2])
	case "send":
		if len(args) < 3 {
			fmt.Println("Использование: messages send [chat_id] [текст сообщения]")
			return nil
		}
		return c.handleMessageSend(args[1], strings.Join(args[2:], " "))
	default:
		fmt.Printf("Неизвестная подкоманда: %s. Доступные: list, get, send\n", args[0])
		return nil
	}
}

// handleMessageList обрабатывает команду messages list
func (c *Transport) handleMessageList() error {
	messages, err := c.messageController.ListMessages(10, 0)
	if err != nil {
		return fmt.Errorf("ошибка при получении списка сообщений: %w", err)
	}

	if len(messages) == 0 {
		fmt.Println("Список сообщений пуст")
		return nil
	}

	fmt.Println("Список последних сообщений:")
	for i, msg := range messages {
		text := c.messageController.GetMessageText(msg)
		fmt.Printf("%d. Чат: %d, ID: %d, Текст: %s\n", i+1, msg.ChatId, msg.Id, text)
	}

	return nil
}

// handleMessageGet обрабатывает команду messages get
func (c *Transport) handleMessageGet(chatIDStr, messageIDStr string) error {
	var chatID, messageID int64
	if _, err := fmt.Sscanf(chatIDStr, "%d", &chatID); err != nil {
		return fmt.Errorf("неверный формат chat_id: %w", err)
	}
	if _, err := fmt.Sscanf(messageIDStr, "%d", &messageID); err != nil {
		return fmt.Errorf("неверный формат message_id: %w", err)
	}

	message, err := c.messageController.GetMessage(chatID, messageID)
	if err != nil {
		return fmt.Errorf("ошибка при получении сообщения: %w", err)
	}

	text := c.messageController.GetMessageText(message)
	fmt.Printf("Сообщение:\nID: %d\nЧат: %d\nТекст: %s\n",
		message.Id, message.ChatId, text)
	return nil
}

// handleMessageSend обрабатывает команду messages send
func (c *Transport) handleMessageSend(chatIDStr, text string) error {
	var chatID int64
	if _, err := fmt.Sscanf(chatIDStr, "%d", &chatID); err != nil {
		return fmt.Errorf("неверный формат chat_id: %w", err)
	}

	message, err := c.messageController.SendMessage(chatID, text)
	if err != nil {
		return fmt.Errorf("ошибка при отправке сообщения: %w", err)
	}

	fmt.Printf("Сообщение отправлено:\nID: %d\nЧат: %d\nТекст: %s\n",
		message.Id, message.ChatId, c.messageController.GetMessageText(message))
	return nil
}

// handleRules обрабатывает команду rules
func (c *Transport) handleRules(args []string) error {
	if len(args) == 0 {
		fmt.Println("Использование: rules [list|get|add|delete] ...")
		return nil
	}

	switch args[0] {
	case "list":
		return c.handleRulesList()
	case "get":
		if len(args) < 2 {
			fmt.Println("Использование: rules get [id]")
			return nil
		}
		return c.handleRuleGet(args[1])
	case "add":
		if len(args) < 4 {
			fmt.Println("Использование: rules add [from_chat_id] [to_chat_id] [active=true|false]")
			return nil
		}
		return c.handleRuleAdd(args[1], args[2], args[3])
	case "delete":
		if len(args) < 2 {
			fmt.Println("Использование: rules delete [id]")
			return nil
		}
		return c.handleRuleDelete(args[1])
	default:
		fmt.Printf("Неизвестная подкоманда: %s. Доступные: list, get, add, delete\n", args[0])
		return nil
	}
}

// handleRulesList обрабатывает команду rules list
func (c *Transport) handleRulesList() error {
	// rules, err := c.forwardController.ListForwardRules()
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
func (c *Transport) handleRuleGet(id string) error {
	rule, err := c.forwardController.GetForwardRule(id)
	if err != nil {
		return fmt.Errorf("ошибка при получении правила: %w", err)
	}

	fmt.Printf("Правило:\nID: %s\nОт: %d\nК: %v\nАктивно: %t\n",
		rule.ID, rule.From, rule.To, rule.Status == entity.RuleStatusActive)
	return nil
}

// handleRuleAdd обрабатывает команду rules add
func (c *Transport) handleRuleAdd(fromStr, toStr, activeStr string) error {
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

	if err := c.forwardController.SaveForwardRule(rule); err != nil {
		return fmt.Errorf("ошибка при добавлении правила: %w", err)
	}

	fmt.Println("Правило пересылки успешно добавлено")
	return nil
}

// handleRuleDelete обрабатывает команду rules delete
func (c *Transport) handleRuleDelete(id string) error {
	// if err := c.forwardController.DeleteForwardRule(id); err != nil {
	// 	return fmt.Errorf("ошибка при удалении правила: %w", err)
	// }

	// fmt.Println("Правило пересылки успешно удалено")
	return nil
}

// handleReport обрабатывает команду report
func (c *Transport) handleReport(args []string) error {
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
		_, err = c.reportController.GenerateActivityReport(startDate, endDate)
	case "forwarding":
		_, err = c.reportController.GenerateForwardingReport(startDate, endDate)
	case "error":
		_, err = c.reportController.GenerateErrorReport(startDate, endDate)
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
	state, err := t.authController.GetAuthorizationState()
	if err != nil {
		return fmt.Errorf("ошибка при получении состояния авторизации: %w", err)
	}

	slog.Debug("GetAuthorizationState()", "state", state.AuthorizationStateType())

	switch state.AuthorizationStateType() {
	case client.TypeAuthorizationStateWaitPhoneNumber:

		// TODO: заменить на ввод номера телефона из конфигурации
		// if config.Telegram.PhoneNumber != "" {
		// 	fmt.Println("Используется номер телефона из конфигурации")
		// 	time.Sleep(2 * time.Second)
		// 	clientAuthorizer.PhoneNumber <- config.Telegram.PhoneNumber
		// 	maskedPhone := maskPhoneNumber(config.Telegram.PhoneNumber)
		// 	fmt.Println("Номер телефона:", maskedPhone)
		// } else {
		// 	fmt.Print("Введите номер телефона: ")
		// 	var phoneNumber string
		// 	fmt.Scanln(&phoneNumber)
		// 	clientAuthorizer.PhoneNumber <- phoneNumber
		// 	maskedPhone := maskPhoneNumber(phoneNumber)
		// 	fmt.Println("Используется номер:", maskedPhone)
		// }

		fmt.Println("Введите номер телефона: ")
		phoneNumber, err := t.hiddenReadLine()
		if err != nil {
			return fmt.Errorf("ошибка при чтении телефона: %w", err)
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

func (c *Transport) hiddenReadLine() (string, error) {
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	return string(password), err
}
