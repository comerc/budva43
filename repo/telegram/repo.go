package telegram

import (
	"context"
	"path"
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
			defer r.log.DebugOrError("client.NewClient", &err)

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

	version := r.GetVersion()
	r.log.DebugOrError("TDLib", nil, "version", version)

	me := r.GetMe()
	r.log.DebugOrError("Me", nil,
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

// GetVersion выводит информацию о версии TDLib
func (r *Repo) GetVersion() string {
	var err error
	defer r.log.DebugOrError("GetVersion", &err)

	var versionOption client.OptionValue
	versionOption, err = r.GetClient().GetOption(&client.GetOptionRequest{
		Name: "version",
	})
	if err != nil {
		return ""
	}
	return versionOption.(*client.OptionValueString).Value
}

// GetMe выводит информацию о пользователе
func (r *Repo) GetMe() *client.User {
	var err error
	defer r.log.DebugOrError("GetMe", &err)

	var me *client.User
	me, err = r.GetClient().GetMe()
	if err != nil {
		return nil
	}
	return me
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
			Path:           path.Join(config.Telegram.LogDirectory, "telegram.log"),
			MaxFileSize:    config.Telegram.LogMaxFileSize,
			RedirectStderr: true,
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
