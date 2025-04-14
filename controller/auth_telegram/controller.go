package auth_telegram

import (
	"github.com/zelenin/go-tdlib/client"
)

// authTelegramService определяет интерфейс сервиса авторизации Telegram
type authTelegramService interface {
	SubmitPhoneNumber(phone string)
	SubmitCode(code string)
	SubmitPassword(password string)
	GetAuthorizationState() (client.AuthorizationState, error)
	InitClientDone() chan any
}

// Controller представляет контроллер для авторизации в Telegram
type Controller struct {
	authTelegramService authTelegramService
}

// New создает новый экземпляр контроллера авторизации Telegram
func New(authTelegramService authTelegramService) *Controller {
	return &Controller{
		authTelegramService: authTelegramService,
	}
}

// SubmitPhoneNumber отправляет номер телефона для авторизации
func (c *Controller) SubmitPhoneNumber(phone string) {
	c.authTelegramService.SubmitPhoneNumber(phone)
}

// SubmitCode отправляет код подтверждения
func (c *Controller) SubmitCode(code string) {
	c.authTelegramService.SubmitCode(code)
}

// SubmitPassword отправляет пароль двухфакторной аутентификации
func (c *Controller) SubmitPassword(password string) {
	c.authTelegramService.SubmitPassword(password)
}

// GetAuthorizationState возвращает текущее состояние авторизации
func (c *Controller) GetAuthorizationState() (client.AuthorizationState, error) {
	return c.authTelegramService.GetAuthorizationState()
}

// InitClientDone возвращает канал, который будет закрыт после инициализации клиента
func (c *Controller) InitClientDone() chan any {
	return c.authTelegramService.InitClientDone()
}
