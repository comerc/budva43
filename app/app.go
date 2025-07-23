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

// TODO: вместо append использовать выделение памяти и назначение по индексу, когда размер массива известен
// TODO: copy-once просто отклоняет редактирование (как и при форвардинге), а хотелось бы отправлять копию со ссылкой на сообщение предыдущей редакции (только для копирования, т.к. при форвардинге невозможно добавить ссылку)
// TODO: вместо copy-once -> save-revision; и выполнять не только при копировании, но и при форвардинге (или удалять старое и вставлять новое, или не удалять старое и вставлять новое)
// TODO: сделать task init - для установки всех зависимостей
// TODO: pkg/tdlib-ubuntu - в какой папке лучше держать?
// TODO: при старте проверять новые необработанные сообщения в чатах
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

type RunFunc = func(
	ctx context.Context,
	cancel func(),
	gracefulShutdown func(closer io.Closer),
	wait func(),
) error

func NewApp() *App {
	return &App{
		log: log.NewLogger(),
	}
}

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
