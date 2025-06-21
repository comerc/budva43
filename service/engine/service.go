package engine

import (
	"context"
	"errors"

	"github.com/zelenin/go-tdlib/client"

	"github.com/comerc/budva43/app/engine_config"
	"github.com/comerc/budva43/app/entity"
	"github.com/comerc/budva43/app/log"
)

//go:generate mockery --name=telegramRepo --exported
type telegramRepo interface {
	GetClientDone() <-chan any
	// tdlibClient methods
	LoadChats(*client.LoadChatsRequest) (*client.Ok, error)
	GetChatHistory(*client.GetChatHistoryRequest) (*client.Messages, error)
	GetListener() *client.Listener
}

//go:generate mockery --name=updateNewMessageHandler --exported
type updateNewMessageHandler interface {
	Run(ctx context.Context, update *client.UpdateNewMessage)
}

//go:generate mockery --name=updateMessageEditedHandler --exported
type updateMessageEditedHandler interface {
	Run(update *client.UpdateMessageEdited)
}

//go:generate mockery --name=updateDeleteMessagesHandler --exported
type updateDeleteMessagesHandler interface {
	Run(update *client.UpdateDeleteMessages)
}

//go:generate mockery --name=updateMessageSendHandler --exported
type updateMessageSendHandler interface {
	Run(update *client.UpdateMessageSendSucceeded)
}

// Service предоставляет функциональность движка пересылки сообщений
type Service struct {
	log *log.Logger
	ctx context.Context
	//
	telegramRepo                telegramRepo
	updateNewMessageHandler     updateNewMessageHandler
	updateMessageEditedHandler  updateMessageEditedHandler
	updateDeleteMessagesHandler updateDeleteMessagesHandler
	updateMessageSendHandler    updateMessageSendHandler
}

// New создает новый экземпляр сервиса engine
func New(
	telegramRepo telegramRepo,
	updateNewMessageHandler updateNewMessageHandler,
	updateMessageEditedHandler updateMessageEditedHandler,
	updateDeleteMessagesHandler updateDeleteMessagesHandler,
	updateMessageSendHandler updateMessageSendHandler,
) *Service {
	return &Service{
		log: log.NewLogger("service.engine"),
		//
		telegramRepo:                telegramRepo,
		updateNewMessageHandler:     updateNewMessageHandler,
		updateMessageEditedHandler:  updateMessageEditedHandler,
		updateDeleteMessagesHandler: updateDeleteMessagesHandler,
		updateMessageSendHandler:    updateMessageSendHandler,
	}
}

// Start запускает обработчик обновлений от Telegram
func (s *Service) Start(ctx context.Context) error {

	s.ctx = ctx

	go s.run()

	return nil
}

// Close останавливает сервис
func (s *Service) Close() error {
	return nil
}

// run запускает обработчик обновлений от Telegram
func (s *Service) run() {
	// Ждём авторизации клиента и получаем канал обновлений от Telegram
	select {
	case <-s.ctx.Done():
		return
	case <-s.telegramRepo.GetClientDone():
		s.loadConfig()
		listener := s.telegramRepo.GetListener()
		defer listener.Close()
		s.handleUpdates(listener)
	}
}

// loadConfig загружает конфигурацию
func (s *Service) loadConfig() {
	// Загружаем в первый раз engine.yml
	s.handleConfigReload()

	// Подключаем отслеживание изменений engine.yml
	engine_config.Watch(s.handleConfigReload)
}

// handleConfigReload обрабатывает изменения конфигурации
func (s *Service) handleConfigReload() {
	var err error
	defer s.log.ErrorOrDebug(&err, "handleConfigReload")

	err = engine_config.Reload(s.newFuncInitializeDestinations())

	var emptyConfigData *engine_config.ErrEmptyConfigData
	if errors.As(err, &emptyConfigData) {
		s.log.Warn(err.Error(), emptyConfigData.Args...)
		err = nil
	}
}

type initializeDestinations = func([]entity.ChatId)

// newFuncInitializeDestinations создает колбек для загрузки чатов
func (s *Service) newFuncInitializeDestinations() initializeDestinations {
	var fn initializeDestinations
	level := 0
	notFound := make(map[entity.ChatId]struct{})

	fn = func(destinations []entity.ChatId) {

		ok := func() bool {
			var err error
			defer s.log.ErrorOrDebug(&err, "initializeDestinations", "level", level)

			_, err = s.telegramRepo.LoadChats(&client.LoadChatsRequest{
				Limit: 200,
			})
			if err != nil {
				err = log.WrapError(err)
				return false
			}
			for _, dstChatId := range destinations {
				_, err := s.telegramRepo.GetChatHistory(&client.GetChatHistoryRequest{
					ChatId:    dstChatId,
					Limit:     1,
					OnlyLocal: true,
				})
				if err != nil {
					notFound[dstChatId] = struct{}{}
					continue
				}
				delete(notFound, dstChatId)
			}
			if len(notFound) == 0 {
				return false
			}
			if level == 2 {
				err = log.NewError("not found", "destinations", destinations)
				return false
			}
			level++
			return true
		}()
		if !ok {
			return
		}

		fn(destinations) // (!) хвостовая рекурсия
	}
	return fn
}

// handleUpdates обрабатывает обновления от Telegram
func (s *Service) handleUpdates(listener *client.Listener) {
	for {
		select {
		case <-s.ctx.Done():
			return
		case update, ok := <-listener.Updates:
			if !ok {
				return
			}

			if update.GetClass() != client.ClassUpdate {
				continue
			}

			switch updateByType := update.(type) {
			case *client.UpdateNewMessage:
				s.updateNewMessageHandler.Run(s.ctx, updateByType)
			case *client.UpdateMessageEdited:
				s.updateMessageEditedHandler.Run(updateByType)
			case *client.UpdateDeleteMessages:
				s.updateDeleteMessagesHandler.Run(updateByType)
			case *client.UpdateMessageSendSucceeded:
				s.updateMessageSendHandler.Run(updateByType)
			}
		}
	}
}
