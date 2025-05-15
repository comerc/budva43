package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	config "github.com/comerc/budva43/config"
	authController "github.com/comerc/budva43/controller/auth"
	queueRepo "github.com/comerc/budva43/repo/queue"
	storageRepo "github.com/comerc/budva43/repo/storage"
	telegramRepo "github.com/comerc/budva43/repo/telegram"
	authService "github.com/comerc/budva43/service/auth"
	engineService "github.com/comerc/budva43/service/engine"
	mediaAlbumService "github.com/comerc/budva43/service/media_album"
	messsageService "github.com/comerc/budva43/service/message"
	rateLimiterService "github.com/comerc/budva43/service/rate_limiter"
	storageService "github.com/comerc/budva43/service/storage"
	transformService "github.com/comerc/budva43/service/transform"
	cliTransport "github.com/comerc/budva43/transport/cli"
	webTransport "github.com/comerc/budva43/transport/web"
)

// TODO: сделать образ tdlib для ubuntu в докере подобно ghcr.io/zelenin/tdlib-docker
// TODO: прикрутить готовый образ tdlib в докере для make build

// Основная функция приложения
func main() {
	setupLogger()

	slog.Info("Запуск приложения Budva43")

	// go config.Watch(func(e fsnotify.Event) {
	// 	slog.Info("Config file changed", "file", e.Name)
	// }) // TODO: перезагрузка приложения при изменении конфигурации

	// Создаем контекст, который будет отменен при сигнале остановки
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Набор ошибок для graceful shutdown
	errSet := &errSet{}

	// Настраиваем обработку сигналов остановки
	setupSignalHandler(cancel)

	// Запускаем приложение и обрабатываем ошибки
	if err := runApp(ctx, errSet); err != nil {
		slog.Error("Ошибка при запуске приложения", "err", err)
		os.Exit(1)
	}

	// Выводим накопленные ошибки при завершении
	if errMsg := errSet.Error(); errMsg != "" {
		slog.Warn(errMsg)
	}

	slog.Info("Приложение успешно завершило работу")
}

// errSet представляет собой коллекцию ошибок, которые могут возникнуть при shutdown
type errSet struct {
	errors []error
}

// add добавляет ошибку в набор ошибок, если она не nil
func (e *errSet) add(err error) {
	if err != nil {
		e.errors = append(e.errors, err)
	}
}

// Error возвращает строковое представление всех ошибок
func (e *errSet) Error() string {
	if len(e.errors) == 0 {
		return ""
	}

	result := "Ошибки при завершении работы:\n"
	for i, err := range e.errors {
		result += fmt.Sprintf("  %d. %v\n", i+1, err)
	}
	return result
}

// gracefulShutdown выполняет корректное завершение компонента и добавляет ошибки в набор
func gracefulShutdown(componentName string, errSet *errSet, closer io.Closer) {
	slog.Info("Останавливаем", "componentName", componentName)
	if err := closer.Close(); err != nil {
		slog.Error(
			"Ошибка при остановке",
			"componentName", componentName,
			"err", err,
		)
		errSet.add(fmt.Errorf("ошибка при остановке %s: %w", componentName, err))
	}
}

// setupLogger настраивает логгер
func setupLogger() {
	logHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level:     config.LogOptions.Level,
		AddSource: config.LogOptions.AddSource,
	})
	logger := slog.New(logHandler)
	slog.SetDefault(logger)
}

// setupSignalHandler настраивает обработку сигналов остановки
func setupSignalHandler(shutdown func()) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		slog.Info("Получен сигнал остановки", "сигнал", sig)
		shutdown()
	}()
}

// runApp запускает основные компоненты приложения
func runApp(ctx context.Context, errSet *errSet) error {
	// - Инициализация репозиториев

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	storageRepo := storageRepo.New()
	if err := storageRepo.Start(ctx); err != nil {
		return fmt.Errorf("ошибка запуска storageRepo: %w", err)
	}
	defer gracefulShutdown("storageRepo", errSet, storageRepo)
	slog.Info("storageRepo запущен")

	telegramRepo := telegramRepo.New()
	if err := telegramRepo.Start(ctx); err != nil {
		return fmt.Errorf("ошибка запуска telegramRepo: %w", err)
	}
	defer gracefulShutdown("telegramRepo", errSet, telegramRepo)
	slog.Info("telegramRepo запущен")

	queueRepo := queueRepo.New()
	if err := queueRepo.Start(ctx); err != nil {
		return fmt.Errorf("ошибка запуска queueRepo: %w", err)
	}
	defer gracefulShutdown("queueRepo", errSet, queueRepo)
	slog.Info("queueRepo запущен")

	// - Инициализация сервисов
	messageService := messsageService.New()
	// reportService := reportService.New()
	transformService := transformService.New()
	storageService := storageService.New(storageRepo)
	mediaAlbumService := mediaAlbumService.New()
	rateLimiterService := rateLimiterService.New()
	engineService := engineService.New(
		telegramRepo,
		queueRepo,
		storageService,
		messageService,
		transformService,
		mediaAlbumService,
		rateLimiterService,
	)
	if err := engineService.Start(ctx); err != nil {
		return fmt.Errorf("ошибка запуска engineService: %w", err)
	}
	defer gracefulShutdown("engineService", errSet, engineService)
	slog.Info("engineService запущен")
	authService := authService.New(telegramRepo)
	if err := authService.Start(ctx); err != nil {
		return fmt.Errorf("ошибка запуска authService: %w", err)
	}
	defer gracefulShutdown("authService", errSet, authService)
	slog.Info("authService запущен")

	// - Инициализация контроллеров
	// reportController := reportController.New(
	// 	reportService,
	// )

	// Инициализация контроллера авторизации
	authController := authController.New(authService)

	// - Инициализация транспортных адаптеров

	// botTransport := botTransport.New(
	// 	reportController,
	//  authController,
	// )
	// if err := botTransport.Start(ctx, cancel); err != nil {
	// 	return fmt.Errorf("ошибка запуска botTransport: %w", err)
	// }
	// defer gracefulShutdown("botTransport", errSet, botTransport)
	// slog.Info("botTransport запущен")

	cliTransport := cliTransport.New(
		// reportController,
		authController,
	)
	if err := cliTransport.Start(ctx, cancel); err != nil {
		return fmt.Errorf("ошибка запуска cliTransport: %w", err)
	}
	defer gracefulShutdown("cliTransport", errSet, cliTransport)
	slog.Info("cliTransport запущен")

	webTransport := webTransport.New(
		// reportController,
		authController,
	)
	if err := webTransport.Start(ctx, cancel); err != nil {
		return fmt.Errorf("ошибка запуска webTransport: %w", err)
	}
	defer gracefulShutdown("webTransport", errSet, webTransport)
	slog.Info("webTransport запущен")

	// Ожидаем завершения контекста
	<-ctx.Done()
	slog.Info("Получен сигнал завершения, начинаем graceful shutdown")

	return nil
}
