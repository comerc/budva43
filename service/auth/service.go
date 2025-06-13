package auth

import (
	"context"

	"github.com/zelenin/go-tdlib/client"

	"github.com/comerc/budva43/app/config"
	"github.com/comerc/budva43/app/log"
)

// telegramRepo представляет базовые методы репозитория Telegram, необходимые для авторизации
type telegramRepo interface {
	CreateClient(runAuthorizationStateHandler func() client.AuthorizationStateHandler)
	GetVersion() string
	GetMe() *client.User
}

// Service управляет процессом авторизации в Telegram
type Service struct {
	log *log.Logger
	//
	telegramRepo telegramRepo
	initFlag     bool
	initDone     chan any
	inputChan    chan string
	state        client.AuthorizationState
}

// New создает новый экземпляр сервиса авторизации
func New(telegramRepo telegramRepo) *Service {
	return &Service{
		log: log.NewLogger("service.auth"),
		//
		telegramRepo: telegramRepo,
		initDone:     make(chan any), // закроется, когда авторизатор запущен
		inputChan:    make(chan string, 1),
	}
}

// Start запускает процесс авторизации
func (s *Service) Start(ctx context.Context) error {

	go s.telegramRepo.CreateClient(s.runAuthorizationStateHandler(ctx))

	return nil
}

// Close останавливает сервис
func (s *Service) Close() error {

	close(s.inputChan)

	return nil
}

// GetInitDone возвращает канал, который будет закрыт после инициализации клиента
func (s *Service) GetInitDone() <-chan any {
	return s.initDone
}

// GetInputChan возвращает канал для ввода данных
func (s *Service) GetInputChan() chan<- string {
	return s.inputChan
}

// GetState возвращает текущее состояние авторизации
func (s *Service) GetState() client.AuthorizationState {
	return s.state
}

// GetVersion возвращает версию TDLib
func (s *Service) GetVersion() string {
	return s.telegramRepo.GetVersion()
}

// GetMe возвращает информацию о пользователе
func (s *Service) GetMe() *client.User {
	return s.telegramRepo.GetMe()
}

// runAuthorizationStateHandler обрабатывает состояния авторизации
func (s *Service) runAuthorizationStateHandler(ctx context.Context) func() client.AuthorizationStateHandler {
	return func() client.AuthorizationStateHandler {

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

		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case state, ok := <-authorizer.State:
					s.state = state
					s.log.Debug("authorizer.State", "state", s.state, "ok", ok)
					if !ok {
						return
					}
				}
				switch s.state.(type) {
				case *client.AuthorizationStateWaitTdlibParameters:
					if !s.initFlag {
						close(s.initDone)
						s.initFlag = true
					}
				case *client.AuthorizationStateWaitPhoneNumber:
					authorizer.PhoneNumber <- <-s.inputChan
				case *client.AuthorizationStateWaitCode:
					authorizer.Code <- <-s.inputChan
				case *client.AuthorizationStateWaitPassword:
					authorizer.Password <- <-s.inputChan
				}
			}
		}()

		return authorizer
	}
}
