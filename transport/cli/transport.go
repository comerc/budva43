package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/zelenin/go-tdlib/client"
	"golang.org/x/term"

	"github.com/comerc/budva43/config"
	"github.com/comerc/budva43/util"
)

// TODO: не нравится, что нужно вводить auth для каждого последующего шага

// type reportController interface {
// 	GenerateActivityReport(startDate, endDate time.Time) *entity.ActivityReport
// 	GenerateForwardingReport(startDate, endDate time.Time) *entity.ForwardingReport
// 	GenerateErrorReport(startDate, endDate time.Time) *entity.ErrorReport
// }

type authController interface {
	GetInitDone() <-chan any
	GetState() client.AuthorizationState
	GetInputChan() chan<- string
}

// Transport представляет интерфейс командной строки
type Transport struct {
	log *util.Logger
	//
	// reportController reportController
	authController authController
	scanner        *bufio.Scanner
	commands       []command
	commandMap     map[string]*command
	shutdown       func()
}

// command представляет команду CLI
type command struct {
	name        string
	description string
	handler     func(args []string)
}

// New создает новый экземпляр CLI
func New(
	// reportController reportController,
	authController authController,
) *Transport {
	cli := &Transport{
		log: util.NewLogger("transport.cli"),
		//
		// reportController: reportController,
		authController: authController,
		scanner:        bufio.NewScanner(os.Stdin),
		commands:       []command{},
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
		// {
		// 	name:        "report",
		// 	description: "Генерация отчетов: activity, forwarding, error",
		// 	handler:     t.handleReport,
		// },
		// TODO: перенести в запуск cli
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
		case <-t.authController.GetInitDone():
			// t.log.Info("TDLib клиент готов к выполнению авторизации")
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

				t.processCommand(input)
			}
		}
	}()

	return nil
}

func (t *Transport) Close() error {
	return nil
}

// processCommand обрабатывает введенную команду
func (t *Transport) processCommand(input string) {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		// t.log.Warn("processCommand", "input", input)
		return
	}

	cmd := parts[0]
	var args []string
	if len(parts) > 1 {
		args = parts[1:]
	}

	if command, ok := t.commandMap[cmd]; ok {
		command.handler(args)
		return
	}

	// t.log.Info("Unknown command", "command", cmd)
	fmt.Printf("Неизвестная команда: %s. Введите 'help' для просмотра доступных команд.\n", cmd)
}

// handleHelp обрабатывает команду help
func (t *Transport) handleHelp(args []string) {
	fmt.Println("Доступные команды:")
	for _, cmd := range t.commands {
		fmt.Printf("  %-15s - %s\n", cmd.name, cmd.description)
	}
}

// handleExit обрабатывает команду exit
func (t *Transport) handleExit(args []string) {
	fmt.Println("Выход из программы...")
	t.shutdown()
}

// // handleReport обрабатывает команду report
// func (t *Transport) handleReport(args []string) {
// 	if len(args) == 0 {
// 		fmt.Println("Использование: report [activity|forwarding|error] [days=7]")
// 		return nil
// 	}

// 	reportType := args[0]
// 	days := 7
// 	if len(args) > 1 {
// 		if _, err := fmt.Sscanf(args[1], "%d", &days); err != nil {
// 			fmt.Println("Используется период по умолчанию (7 дней)")
// 		}
// 	}

// 	// Получаем даты для отчета
// 	endDate := time.Now()
// 	startDate := endDate.AddDate(0, 0, -days)

// 	fmt.Printf("Генерация отчета '%s' за период %s - %s...\n",
// 		reportType, startDate.Format("02.01.2006"), endDate.Format("02.01.2006"))

// 	var err error
// 	switch reportType {
// 	case "activity":
// 		_, err = t.reportController.GenerateActivityReport(startDate, endDate)
// 	case "forwarding":
// 		_, err = t.reportController.GenerateForwardingReport(startDate, endDate)
// 	case "error":
// 		_, err = t.reportController.GenerateErrorReport(startDate, endDate)
// 	default:
// 		return fmt.Errorf("неизвестный тип отчета: %s. Доступные: activity, forwarding, error", reportType)
// 	}

// 	if err != nil {
// 		return fmt.Errorf("ошибка при генерации отчета: %w", err)
// 	}

// 	fmt.Printf("Отчет '%s' успешно сгенерирован\n", reportType)
// 	return nil
// }

// handleAuth обрабатывает команду auth
func (t *Transport) handleAuth(args []string) {
	var err error
	defer func() {
		if err != nil {
			// t.log.Error("handleAuth", "err", err)
		}
	}()

	state := t.authController.GetState()
	if state == nil {
		// t.log.Warn("GetState вернул nil")
		return
	}

	stateType := state.AuthorizationStateType()
	// t.log.Debug("GetState", "stateType", stateType)

	switch stateType {
	case client.TypeAuthorizationStateWaitPhoneNumber:
		var phoneNumber string
		if config.Telegram.PhoneNumber != "" {
			// TODO: перенести в authController?
			phoneNumber = config.Telegram.PhoneNumber
			fmt.Println("Используется номер телефона из конфигурации:", util.MaskPhoneNumber(phoneNumber))
			time.Sleep(3 * time.Second)
		} else {
			fmt.Println("Введите номер телефона: ")
			phoneNumber, err = t.hiddenReadLine()
			if err != nil {
				// return fmt.Errorf("ошибка при чтении телефона: %w", err)
				return
			}
		}
		t.authController.GetInputChan() <- phoneNumber

	case client.TypeAuthorizationStateWaitCode:
		fmt.Println("Введите код подтверждения: ")
		code, err := t.hiddenReadLine()
		if err != nil {
			// return fmt.Errorf("ошибка при чтении кода: %w", err)
			return
		}
		t.authController.GetInputChan() <- code

	case client.TypeAuthorizationStateWaitPassword:
		fmt.Println("Введите пароль: ")
		password, err := t.hiddenReadLine()
		if err != nil {
			// return fmt.Errorf("ошибка при чтении пароля: %w", err)
			return
		}
		t.authController.GetInputChan() <- password

	}
}

func (t *Transport) hiddenReadLine() (string, error) {
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	return string(password), err
}
