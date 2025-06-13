package auth

import (
	"github.com/zelenin/go-tdlib/client"

	"github.com/comerc/budva43/app/log"
)

type authService interface {
	GetInitDone() <-chan any
	GetState() client.AuthorizationState
	GetInputChan() chan<- string
	GetVersion() string
	GetMe() *client.User
}

// Controller представляет контроллер для авторизации в Telegram
type Controller struct {
	log *log.Logger
	//
	authService authService
}

// New создает новый экземпляр контроллера авторизации Telegram
func New(authService authService) *Controller {
	return &Controller{
		log: log.NewLogger("controller.auth_telegram"),
		//
		authService: authService,
	}
}

// GetInitDone возвращает канал, который будет закрыт после инициализации клиента
func (c *Controller) GetInitDone() <-chan any {
	return c.authService.GetInitDone()
}

// GetAuthState возвращает состояние авторизации
func (c *Controller) GetState() client.AuthorizationState {
	return c.authService.GetState()
}

// GetInputChan возвращает канал для ввода данных
func (c *Controller) GetInputChan() chan<- string {
	return c.authService.GetInputChan()
}

// GetVersion возвращает версию TDLib
func (c *Controller) GetVersion() string {
	return c.authService.GetVersion()
}

// GetMe возвращает информацию о пользователе
func (c *Controller) GetMe() *client.User {
	return c.authService.GetMe()
}
