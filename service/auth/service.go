package auth

import (
	"context"
	"log/slog"

	"github.com/comerc/budva43/config"
	"github.com/zelenin/go-tdlib/client"
)

// telegramRepo представляет базовые методы репозитория Telegram, необходимые для авторизации
type telegramRepo interface {
	CreateClient(ctx context.Context, authorizationStateHandler client.AuthorizationStateHandler)
}

// Service управляет процессом авторизации в Telegram
type Service struct {
	log *slog.Logger
	//
	telegramRepo     telegramRepo
	clientAuthorizer *clientAuthorizer
	initDone         chan any
	inputChan        chan string
	state            client.AuthorizationState
}

// New создает новый экземпляр сервиса авторизации
func New(telegramRepo telegramRepo) *Service {
	tdlibParameters := &client.SetTdlibParametersRequest{
		UseTestDc:           config.Telegram.UseTestDc,
		DatabaseDirectory:   config.Telegram.DatabaseDirectory,
		FilesDirectory:      config.Telegram.FilesDirectory,
		UseFileDatabase:     config.Telegram.UseFileDatabase,
		UseChatInfoDatabase: config.Telegram.UseChatInfoDatabase,
		UseMessageDatabase:  config.Telegram.UseMessageDatabase,
		UseSecretChats:      config.Telegram.UseSecretChats,
		ApiId:               config.Telegram.ApiId,
		ApiHash:             config.Telegram.ApiHash,
		SystemLanguageCode:  config.Telegram.SystemLanguageCode,
		DeviceModel:         config.Telegram.DeviceModel,
		SystemVersion:       config.Telegram.SystemVersion,
		ApplicationVersion:  config.Telegram.ApplicationVersion,
	}
	authorizer := client.ClientAuthorizer(tdlibParameters)
	clientAuthorizer := &clientAuthorizer{
		AuthorizationStateHandler: authorizer,
		State:                     authorizer.State,
		PhoneNumber:               authorizer.PhoneNumber,
		Code:                      authorizer.Code,
		Password:                  authorizer.Password,
	}

	return &Service{
		log: slog.With("module", "service.auth"),
		//
		telegramRepo:     telegramRepo,
		initDone:         make(chan any), // закроется, когда авторизатор запущен
		inputChan:        make(chan string, 1),
		clientAuthorizer: clientAuthorizer,
	}
}

// Start запускает процесс авторизации
func (s *Service) Start(ctx context.Context) error {
	s.log.Info("Запуск процесса авторизации")

	go s.handleAuthorizationStates(ctx)

	go s.telegramRepo.CreateClient(ctx, s.clientAuthorizer)

	return nil
}

// Close останавливает сервис
func (s *Service) Close() error {
	return nil
}

// handleAuthStates обрабатывает состояния авторизации
func (s *Service) handleAuthorizationStates(ctx context.Context) {
	initFlag := false
	for {
		select {
		case <-ctx.Done():
			s.log.Info("Завершение обработки состояний авторизации")
			return
		case s.state = <-s.clientAuthorizer.State:
			if !initFlag {
				close(s.initDone)
				initFlag = true
			}
			switch s.state.(type) {
			// case *client.AuthorizationStateWaitTdlibParameters:
			// TODO: ошибка установки параметров TDLib приведёт к зацикливанию?
			// невозможно вычленить ошибку - "500 Request aborted"
			case *client.AuthorizationStateWaitPhoneNumber:
				s.log.Info("Требуется номер телефона")
				s.clientAuthorizer.PhoneNumber <- <-s.inputChan
			case *client.AuthorizationStateWaitCode:
				s.log.Info("Требуется код подтверждения")
				s.clientAuthorizer.Code <- <-s.inputChan
			case *client.AuthorizationStateWaitPassword:
				s.log.Info("Требуется пароль")
				s.clientAuthorizer.Password <- <-s.inputChan
			case *client.AuthorizationStateReady:
				s.log.Info("Авторизация успешно завершена")
				break
			}
		}
	}
}

// GetInitDone возвращает канал, который будет закрыт после инициализации клиента
func (s *Service) GetInitDone() chan any {
	return s.initDone
}

// GetInputChan возвращает канал для ввода данных
func (s *Service) GetInputChan() chan string {
	return s.inputChan
}

// GetState возвращает текущее состояние авторизации
func (s *Service) GetState() client.AuthorizationState {
	return s.state
}

// clientAuthorizer адаптирует client.clientAuthorizer
type clientAuthorizer struct {
	client.AuthorizationStateHandler // адаптер для client.AuthorizationStateHandler
	State                            chan client.AuthorizationState
	PhoneNumber                      chan string
	Code                             chan string
	Password                         chan string
}
