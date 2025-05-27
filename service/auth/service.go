package auth

import (
	"context"
	"log/slog"
	"sync"

	"github.com/zelenin/go-tdlib/client"

	"github.com/comerc/budva43/config"
)

// telegramRepo представляет базовые методы репозитория Telegram, необходимые для авторизации
type telegramRepo interface {
	CreateClient(runAuthorizationStateHandler func() client.AuthorizationStateHandler)
}

// Service управляет процессом авторизации в Telegram
type Service struct {
	log *slog.Logger
	//
	telegramRepo telegramRepo
	initFlag     bool
	initDone     chan any
	inputChan    chan string
	state        client.AuthorizationState
	wg           sync.WaitGroup
}

// New создает новый экземпляр сервиса авторизации
func New(telegramRepo telegramRepo) *Service {
	return &Service{
		log: slog.With("module", "service.auth"),
		//
		telegramRepo: telegramRepo,
		initDone:     make(chan any), // закроется, когда авторизатор запущен
		inputChan:    make(chan string, 1),
	}
}

// Start запускает процесс авторизации
func (s *Service) Start(ctx context.Context) error {
	// s.log.Info("Запуск процесса авторизации")

	go s.telegramRepo.CreateClient(s.runAuthorizationStateHandler(ctx))

	return nil
}

// Close останавливает сервис
func (s *Service) Close() error {
	close(s.inputChan)
	s.wg.Wait()
	return nil
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

		s.wg.Wait()
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case state, ok := <-authorizer.State:
					s.state = state
					// s.log.Info("authorizer.State", "state", s.state, "ok", ok)
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
