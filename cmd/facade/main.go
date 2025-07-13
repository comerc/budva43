package main

import (
	"context"
	"io"
	"os"

	"github.com/comerc/budva43/app"
	telegramRepo "github.com/comerc/budva43/repo/telegram"
	termRepo "github.com/comerc/budva43/repo/term"
	authService "github.com/comerc/budva43/service/auth"
	facadeService "github.com/comerc/budva43/service/facade"
	facadeGQL "github.com/comerc/budva43/service/facade_gql"
	facadeGRPC "github.com/comerc/budva43/service/facade_grpc"
	loaderService "github.com/comerc/budva43/service/loader"
	mediaAlbumService "github.com/comerc/budva43/service/media_album"
	messageService "github.com/comerc/budva43/service/message"
	grpcTransport "github.com/comerc/budva43/transport/grpc"
	termTransport "github.com/comerc/budva43/transport/term"
	webTransport "github.com/comerc/budva43/transport/web"
)

// Facade - это сервис, который предоставляет API: GraphQL, gRPC, REST.
// Допустима отправка новых сообщений в исходящие чаты.

func main() {
	if err := app.NewApp().Run(runFacade); err != nil {
		os.Exit(1)
	}
}

func runFacade(
	ctx context.Context,
	cancel func(),
	gracefulShutdown func(closer io.Closer),
	waitFunc func(),
) error {
	var err error

	// - Инициализация репозиториев
	// storageRepo := storageRepo.New()
	// err = storageRepo.StartContext(ctx)
	// if err != nil {
	// 	return err
	// }
	// defer gracefulShutdown(storageRepo)
	telegramRepo := telegramRepo.New()
	err = telegramRepo.Start()
	if err != nil {
		return (err)
	}
	defer gracefulShutdown(telegramRepo)
	// queueRepo := queueRepo.New()
	// err = queueRepo.StartContext(ctx)
	// if err != nil {
	// 	return err
	// }
	// defer gracefulShutdown(queueRepo)
	termRepo := termRepo.New()
	err = termRepo.Start()
	if err != nil {
		return err
	}
	defer gracefulShutdown(termRepo)

	// - Инициализация вспомогательных сервисов
	// storageService := storageService.New(storageRepo)
	loaderService := loaderService.New(telegramRepo)
	messageService := messageService.New()
	mediaAlbumService := mediaAlbumService.New()
	// transformService := transformService.New(
	// 	telegramRepo,
	// 	storageService,
	// 	messageService,
	// )
	// rateLimiterService := rateLimiterService.New()
	// filtersModeService := filtersModeService.New()
	// forwardedToService := forwardedToService.New()
	// forwarderService := forwarderService.New(
	// 	telegramRepo,
	// 	storageService,
	// 	messageService,
	// 	transformService,
	// 	rateLimiterService,
	// )

	// - Инициализация сервиса авторизации
	authService := authService.New(
		telegramRepo,
	)
	err = authService.StartContext(ctx)
	if err != nil {
		return err
	}
	defer gracefulShutdown(authService)

	// - Инициализация основного сервиса и его обработчиков
	// updateNewMessageHandler := updateNewMessageHandler.New(
	// 	telegramRepo,
	// 	queueRepo,
	// 	storageService,
	// 	messageService,
	// 	mediaAlbumService,
	// 	filtersModeService,
	// 	forwardedToService,
	// 	forwarderService,
	// )
	// updateMessageEditedHandler := updateMessageEditedHandler.New(
	// 	telegramRepo,
	// 	queueRepo,
	// 	storageService,
	// 	messageService,
	// 	transformService,
	// 	filtersModeService,
	// 	forwarderService,
	// )
	// updateDeleteMessagesHandler := updateDeleteMessagesHandler.New(
	// 	telegramRepo,
	// 	queueRepo,
	// 	storageService,
	// )
	// updateMessageSendHandler := updateMessageSendHandler.New(
	// 	queueRepo,
	// 	storageService,
	// )
	// engineService := engineService.New(
	// 	telegramRepo,
	// 	updateNewMessageHandler,
	// 	updateMessageEditedHandler,
	// 	updateDeleteMessagesHandler,
	// 	updateMessageSendHandler,
	// )
	// err = engineService.StartContext(ctx)
	// if err != nil {
	// 	return err
	// }
	// defer gracefulShutdown(engineService)

	// - Инициализация основного сервиса
	facadeService := facadeService.New(
		telegramRepo,
		loaderService,
	)
	err = facadeService.StartContext(ctx)
	if err != nil {
		return err
	}
	defer gracefulShutdown(facadeService)

	// - Инициализация фасадов
	facadeGQL := facadeGQL.New(
		telegramRepo,
	)
	facadeGRPC := facadeGRPC.New(
		telegramRepo,
		messageService,
		mediaAlbumService,
	)

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
	webTransport := webTransport.New(
		authService,
		facadeGQL,
	)
	err = webTransport.StartContext(ctx, cancel)
	if err != nil {
		return err
	}
	defer gracefulShutdown(webTransport)
	grpcTransport := grpcTransport.New(
		facadeGRPC,
	)
	err = grpcTransport.Start()
	if err != nil {
		return err
	}
	defer gracefulShutdown(grpcTransport)

	waitFunc()

	return nil
}
