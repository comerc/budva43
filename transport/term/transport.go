package term

import (
	"context"
	"fmt"
	"strings"

	"github.com/zelenin/go-tdlib/client"

	"github.com/comerc/budva43/app/config"
	"github.com/comerc/budva43/app/log"
	"github.com/comerc/budva43/app/util"
)

//go:generate mockery --name=termRepo --exported
type termRepo interface {
	HiddenReadLine() (string, error)
	ReadLine() (string, error)
}

type notify = func(state client.AuthorizationState)

//go:generate mockery --name=authService --exported
type authService interface {
	Subscribe(notify)
	GetInputChan() chan<- string
	GetClientDone() <-chan any
	GetStatus() string
}

// Transport представляет терминальный интерфейс
type Transport struct {
	log *log.Logger
	//
	termRepo      termRepo
	authService   authService
	authStateChan chan client.AuthorizationState
	commands      []command
	commandMap    map[string]*command
	shutdown      func()
	phoneNumber   string
}

// command представляет команду CLI
type command struct {
	name        string
	description string
	handler     func(args []string)
}

// New создает новый экземпляр терминального транспорта
func New(termRepo termRepo, authService authService) *Transport {
	term := &Transport{
		log: log.NewLogger(),
		//
		termRepo:      termRepo,
		authService:   authService,
		authStateChan: make(chan client.AuthorizationState, 10),
		commands:      []command{},
		phoneNumber:   config.Telegram.PhoneNumber,
	}

	// Регистрация команд
	term.registerCommands()

	return term
}

// WithPhoneNumber устанавливает номер телефона для авторизации
func (t *Transport) WithPhoneNumber(v string) *Transport {
	t.phoneNumber = v
	return t
}

// Start запускает терминальный интерфейс
func (t *Transport) Start(ctx context.Context, shutdown func()) error {
	t.shutdown = shutdown

	t.authService.Subscribe(newFuncNotify(t))

	go t.runInputLoop(ctx)

	return nil
}

// Close закрывает транспорт
func (t *Transport) Close() error {
	close(t.authStateChan)
	return nil
}

// runInputLoop запускает цикл обработки ввода
func (t *Transport) runInputLoop(ctx context.Context) {
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
			input, err := t.termRepo.ReadLine()
			if err != nil {
				return
			}
			t.processCommand(input)
		case state := <-t.authStateChan:
			t.processAuth(state)
		}
	}
}

// newFuncNotify создает функцию для отправки состояния авторизации
func newFuncNotify(t *Transport) notify {
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
	defer t.log.ErrorOrDebug(&err, "")

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
	defer t.log.ErrorOrDebug(&err, "")

	if state == nil {
		err = log.NewError("state is nil")
		return
	}

	stateType := state.AuthorizationStateType()

	switch stateType {
	case client.TypeAuthorizationStateWaitPhoneNumber:
		phoneNumber := t.phoneNumber
		if phoneNumber == "" {
			fmt.Println("Введите номер телефона: ")
			phoneNumber, err = t.termRepo.HiddenReadLine()
			if err != nil {
				return
			}
		} else {
			fmt.Println("Номер телефона:", util.MaskPhoneNumber(phoneNumber))
		}
		t.authService.GetInputChan() <- phoneNumber

	case client.TypeAuthorizationStateWaitCode:
		fmt.Println("Введите код подтверждения: ")
		var code string
		code, err = t.termRepo.HiddenReadLine()
		if err != nil {
			return
		}
		t.authService.GetInputChan() <- code

	case client.TypeAuthorizationStateWaitPassword:
		fmt.Println("Введите пароль: ")
		var password string
		password, err = t.termRepo.HiddenReadLine()
		if err != nil {
			return
		}
		t.authService.GetInputChan() <- password

	}
}
