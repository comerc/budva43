package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	config "github.com/comerc/budva43/config"
	authTelegramController "github.com/comerc/budva43/controller/auth_telegram"
	forwardController "github.com/comerc/budva43/controller/forward"
	messageController "github.com/comerc/budva43/controller/message"
	reportController "github.com/comerc/budva43/controller/report"
	badgerRepo "github.com/comerc/budva43/repo/badger"
	telegramRepo "github.com/comerc/budva43/repo/telegram"
	authTelegramService "github.com/comerc/budva43/service/auth_telegram"
	forwardRuleService "github.com/comerc/budva43/service/forward_rule"
	messsageService "github.com/comerc/budva43/service/message"
	reportService "github.com/comerc/budva43/service/report"

	// botTransport "github.com/comerc/budva43/transport/bot"
	cliTransport "github.com/comerc/budva43/transport/cli"
	webTransport "github.com/comerc/budva43/transport/web"
)

// TODO: сделать образ tdlib для ubuntu в докере подобно ghcr.io/zelenin/tdlib-docker
// TODO: прикрутить готовый образ tdlib в докере для make build

// Основная функция приложения
func main() {
	// Настройка логгера
	logHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level:     config.LogOptions.Level,
		AddSource: config.LogOptions.AddSource,
	})
	logger := slog.New(logHandler)
	slog.SetDefault(logger)

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

// shutdownCallback представляет функцию остановки компонента
type shutdownCallback func() error

// gracefulShutdown выполняет корректное завершение компонента и добавляет ошибки в набор
func gracefulShutdown(componentName string, errSet *errSet, callback shutdownCallback) {
	slog.Info("Останавливаем компонент", "компонент", componentName)
	if err := callback(); err != nil {
		slog.Error(
			"Ошибка при остановке компонента",
			"компонент", componentName,
			"ошибка", err,
		)
		errSet.add(fmt.Errorf("ошибка при остановке %s: %w", componentName, err))
	}
}

// setupSignalHandler настраивает обработку сигналов остановки
func setupSignalHandler(shutdown func()) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

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

	badgerRepo := badgerRepo.New()
	if err := badgerRepo.Start(ctx, cancel); err != nil {
		return fmt.Errorf("ошибка запуска badgerRepo: %w", err)
	}
	defer gracefulShutdown("badgerRepo", errSet, badgerRepo.Stop)
	slog.Info("badgerRepo запущен")

	// Создаем и запускаем репозиторий Telegram
	telegramRepo := telegramRepo.New()
	if err := telegramRepo.Start(ctx, cancel); err != nil {
		return fmt.Errorf("ошибка запуска telegramRepo: %w", err)
	}
	defer gracefulShutdown("telegramRepo", errSet, telegramRepo.Stop)
	slog.Info("telegramRepo запущен")

	// - Инициализация сервисов
	messageService := messsageService.New()
	forwardRuleService := forwardRuleService.New()
	reportService := reportService.New()
	authTelegramService := authTelegramService.New(telegramRepo)

	// - Инициализация контроллеров
	messageController := messageController.New(
		messageService,
	)
	forwardController := forwardController.New(
		forwardRuleService,
		messageService,
	)
	reportController := reportController.New(
		reportService,
	)

	// Инициализация контроллера авторизации Telegram
	authTelegramController := authTelegramController.New(authTelegramService)

	// - Инициализация транспортных адаптеров

	// botTransport := botTransport.New(
	// 	messageController,
	// 	forwardController,
	// 	reportController,
	// 	telegramRepo,
	// )
	// if err := botTransport.Start(ctx, cancel); err != nil {
	// 	return fmt.Errorf("ошибка запуска botTransport: %w", err)
	// }
	// defer gracefulShutdown("botTransport", errSet, botTransport.Stop)
	// slog.Info("botTransport запущен")

	cliTransport := cliTransport.New(
		messageController,
		forwardController,
		reportController,
		authTelegramController,
	)
	if err := cliTransport.Start(ctx, cancel); err != nil {
		return fmt.Errorf("ошибка запуска cliTransport: %w", err)
	}
	defer gracefulShutdown("cliTransport", errSet, cliTransport.Stop)
	slog.Info("cliTransport запущен")

	webTransport := webTransport.New(
		messageController,
		forwardController,
		reportController,
		authTelegramController,
	)
	if err := webTransport.Start(ctx, cancel); err != nil {
		return fmt.Errorf("ошибка запуска webTransport: %w", err)
	}
	defer gracefulShutdown("webTransport", errSet, webTransport.Stop)
	slog.Info("webTransport запущен")

	// Ожидаем завершения контекста
	<-ctx.Done()
	slog.Info("Получен сигнал завершения, начинаем graceful shutdown")

	return nil
}
