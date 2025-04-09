package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	config "github.com/comerc/budva43/config"
	forwardController "github.com/comerc/budva43/controller/forward"
	messageController "github.com/comerc/budva43/controller/message"
	reportController "github.com/comerc/budva43/controller/report"
	badgerRepo "github.com/comerc/budva43/repository/badger"
	telegramRepo "github.com/comerc/budva43/repository/telegram"
	forwardRuleService "github.com/comerc/budva43/service/forward_rule"
	messsageService "github.com/comerc/budva43/service/message"
	reportService "github.com/comerc/budva43/service/report"
	httpTransport "github.com/comerc/budva43/transport/http"
	telegramTransport "github.com/comerc/budva43/transport/telegram"
)

var configPath = flag.String("config", ".", "config path")

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

// Основная функция приложения
func main() {
	// Настройка логгера
	logHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true, // Добавляем информацию об источнике лога
	})
	logger := slog.New(logHandler)
	slog.SetDefault(logger)

	slog.Info("Запуск приложения Budva43")

	// Создаем контекст, который будет отменен при сигнале остановки
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Набор ошибок для graceful shutdown
	errSet := &errSet{}

	// Настраиваем обработку сигналов остановки
	setupSignalHandler(cancel)

	// Запускаем приложение и обрабатываем ошибки
	if err := runApp(ctx, errSet); err != nil {
		slog.Error("Ошибка при запуске приложения", "ошибка", err)
		os.Exit(1)
	}

	// Выводим накопленные ошибки при завершении
	if errMsg := errSet.Error(); errMsg != "" {
		slog.Warn(errMsg)
	}

	slog.Info("Приложение успешно завершило работу")
}

// setupSignalHandler настраивает обработку сигналов остановки
func setupSignalHandler(cancel context.CancelFunc) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		slog.Info("Получен сигнал остановки", "сигнал", sig)
		cancel()
	}()
}

// runApp запускает основные компоненты приложения
func runApp(ctx context.Context, errSet *errSet) error {
	// 1. Загрузка и инициализация конфигурации
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("ошибка загрузки конфигурации: %w", err)
	}
	go config.Watch()
	slog.Info("Конфигурация загружена успешно")

	// 2. Инициализация репозиториев

	// Инициализируем BadgerDB репозиторий
	badgerRepo := badgerRepo.New(cfg.Storage.DatabasePath)
	if err := badgerRepo.Connect(ctx); err != nil {
		return fmt.Errorf("ошибка подключения к BadgerDB: %w", err)
	}
	defer gracefulShutdown("BadgerDB", errSet, badgerRepo.Close)
	slog.Info("Подключение к BadgerDB установлено")

	// Инициализируем Telegram репозиторий
	telegramRepo, err := telegramRepo.New(telegramRepo.AuthInfo{
		ApiID:               cfg.Telegram.ApiID,
		ApiHash:             cfg.Telegram.ApiHash,
		PhoneNumber:         cfg.Telegram.PhoneNumber,
		BotToken:            cfg.Telegram.BotToken,
		UseTestDC:           cfg.Telegram.UseTestDC,
		DatabaseDirectory:   cfg.Telegram.DatabaseDirectory,
		FilesDirectory:      cfg.Telegram.FilesDirectory,
		UseFileDatabase:     cfg.Telegram.UseFileDatabase,
		UseChatInfoDatabase: cfg.Telegram.UseChatInfoDatabase,
		UseMessageDatabase:  cfg.Telegram.UseMessageDatabase,
	})
	if err != nil {
		return fmt.Errorf("ошибка создания Telegram репозитория: %w", err)
	}

	if err := telegramRepo.Connect(ctx); err != nil {
		return fmt.Errorf("ошибка подключения к Telegram API: %w", err)
	}
	defer gracefulShutdown("Telegram API", errSet, telegramRepo.Close)
	slog.Info("Подключение к Telegram API установлено")

	// 3. Инициализация сервисов
	messageService := messsageService.New()
	forwardRuleService := forwardRuleService.New()
	reportService := reportService.New()

	// 4. Инициализация контроллеров
	messageController := messageController.New(
		messageService,
		telegramRepo,
	)

	// Создаем контроллеры с нужными зависимостями
	forwardController := forwardController.New(
		forwardRuleService,
		messageService,
		telegramRepo,
		badgerRepo,
	)

	reportController := reportController.New(
		reportService,
		badgerRepo,
	)

	// 5. Инициализация транспортных адаптеров

	// HTTP транспорт
	httpRouter := httpTransport.New(
		messageController,
		forwardController,
		reportController,
	)
	httpServer := httpTransport.NewServer(
		httpRouter,
		httpTransport.Config{
			Host:            cfg.Web.Host,
			Port:            cfg.Web.Port,
			ReadTimeout:     time.Duration(cfg.Web.ReadTimeout) * time.Second,
			WriteTimeout:    time.Duration(cfg.Web.WriteTimeout) * time.Second,
			ShutdownTimeout: time.Duration(cfg.Web.ShutdownTimeout) * time.Second,
		},
	)

	// Запускаем HTTP сервер
	if err := httpServer.Start(ctx); err != nil {
		return fmt.Errorf("ошибка запуска HTTP сервера: %w", err)
	}
	defer gracefulShutdown("HTTP Server", errSet, httpServer.Stop)
	slog.Info("HTTP сервер запущен")

	// Telegram транспорт
	// В данном случае telegramRepo может выполнять роль telegramClient, поскольку он содержит нужные методы
	telegramHandler := telegramTransport.New(
		messageController,
		forwardController,
		reportController,
		telegramRepo,
		cfg.Telegram.AdminChatID,
	)

	if err := telegramHandler.Start(ctx); err != nil {
		return fmt.Errorf("ошибка запуска Telegram обработчика: %w", err)
	}
	defer gracefulShutdown("Telegram Handler", errSet, telegramHandler.Stop)
	slog.Info("Telegram обработчик запущен")

	// CLI транспорт временно отключен до полной реализации
	// В реальном проекте здесь нужно создать и запустить CLI транспорт
	slog.Info("CLI транспорт временно отключен")

	// Ожидаем завершения контекста
	<-ctx.Done()
	slog.Info("Получен сигнал завершения, начинаем graceful shutdown")

	return nil
}
