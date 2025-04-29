package auth

import (
	"log/slog"

	"github.com/zelenin/go-tdlib/client"
)

type authService interface {
	SubmitPhoneNumber(phone string)
	SubmitCode(code string)
	SubmitPassword(password string)
	GetAuthorizationState() (client.AuthorizationState, error)
	InitClientDone() chan any
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

// SubmitPhoneNumber отправляет номер телефона для авторизации
func (c *Controller) SubmitPhoneNumber(phone string) {
	c.authService.SubmitPhoneNumber(phone)
}

// SubmitCode отправляет код подтверждения
func (c *Controller) SubmitCode(code string) {
	c.authService.SubmitCode(code)
}

// SubmitPassword отправляет пароль двухфакторной аутентификации
func (c *Controller) SubmitPassword(password string) {
	c.authService.SubmitPassword(password)
}

// GetAuthorizationState возвращает текущее состояние авторизации
func (c *Controller) GetAuthorizationState() (client.AuthorizationState, error) {
	return c.authService.GetAuthorizationState()
}

// InitClientDone возвращает канал, который будет закрыт после инициализации клиента
func (c *Controller) InitClientDone() chan any {
	return c.authService.InitClientDone()
}
