package auth_telegram

import (
	"github.com/zelenin/go-tdlib/client"
)

// authTelegramService определяет интерфейс сервиса авторизации Telegram
type authTelegramService interface {
	SetPhoneNumber(phone string)
	SetCode(code string)
	SetPassword(password string)
	GetStateChan() chan client.AuthorizationState
	GetAuthorizationState() client.AuthorizationState
	InitClientDone() chan any
}

// Controller представляет контроллер для авторизации в Telegram
type Controller struct {
	authService authTelegramService
}

// New создает новый экземпляр контроллера авторизации Telegram
func New(authService authTelegramService) *Controller {
	return &Controller{
		authService: authService,
	}
	// TODO: контроллер имеет право обращаться к репо - зачем проксировать через сервис?
}

// SubmitPhoneNumber отправляет номер телефона для авторизации
func (c *Controller) SubmitPhoneNumber(phone string) {
	c.authService.SetPhoneNumber(phone)
}

// SubmitCode отправляет код подтверждения
func (c *Controller) SubmitCode(code string) {
	c.authService.SetCode(code)
}

// SubmitPassword отправляет пароль двухфакторной аутентификации
func (c *Controller) SubmitPassword(password string) {
	c.authService.SetPassword(password)
}

// GetStateChan возвращает текущее состояние авторизации
func (c *Controller) GetStateChan() chan client.AuthorizationState {
	return c.authService.GetStateChan()
}

// GetAuthorizationState возвращает текущее состояние авторизации
func (c *Controller) GetAuthorizationState() client.AuthorizationState {
	return c.authService.GetAuthorizationState()
}

// InitClientDone возвращает канал, который будет закрыт после инициализации клиента
func (c *Controller) InitClientDone() chan any {
	return c.authService.InitClientDone()
}
