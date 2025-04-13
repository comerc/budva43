// service/auth_telegram/service.go
package auth_telegram

import (
	"log/slog"

	"github.com/zelenin/go-tdlib/client"
)

// telegramRepo определяет интерфейс репозитория Telegram
type telegramRepo interface {
	// CreateClient() error // TODO: где запустить?
	GetClient() *client.Client
	GetPhoneNumber() chan string
	GetCode() chan string
	GetStateChan() chan client.AuthorizationState
	GetPassword() chan string
	InitClientDone() chan any
}

// Service управляет процессом авторизации в Telegram
type Service struct {
	telegramRepo telegramRepo
}

// New создает новый экземпляр сервиса авторизации Telegram
func New(telegramRepo telegramRepo) *Service {
	s := &Service{
		telegramRepo: telegramRepo,
	}

	return s
}

func (s *Service) GetStateChan() chan client.AuthorizationState {
	return s.telegramRepo.GetStateChan()
}

func (s *Service) GetAuthorizationState() client.AuthorizationState {
	tdlibClient := s.telegramRepo.GetClient()
	if tdlibClient == nil {
		slog.Error("Клиент TDLib не инициализирован")
		return nil
	}
	state, err := tdlibClient.GetAuthorizationState()
	if err != nil {
		slog.Error("Ошибка при получении состояния авторизации", "error", err)
		return nil
	}
	return state
}

// SetPhoneNumber устанавливает номер телефона для авторизации
func (s *Service) SetPhoneNumber(phoneNumber string) {
	// slog.Info("Попытка установить номер телефона", "phoneNumber", phoneNumber)
	s.telegramRepo.GetPhoneNumber() <- phoneNumber
}

// SetCode устанавливает код подтверждения для авторизации
func (s *Service) SetCode(code string) {
	// slog.Info("Попытка установить код подтверждения", "code", code)
	s.telegramRepo.GetCode() <- code
}

// SetPassword устанавливает пароль двухфакторной аутентификации
func (s *Service) SetPassword(password string) {
	//slog.Info("Попытка установить пароль", "passwordLength", len(password))
	s.telegramRepo.GetPassword() <- password
}

// InitClientDone возвращает канал, который будет закрыт после инициализации клиента
func (s *Service) InitClientDone() chan any {
	return s.telegramRepo.InitClientDone()
}
