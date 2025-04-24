package auth

import (
	"errors"
	"log/slog"

	"github.com/zelenin/go-tdlib/client"
)

type telegramRepo interface {
	GetClient() *client.Client
	InitClientDone() chan any
	CreateClient(
		createAuthorizer func(
			setClient func(*client.Client),
			shutdown func(),
		) client.AuthorizationStateHandler,
	)
}

// Service управляет процессом авторизации в Telegram
type Service struct {
	log *slog.Logger
	//
	telegramRepo telegramRepo
	authorizer   *Authorizer
}

// New создает новый экземпляр сервиса авторизации
func New(telegramRepo telegramRepo) *Service {
	s := &Service{
		log: slog.With("module", "service.auth"),
		//
		telegramRepo: telegramRepo,
	}

	сreateAuthorizer := func(setClient func(*client.Client), shutdown func()) client.AuthorizationStateHandler {
		s.authorizer = NewAuthorizer(setClient, shutdown)
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
func (s *Service) GetAuthorizationState() (client.AuthorizationState, error) {
	tdlibClient := s.telegramRepo.GetClient()
	if tdlibClient == nil {
		return nil, errors.New("клиент не инициализирован")
	}
	return tdlibClient.GetAuthorizationState()
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
