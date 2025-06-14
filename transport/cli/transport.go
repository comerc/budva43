package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/zelenin/go-tdlib/client"
	"golang.org/x/term"

	"github.com/comerc/budva43/app/config"
	"github.com/comerc/budva43/app/log"
	"github.com/comerc/budva43/app/util"
)

// TODO: Добавить автодополнение команд

type notify = func(state client.AuthorizationState)

type authService interface {
	Subscribe(notify)
	GetInputChan() chan<- string
	GetClientDone() <-chan any
	GetStatus() string
}

// Transport представляет интерфейс командной строки
type Transport struct {
	log *log.Logger
	//
	authService   authService
	authStateChan chan client.AuthorizationState
	scanner       *bufio.Scanner
	commands      []command
	commandMap    map[string]*command
	shutdown      func()
}

// command представляет команду CLI
type command struct {
	name        string
	description string
	handler     func(args []string)
}

// New создает новый экземпляр CLI
func New(
	authService authService,
) *Transport {
	cli := &Transport{
		log: log.NewLogger("transport.cli"),
		//
		authService:   authService,
		authStateChan: make(chan client.AuthorizationState, 10),
		scanner:       bufio.NewScanner(os.Stdin),
		commands:      []command{},
	}

	// Регистрация команд
	cli.registerCommands()

	return cli
}

// Start запускает CLI интерфейс
func (t *Transport) Start(ctx context.Context, shutdown func()) error {
	t.shutdown = shutdown

	t.authService.Subscribe(t.createNotify())

	// Запускаем обработку ввода в отдельной горутине
	go func() {

		isAuth := false

		for {
			select {
			case <-ctx.Done():
				return
			case <-t.authService.GetClientDone():
				if !isAuth {
					fmt.Println(t.authService.GetStatus())
					isAuth = true
				}
				fmt.Println(">")
				if !t.scanner.Scan() {
					return
				}
				input := t.scanner.Text()
				t.processCommand(input)
			case state := <-t.authStateChan:
				t.processAuth(state)
			}
		}
	}()

	return nil
}

// Close закрывает транспорт
func (t *Transport) Close() error {
	close(t.authStateChan)
	return nil
}

// createNotify создает функцию для отправки состояния авторизации
func (t *Transport) createNotify() notify {
	return func(state client.AuthorizationState) {
		select {
		case t.authStateChan <- state:
			// успешно отправили
		default:
			// канал переполнен или закрыт - игнорируем
		}
	}
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
	}

	t.commandMap = make(map[string]*command)
	for _, cmd := range t.commands {
		t.commandMap[cmd.name] = &cmd
	}
}

// processCommand обрабатывает введенную команду
func (t *Transport) processCommand(input string) {
	var err error
	defer t.log.ErrorOrDebug(&err, "processCommand")

	parts := strings.Fields(input)
	if len(parts) == 0 {
		err = log.NewError("input is empty")
		return
	}

	cmd := parts[0]
	var args []string
	if len(parts) > 1 {
		args = parts[1:]
	}

	var command *command
	command, ok := t.commandMap[cmd]
	if !ok {
		err = log.NewError("unknown command", "cmd", cmd)
		fmt.Printf("Неизвестная команда: %s. Введите 'help' для просмотра доступных команд.\n", cmd)
		return
	}

	command.handler(args)
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

// processAuth обрабатывает состояние авторизации
func (t *Transport) processAuth(state client.AuthorizationState) {
	var err error
	defer t.log.ErrorOrDebug(&err, "processAuth")

	if state == nil {
		err = log.NewError("state is nil")
		return
	}

	stateType := state.AuthorizationStateType()

	switch stateType {
	case client.TypeAuthorizationStateWaitPhoneNumber:
		var phoneNumber string
		if config.Telegram.PhoneNumber != "" {
			phoneNumber = config.Telegram.PhoneNumber
			fmt.Println("Используется номер телефона из конфигурации:", util.MaskPhoneNumber(phoneNumber))
		} else {
			fmt.Println("Введите номер телефона: ")
			phoneNumber, err = t.hiddenReadLine()
			if err != nil {
				return
			}
		}
		t.authService.GetInputChan() <- phoneNumber

	case client.TypeAuthorizationStateWaitCode:
		fmt.Println("Введите код подтверждения: ")
		var code string
		code, err = t.hiddenReadLine()
		if err != nil {
			return
		}
		t.authService.GetInputChan() <- code

	case client.TypeAuthorizationStateWaitPassword:
		fmt.Println("Введите пароль: ")
		var password string
		password, err = t.hiddenReadLine()
		if err != nil {
			return
		}
		t.authService.GetInputChan() <- password

	}
}

// hiddenReadLine считывает консоль без отображения введенных символов
func (t *Transport) hiddenReadLine() (string, error) {
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	return string(password), log.WrapError(err)
}
