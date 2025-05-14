package auth

import (
	"log/slog"

	"github.com/zelenin/go-tdlib/client"
)

type telegramRepo interface {
	GetInitDone() <-chan any
	GetAuthState() client.AuthorizationState
	GetInputChan() chan string
}

// Service управляет процессом авторизации в Telegram
type Service struct {
	log *slog.Logger
	//
	telegramRepo telegramRepo
}

// New создает новый экземпляр сервиса авторизации
func New(telegramRepo telegramRepo) *Service {
	s := &Service{
		log: slog.With("module", "service.auth"),
		//
		telegramRepo: telegramRepo,
	}

	return s
}

// GetInitDone возвращает канал, который будет закрыт после инициализации клиента
func (s *Service) GetInitDone() <-chan any {
	return s.telegramRepo.GetInitDone()
}

// GetAuthState возвращает состояние авторизации
func (s *Service) GetAuthState() client.AuthorizationState {
	return s.telegramRepo.GetAuthState()
}

// GetInputChan возвращает канал для ввода данных
func (s *Service) GetInputChan() chan string {
	return s.telegramRepo.GetInputChan()
}
