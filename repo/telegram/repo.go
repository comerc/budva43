package telegram

import (
	"context"
	"path/filepath"
	"time"

	"github.com/zelenin/go-tdlib/client"

	"github.com/comerc/budva43/app/config"
	"github.com/comerc/budva43/app/log"
)

// TODO: logout

// Repo предоставляет методы для взаимодействия с Telegram API через TDLib
type Repo struct {
	log *log.Logger
	//
	client     *client.Client
	clientDone chan any
}

// New создает новый экземпляр репозитория Telegram
func New() *Repo {
	r := &Repo{
		log: log.NewLogger("repo.telegram"),
		//
		client:     nil,            // клиент будет создан позже, после успеха авторизатора
		clientDone: make(chan any), // закроется, когда клиент авторизован
	}

	return r
}

// Start устанавливает соединение с Telegram API
func (r *Repo) Start(_ context.Context) error {
	var err error

	err = r.setupClientLog()
	if err != nil {
		return err
	}

	return nil
}

// CreateClient создает клиент TDLib после успешной авторизации
func (r *Repo) CreateClient(runAuthorizationStateHandler func() client.AuthorizationStateHandler) {
	for {
		ok := func() bool {
			var err error
			defer r.log.ErrorOrDebug(&err, "client.NewClient")

			authorizationStateHandler := runAuthorizationStateHandler()
			var tdlibClient *client.Client
			tdlibClient, err = client.NewClient(authorizationStateHandler)
			if err != nil {
				return false
			}

			r.client = tdlibClient
			close(r.clientDone)

			return true
		}()

		if ok {
			break
		}
	}
}

// Close закрывает клиент TDLib
func (r *Repo) Close() error {
	var err error

	if r.client == nil {
		return nil
	}
	_, err = r.client.Close()
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

// GetOption выводит информацию о параметрах TDLib
func (r *Repo) GetOption(req *client.GetOptionRequest) (client.OptionValue, error) {
	return r.GetClient().GetOption(req)
}

// GetMe выводит информацию о пользователе
func (r *Repo) GetMe() (*client.User, error) {
	return r.GetClient().GetMe()
}

// GetClient возвращает клиент TDLib, если он авторизован
func (r *Repo) GetClient() *client.Client {
	<-r.clientDone
	return r.client
}

// GetClientDone возвращает канал, который будет закрыт после авторизации клиента
func (r *Repo) GetClientDone() <-chan any {
	return r.clientDone
}

// setupClientLog устанавливает опции для клиента TDLib
func (r *Repo) setupClientLog() error {
	var err error
	_, err = client.SetLogStream(&client.SetLogStreamRequest{
		LogStream: &client.LogStreamFile{
			Path:           filepath.Join(config.Telegram.LogDirectory, "telegram.log"),
			MaxFileSize:    int64(config.Telegram.LogMaxFileSize) * 1024 * 1024, // MB
			RedirectStderr: false,
		},
	})
	if err != nil {
		return log.WrapError(err)
	}
	_, err = client.SetLogVerbosityLevel(&client.SetLogVerbosityLevelRequest{
		NewVerbosityLevel: config.Telegram.LogVerbosityLevel,
	})
	if err != nil {
		return log.WrapError(err)
	}
	return nil
}
