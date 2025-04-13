// service/auth_telegram/service.go
package auth_telegram

import (
	"log/slog"

	"github.com/zelenin/go-tdlib/client"
)

// telegramRepo определяет интерфейс репозитория Telegram
type telegramRepo interface {
	GetClient() *client.Client
	InitClientDone() chan any
	CreateClient(func(func(*client.Client)) client.AuthorizationStateHandler)
}

// Service управляет процессом авторизации в Telegram
type Service struct {
	telegramRepo telegramRepo
	authorizer   *Authorizer
}

// New создает новый экземпляр сервиса авторизации Telegram
func New(telegramRepo telegramRepo) *Service {
	s := &Service{
		telegramRepo: telegramRepo,
	}

	сreateAuthorizer := func(setClient func(*client.Client)) client.AuthorizationStateHandler {
		s.authorizer = NewAuthorizer(setClient)
		return s.authorizer
	}

	go telegramRepo.CreateClient(сreateAuthorizer)

	return s
}

// GetStateChan возвращает канал состояния авторизации
func (s *Service) GetStateChan() <-chan client.AuthorizationState {
	return s.authorizer.state
}

// GetAuthorizationState возвращает текущее состояние авторизации
func (s *Service) GetAuthorizationState() client.AuthorizationState {
	state, err := s.telegramRepo.GetClient().GetAuthorizationState()
	if err != nil {
		slog.Error("Ошибка при получении состояния авторизации", "error", err)
		return nil
	}
	return state
}

// SubmitPhoneNumber устанавливает номер телефона для авторизации
func (s *Service) SubmitPhoneNumber(phoneNumber string) {
	s.authorizer.phoneNumber <- phoneNumber
}

// SubmitCode устанавливает код подтверждения для авторизации
func (s *Service) SubmitCode(code string) {
	s.authorizer.code <- code
}

// SubmitPassword устанавливает пароль двухфакторной аутентификации
func (s *Service) SubmitPassword(password string) {
	s.authorizer.password <- password
}

// InitClientDone возвращает канал, который будет закрыт после инициализации клиента
func (s *Service) InitClientDone() chan any {
	return s.telegramRepo.InitClientDone()
}
