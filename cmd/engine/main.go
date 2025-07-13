package main

import (
	"context"
	"io"
	"os"

	"github.com/comerc/budva43/app"
	updateDeleteMessagesHandler "github.com/comerc/budva43/handler/update_delete_messages"
	updateMessageEditedHandler "github.com/comerc/budva43/handler/update_message_edited"
	updateMessageSendHandler "github.com/comerc/budva43/handler/update_message_send"
	updateNewMessageHandler "github.com/comerc/budva43/handler/update_new_message"
	queueRepo "github.com/comerc/budva43/repo/queue"
	storageRepo "github.com/comerc/budva43/repo/storage"
	telegramRepo "github.com/comerc/budva43/repo/telegram"
	termRepo "github.com/comerc/budva43/repo/term"
	authService "github.com/comerc/budva43/service/auth"
	engineService "github.com/comerc/budva43/service/engine"
	engineLoaderService "github.com/comerc/budva43/service/engine_loader"
	filtersModeService "github.com/comerc/budva43/service/filters_mode"
	forwardedToService "github.com/comerc/budva43/service/forwarded_to"
	forwarderService "github.com/comerc/budva43/service/forwarder"
	mediaAlbumService "github.com/comerc/budva43/service/media_album"
	messageService "github.com/comerc/budva43/service/message"
	rateLimiterService "github.com/comerc/budva43/service/rate_limiter"
	storageService "github.com/comerc/budva43/service/storage"
	transformService "github.com/comerc/budva43/service/transform"
	termTransport "github.com/comerc/budva43/transport/term"
)

// Engine - это сервис, который выполняет пересылку сообщений.
// Недопустима отправка новых сообщений в исходящие чаты.

func main() {
	if err := app.NewApp().Run(runEngine); err != nil {
		os.Exit(1)
	}
}

func runEngine(
	ctx context.Context,
	cancel func(),
	gracefulShutdown func(closer io.Closer),
	waitFunc func(),
) error {
	var err error

	// - Инициализация репозиториев
	storageRepo := storageRepo.New()
	err = storageRepo.StartContext(ctx)
	if err != nil {
		return err
	}
	defer gracefulShutdown(storageRepo)
	telegramRepo := telegramRepo.New()
	err = telegramRepo.Start()
	if err != nil {
		return (err)
	}
	defer gracefulShutdown(telegramRepo)
	queueRepo := queueRepo.New()
	err = queueRepo.StartContext(ctx)
	if err != nil {
		return err
	}
	defer gracefulShutdown(queueRepo)
	termRepo := termRepo.New()
	err = termRepo.Start()
	if err != nil {
		return err
	}
	defer gracefulShutdown(termRepo)

	// - Инициализация вспомогательных сервисов
	storageService := storageService.New(storageRepo)
	engineLoaderService := engineLoaderService.New(telegramRepo)
	messageService := messageService.New()
	mediaAlbumService := mediaAlbumService.New()
	transformService := transformService.New(
		telegramRepo,
		storageService,
		messageService,
	)
	rateLimiterService := rateLimiterService.New()
	filtersModeService := filtersModeService.New()
	forwardedToService := forwardedToService.New()
	forwarderService := forwarderService.New(
		telegramRepo,
		storageService,
		messageService,
		transformService,
		rateLimiterService,
	)
	authService := authService.New(telegramRepo)
	err = authService.StartContext(ctx)
	if err != nil {
		return err
	}
	defer gracefulShutdown(authService)

	// - Инициализация основного сервиса и его обработчиков
	updateNewMessageHandler := updateNewMessageHandler.New(
		telegramRepo,
		queueRepo,
		storageService,
		messageService,
		mediaAlbumService,
		filtersModeService,
		forwardedToService,
		forwarderService,
	)
	updateMessageEditedHandler := updateMessageEditedHandler.New(
		telegramRepo,
		queueRepo,
		storageService,
		messageService,
		transformService,
		filtersModeService,
		forwarderService,
	)
	updateDeleteMessagesHandler := updateDeleteMessagesHandler.New(
		telegramRepo,
		queueRepo,
		storageService,
	)
	updateMessageSendHandler := updateMessageSendHandler.New(
		queueRepo,
		storageService,
	)
	engineService := engineService.New(
		telegramRepo,
		engineLoaderService,
		updateNewMessageHandler,
		updateMessageEditedHandler,
		updateDeleteMessagesHandler,
		updateMessageSendHandler,
	)
	err = engineService.StartContext(ctx)
	if err != nil {
		return err
	}
	defer gracefulShutdown(engineService)

	// - Инициализация фасадов
	// facadeGQL := facadeGQL.New(
	// 	telegramRepo,
	// )
	// facadeGRPC := facadeGRPC.New(
	// 	telegramRepo,
	// 	messageService,
	// )

	// - Инициализация транспортных адаптеров
	termTransport := termTransport.New(
		termRepo,
		authService,
	)
	err = termTransport.StartContext(ctx, cancel)
	if err != nil {
		return err
	}
	defer gracefulShutdown(termTransport)
	// webTransport := webTransport.New(
	// 	authService,
	// 	facadeGQL,
	// )
	// err = webTransport.StartContext(ctx, cancel)
	// if err != nil {
	// 	return err
	// }
	// defer gracefulShutdown(webTransport)
	// grpcTransport := grpcTransport.New(
	// 	facadeGRPC,
	// )
	// err = grpcTransport.Start()
	// if err != nil {
	// 	return err
	// }
	// defer gracefulShutdown(grpcTransport)

	waitFunc()

	return nil
}
