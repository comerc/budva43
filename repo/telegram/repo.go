package telegram

import (
	"context"
	"log/slog"
	"time"

	"github.com/zelenin/go-tdlib/client"

	"github.com/comerc/budva43/config"
)

// TODO: logout

// Repo предоставляет методы для взаимодействия с Telegram API через TDLib
type Repo struct {
	log *slog.Logger
	//
	client         *client.Client
	initClientDone chan any
	listenerChan   chan *client.Listener
	ctx            context.Context
	shutdown       func()
}

// New создает новый экземпляр репозитория Telegram
func New() *Repo {
	return &Repo{
		log: slog.With("module", "repo.telegram"),
		//
		client:         nil,
		initClientDone: make(chan any),
		listenerChan:   make(chan *client.Listener),
	}
}

// Start устанавливает соединение с Telegram API
func (r *Repo) Start(ctx context.Context, shutdown context.CancelFunc) error {
	// Инициализируем базовые настройки репозитория
	// Клиент будет создан позже, после установки авторизатора
	r.log.Info("Telegram repository initialization started")

	r.ctx = ctx // TODO: некрасиво!
	r.shutdown = shutdown

	return nil
}

// CreateClient создает клиент TDLib после установки авторизатора
func (r *Repo) CreateClient(
	createAuthorizer func(
		setClient func(*client.Client),
		shutdown func(),
	) client.AuthorizationStateHandler,
) {
	r.log.Info("Creating TDLib client")

	// Если неудачная авторизации, то клиент закрывается, потому перезапуск цикла
	for {
		authorizationStateHandler := createAuthorizer(r.setClient, r.shutdown)
		_, err := client.NewClient(authorizationStateHandler, setupOptions()...)
		if err != nil {
			r.log.Error("ошибка при создании клиента TDLib", "err", err)
			select {
			case <-r.ctx.Done():
				r.log.Info("ctx.Done()")
				return
			default:
				continue
			}
		}
		r.log.Info("TDLib client authorized")
		r.listenerChan <- r.client.GetListener()
		break
	}

	version := r.GetVersion()
	r.log.Info("TDLib", "version", version)

	me := r.GetMe()
	r.log.Info("Me",
		"FirstName", me.FirstName,
		"LastName", me.LastName,
		"Username", func() string {
			if me.Usernames != nil {
				return me.Usernames.EditableUsername
			}
			return ""
		}(),
	)
}

// GetVersion выводит информацию о версии TDLib
func (r *Repo) GetVersion() string {
	versionOption, err := r.client.GetOption(&client.GetOptionRequest{
		Name: "version",
	})
	if err != nil {
		r.log.Error("GetOption", "err", err)
		return ""
	}
	return versionOption.(*client.OptionValueString).Value
}

// GetMe выводит информацию о пользователе
func (r *Repo) GetMe() *client.User {
	me, err := r.client.GetMe()
	if err != nil {
		r.log.Error("GetMe", "err", err)
		return nil
	}
	return me
}

// GetListener возвращает канал, который вернёт Listener после авторизации клиента
func (r *Repo) GetListener() chan *client.Listener {
	return r.listenerChan
}

// GetClient возвращает клиент TDLib
func (r *Repo) GetClient() *client.Client {
	return r.client
}

// setClient устанавливает клиент TDLib
func (r *Repo) setClient(tdlibClient *client.Client) {
	// r.log.Info("setClient")
	if r.client != nil {
		return
	}
	r.client = tdlibClient
	close(r.initClientDone)
	// select {
	// case _, ok := <-r.initClientDone:
	// 	// r.log.Info("<-r.setClientDone", "ok", ok)
	// 	if !ok {
	// 		// r.log.Info("setClientDone closed")
	// 		return
	// 	}
	// default:
	// 	// r.log.Info("Closing setClientDone")
	// 	close(r.initClientDone)
	// }
}

// Close закрывает соединение с Telegram API
func (r *Repo) Close() error {
	if r.client == nil {
		return nil
	}
	_, err := r.client.Close()
	if err != nil {
		return err
	}
	r.client = nil
	// иногда при выходе наблюдаю ошибку в консоли (не зависит от service/engine):
	/*
	   [ 0][t 1][1745435133.056575059][Status.h:371][&ptr_ != nullptr && get_info().static_flag]       0x0 -3
	   signal: abort trap
	   make: *** [run] Error 1
	*/
	// возможно, нужно дождаться закрытия клиента
	time.Sleep(1 * time.Second)
	return nil
}

// setClientDone возвращает канал, который будет закрыт после инициализации клиента
func (r *Repo) InitClientDone() chan any {
	return r.initClientDone
}

// setupOptions устанавливает опции для клиента TDLib
func setupOptions() []client.Option {
	options := []client.Option{
		func(tdlibClient *client.Client) {
			tdlibClient.SetLogStream(&client.SetLogStreamRequest{
				LogStream: &client.LogStreamFile{
					Path:           config.Telegram.LogDirectory,
					MaxFileSize:    config.Telegram.LogMaxFileSize,
					RedirectStderr: true,
				},
			})
		},
		func(tdlibClient *client.Client) {
			tdlibClient.SetLogVerbosityLevel(&client.SetLogVerbosityLevelRequest{
				NewVerbosityLevel: config.Telegram.LogVerbosityLevel,
			})
		},
	}
	return options
}
