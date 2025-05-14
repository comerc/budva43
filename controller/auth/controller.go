package auth

import (
	"log/slog"

	"github.com/zelenin/go-tdlib/client"
)

type authService interface {
	GetInitDone() <-chan any
	GetAuthState() client.AuthorizationState
	GetInputChan() chan string
}

// Controller представляет контроллер для авторизации в Telegram
type Controller struct {
	log *slog.Logger
	//
	authService authService
}

// New создает новый экземпляр контроллера авторизации Telegram
func New(authService authService) *Controller {
	return &Controller{
		log: slog.With("module", "controller.auth_telegram"),
		//
		authService: authService,
	}
}

// GetInitDone возвращает канал, который будет закрыт после инициализации клиента
func (c *Controller) GetInitDone() <-chan any {
	return c.authService.GetInitDone()
}

// GetAuthState возвращает состояние авторизации
func (c *Controller) GetAuthState() client.AuthorizationState {
	return c.authService.GetAuthState()
}

// GetInputChan возвращает канал для ввода данных
func (c *Controller) GetInputChan() chan string {
	return c.authService.GetInputChan()
}
