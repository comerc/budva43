package auth

import (
	"context"
	"fmt"
	"sync"

	"github.com/zelenin/go-tdlib/client"

	"github.com/comerc/budva43/app/log"
)

type runAuthorizationStateHandler = func() client.AuthorizationStateHandler

//go:generate mockery --name=telegramRepo --exported
type telegramRepo interface {
	CreateTdlibParameters() *client.SetTdlibParametersRequest
	CreateClient(runAuthorizationStateHandler)
	GetClientDone() <-chan any
	// tdlibClient methods
	GetOption(*client.GetOptionRequest) (client.OptionValue, error)
	GetMe() (*client.User, error)
}

type notify = func(state client.AuthorizationState)

// Service управляет процессом авторизации в Telegram
type Service struct {
	log *log.Logger
	//
	telegramRepo telegramRepo
	inputChan    chan string
	// Для широковещательного оповещения
	subscribers []notify
	mu          sync.RWMutex
}

// New создает новый экземпляр сервиса авторизации
func New(telegramRepo telegramRepo) *Service {
	return &Service{
		log: log.NewLogger(),
		//
		telegramRepo: telegramRepo,
		inputChan:    make(chan string, 1),
		subscribers:  make([]notify, 0),
	}
}

// StartContext запускает процесс авторизации
func (s *Service) StartContext(ctx context.Context) error {

	go s.telegramRepo.CreateClient(newFuncRunAuthorizationStateHandler(ctx, s))

	return nil
}

// Close останавливает сервис
func (s *Service) Close() error {

	close(s.inputChan)

	return nil
}

// GetInputChan возвращает канал для ввода данных
func (s *Service) GetInputChan() chan<- string {
	return s.inputChan
}

// GetClientDone возвращает канал, который будет закрыт после завершения авторизации
func (s *Service) GetClientDone() <-chan any {
	return s.telegramRepo.GetClientDone()
}

// GetStatus возвращает статус авторизации
func (s *Service) GetStatus() string {
	var err error
	defer s.log.ErrorOrDebug(&err, "")

	var versionOption client.OptionValue
	versionOption, err = s.telegramRepo.GetOption(&client.GetOptionRequest{
		Name: "version",
	})
	if err != nil {
		return ""
	}
	version := versionOption.(*client.OptionValueString).Value

	var me *client.User
	me, err = s.telegramRepo.GetMe()
	if err != nil {
		return ""
	}

	return fmt.Sprintf("TDLib version: %s userId: %d", version, me.Id)
}

// Subscribe добавляет подписчика на изменения состояния авторизации
func (s *Service) Subscribe(notify notify) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.subscribers = append(s.subscribers, notify)
}

// broadcast отправляет состояние всем подписчикам
func (s *Service) broadcast(state client.AuthorizationState) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, notify := range s.subscribers {
		go notify(state)
	}
}

// newFuncRunAuthorizationStateHandler обрабатывает состояния авторизации
func newFuncRunAuthorizationStateHandler(ctx context.Context, s *Service) runAuthorizationStateHandler {
	return func() client.AuthorizationStateHandler {

		tdlibParameters := s.telegramRepo.CreateTdlibParameters()
		authorizer := client.ClientAuthorizer(tdlibParameters)

		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case state, ok := <-authorizer.State:
					if !ok {
						return
					}
					stateType := state.AuthorizationStateType()
					if stateType == client.TypeAuthorizationStateClosing {
						continue // пропускаем broadcast, но продолжаем <-authorizer.State
					}
					s.broadcast(state)
					switch stateType {
					case client.TypeAuthorizationStateWaitPhoneNumber:
						authorizer.PhoneNumber <- <-s.inputChan
					case client.TypeAuthorizationStateWaitCode:
						authorizer.Code <- <-s.inputChan
					case client.TypeAuthorizationStateWaitPassword:
						authorizer.Password <- <-s.inputChan
					}
				}
			}
		}()

		return authorizer
	}
}
