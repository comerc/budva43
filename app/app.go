package app

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/comerc/budva43/app/log"
	"github.com/comerc/budva43/app/util"
)

// TODO: переделать входные параметры функций на объект *Request (как у go-tdlib)
// TODO: pkg/tdlib-buntu - в какой папке лучше держать?
// TODO: при старте проверять новые необработанныесообщения в чатах
// TODO: реализовать InlineKeyboardButton (см. README.md -> examples )
// TODO: реализовать storage.BackupEnabled
// TODO: проверить на Race Condition
// TODO: заменить примитивы синхронизации на [CSP](../go-secrets/README_V2/Communicating Sequential Processes (CSP) и потокобезопасный счетчик.md)
// TODO: проверить весь перенесённый код на early return

// TODO: сделать образ tdlib для ubuntu в докере подобно ghcr.io/zelenin/tdlib-docker
// TODO: прикрутить готовый образ tdlib в докере для make build

type App struct {
	log *log.Logger
}

func NewApp() *App {
	return &App{
		log: log.NewLogger(),
	}
}

type RunFunc = func(
	ctx context.Context,
	cancel func(),
	gracefulShutdown func(closer io.Closer),
	waitFunc func(),
) error

// Run запускает основные компоненты приложения
func (a *App) Run(runFunc RunFunc) error { //nolint:error_log_or_return
	releaseVersion := util.GetReleaseVersion()
	fmt.Println("Release version:", releaseVersion)

	var err error
	// Исключение: логируем ошибку на этом уровне, но передаём выше
	// т.к. os.Exit(1) прерывает выполнение программы без обработки defer
	defer a.log.ErrorOrDebug(&err, "Приложение завершило работу")

	a.log.ErrorOrDebug(nil, "Запуск приложения")

	// Создаем контекст, который будет отменен при сигнале остановки
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Настраиваем обработку сигналов остановки
	a.setupSignalHandler(cancel)

	// Запускаем компоненты приложения
	err = runFunc(ctx, cancel, a.gracefulShutdown, func() {
		// Ожидаем завершения контекста
		<-ctx.Done()
		a.log.ErrorOrDebug(nil, "Начинаем graceful shutdown")
	})
	if err != nil {
		return err
	}

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
