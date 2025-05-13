package telegram

import (
	"context"
	"log/slog"
	"path"
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
		client:         nil, // клиент будет создан позже, после установки авторизатора
		initClientDone: make(chan any),
		listenerChan:   make(chan *client.Listener, 1),
	}
}

// Start устанавливает соединение с Telegram API
func (r *Repo) Start(ctx context.Context, shutdown context.CancelFunc) error {
	r.ctx = ctx // TODO: некрасиво!
	r.shutdown = shutdown

	return nil
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
	// TODO: возможно, нужно дождаться закрытия клиента
	time.Sleep(1 * time.Second)
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

	err := r.setupClientLog()
	if err != nil {
		r.log.Error("setupClientLog", "err", err)
		return
	}

	// Если неудачная авторизации, то клиент закрывается, потому перезапуск цикла
	for {
		authorizationStateHandler := createAuthorizer(r.setClient, r.shutdown)
		_, err := client.NewClient(authorizationStateHandler)
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
		// "LastName", me.LastName,
		// "Username", func() string {
		// 	if me.Usernames != nil {
		// 		return me.Usernames.EditableUsername
		// 	}
		// 	return ""
		// }(),
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

// setClientDone возвращает канал, который будет закрыт после инициализации клиента
func (r *Repo) InitClientDone() chan any {
	return r.initClientDone
}

// setupClientLog устанавливает опции для клиента TDLib
func (r *Repo) setupClientLog() error {
	var err error
	_, err = client.SetLogStream(&client.SetLogStreamRequest{
		LogStream: &client.LogStreamFile{
			Path:           path.Join(config.Telegram.LogDirectory, "telegram.log"),
			MaxFileSize:    config.Telegram.LogMaxFileSize,
			RedirectStderr: true,
		},
	})
	if err != nil {
		return err
	}
	_, err = client.SetLogVerbosityLevel(&client.SetLogVerbosityLevelRequest{
		NewVerbosityLevel: config.Telegram.LogVerbosityLevel,
	})
	if err != nil {
		return err
	}
	return nil
}
