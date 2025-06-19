package engine

import (
	"context"

	"github.com/zelenin/go-tdlib/client"

	"github.com/comerc/budva43/app/engine_config"
	"github.com/comerc/budva43/app/log"
)

//go:generate mockery --name=telegramRepo --exported
type telegramRepo interface {
	GetClientDone() <-chan any
	// tdlibClient methods
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

	// Настраиваем отслеживание изменений engine.yml
	engine_config.Watch(s.handleConfigReload)

	go s.run()

	return nil
}

// Close останавливает сервис
func (s *Service) Close() error {
	return nil
}

// handleConfigReload обрабатывает изменения конфигурации
func (s *Service) handleConfigReload() {
	var err error
	defer s.log.ErrorOrDebug(&err, "handleConfigReload")

	err = engine_config.Reload()
}

// run запускает обработчик обновлений от Telegram
func (s *Service) run() {
	// Ждём авторизации клиента и получаем канал обновлений от Telegram
	select {
	case <-s.ctx.Done():
		return
	case <-s.telegramRepo.GetClientDone():
		listener := s.telegramRepo.GetListener()
		defer listener.Close()
		s.handleUpdates(listener)
	}
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
