package main

import (
	"context"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/comerc/budva43/app/log"
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
	filtersModeService "github.com/comerc/budva43/service/filters_mode"
	forwardedToService "github.com/comerc/budva43/service/forwarded_to"
	forwarderService "github.com/comerc/budva43/service/forwarder"
	mediaAlbumService "github.com/comerc/budva43/service/media_album"
	messageService "github.com/comerc/budva43/service/message"
	rateLimiterService "github.com/comerc/budva43/service/rate_limiter"
	storageService "github.com/comerc/budva43/service/storage"
	transformService "github.com/comerc/budva43/service/transform"
	cliTransport "github.com/comerc/budva43/transport/cli"
	webTransport "github.com/comerc/budva43/transport/web"
)

// TODO: pkg/tdlib-buntu - в какой папке лучше держать?
// TODO: при старте проверять новые необработанныесообщения в чатах
// TODO: реализовать InlineKeyboardButton (см. README.md -> examples )
// TODO: реализовать storage.BackupEnabled
// TODO: проверить на Race Condition
// TODO: заменить примитивы синхронизации на [CSP](../go-secrets/README_V2/Communicating Sequential Processes (CSP) и потокобезопасный счетчик.md)
// TODO: проверить весь перенесённый код на early return

// TODO: сделать образ tdlib для ubuntu в докере подобно ghcr.io/zelenin/tdlib-docker
// TODO: прикрутить готовый образ tdlib в докере для make build

// Основная функция приложения
func main() {
	// Запускаем приложение и обрабатываем ошибки
	if err := NewApp().Run(); err != nil {
		os.Exit(1)
	}
}

type App struct {
	log *log.Logger
}

func NewApp() *App {
	return &App{
		log: log.NewLogger(),
	}
}

// Run запускает основные компоненты приложения
func (a *App) Run() error {
	var err error
	// Исключение: логируем ошибку на этом уровне, но передаём выше
	// т.к. os.Exit(1) прерывает выполнение программы без обработки defer
	defer a.log.ErrorOrDebug(&err, "Приложение завершило работу")

	a.log.ErrorOrDebug(&err, "Запуск приложения")

	// Создаем контекст, который будет отменен при сигнале остановки
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Настраиваем обработку сигналов остановки
	a.setupSignalHandler(cancel)

	// - Инициализация репозиториев
	storageRepo := storageRepo.New()
	err = storageRepo.Start(ctx)
	if err != nil {
		return err
	}
	defer a.gracefulShutdown(storageRepo)

	telegramRepo := telegramRepo.New()
	err = telegramRepo.Start(ctx)
	if err != nil {
		return (err)
	}
	defer a.gracefulShutdown(telegramRepo)

	queueRepo := queueRepo.New()
	err = queueRepo.Start(ctx)
	if err != nil {
		return err
	}
	defer a.gracefulShutdown(queueRepo)

	termRepo := termRepo.New()
	err = termRepo.Start()
	if err != nil {
		return err
	}
	defer a.gracefulShutdown(termRepo)

	// - Инициализация вспомогательных сервисов
	storageService := storageService.New(storageRepo)
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
	err = authService.Start(ctx)
	if err != nil {
		return err
	}
	defer a.gracefulShutdown(authService)

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
		updateNewMessageHandler,
		updateMessageEditedHandler,
		updateDeleteMessagesHandler,
		updateMessageSendHandler,
	)
	err = engineService.Start(ctx)
	if err != nil {
		return err
	}
	defer a.gracefulShutdown(engineService)

	// - Инициализация транспортных адаптеров

	cliTransport := cliTransport.New(
		termRepo,
		authService,
	)
	err = cliTransport.Start(ctx, cancel)
	if err != nil {
		return err
	}
	defer a.gracefulShutdown(cliTransport)

	webTransport := webTransport.New(
		authService,
	)
	err = webTransport.Start(ctx, cancel)
	if err != nil {
		return err
	}
	defer a.gracefulShutdown(webTransport)

	// Ожидаем завершения контекста
	<-ctx.Done()
	a.log.ErrorOrDebug(nil, "Начинаем graceful shutdown")

	return nil
}

// setupSignalHandler настраивает обработку сигналов остановки
func (a *App) setupSignalHandler(shutdown func()) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		a.log.ErrorOrDebug(nil, "Получен сигнал остановки", "sig", sig)
		shutdown()
	}()
}

// gracefulShutdown выполняет корректное завершение компонента и добавляет ошибки в набор
func (a *App) gracefulShutdown(closer io.Closer) {
	if err := closer.Close(); err != nil {
		a.log.ErrorOrDebug(&err, "")
	}
}
