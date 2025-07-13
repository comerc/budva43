package engine

import (
	"context"

	"github.com/zelenin/go-tdlib/client"

	"github.com/comerc/budva43/app/log"
)

//go:generate mockery --name=telegramRepo --exported
type telegramRepo interface {
	GetClientDone() <-chan any
	// tdlibClient methods
	GetListener() *client.Listener
}

//go:generate mockery --name=loaderService --exported
type loaderService interface {
	Run()
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
	//
	telegramRepo                telegramRepo
	loaderService               loaderService
	updateNewMessageHandler     updateNewMessageHandler
	updateMessageEditedHandler  updateMessageEditedHandler
	updateDeleteMessagesHandler updateDeleteMessagesHandler
	updateMessageSendHandler    updateMessageSendHandler
}

// New создает новый экземпляр сервиса engine
func New(
	telegramRepo telegramRepo,
	loaderService loaderService,
	updateNewMessageHandler updateNewMessageHandler,
	updateMessageEditedHandler updateMessageEditedHandler,
	updateDeleteMessagesHandler updateDeleteMessagesHandler,
	updateMessageSendHandler updateMessageSendHandler,
) *Service {
	return &Service{
		log: log.NewLogger(),
		//
		telegramRepo:                telegramRepo,
		loaderService:               loaderService,
		updateNewMessageHandler:     updateNewMessageHandler,
		updateMessageEditedHandler:  updateMessageEditedHandler,
		updateDeleteMessagesHandler: updateDeleteMessagesHandler,
		updateMessageSendHandler:    updateMessageSendHandler,
	}
}

// StartContext запускает обработчик обновлений от Telegram
func (s *Service) StartContext(ctx context.Context) error {

	go s.run(ctx)

	return nil
}

// Close останавливает сервис
func (s *Service) Close() error {
	return nil
}

// run запускает обработчик обновлений от Telegram
func (s *Service) run(ctx context.Context) {
	// Ждём авторизации клиента и получаем канал обновлений от Telegram
	select {
	case <-ctx.Done():
		return
	case <-s.telegramRepo.GetClientDone():
		s.loaderService.Run()
		listener := s.telegramRepo.GetListener()
		defer listener.Close()
		s.handleUpdates(ctx, listener)
	}
}

// handleUpdates обрабатывает обновления от Telegram
func (s *Service) handleUpdates(ctx context.Context, listener *client.Listener) {
	for {
		select {
		case <-ctx.Done():
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
				s.updateNewMessageHandler.Run(ctx, updateByType)
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
