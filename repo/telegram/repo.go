package telegram

import (
	"context"
	"fmt"
	"path"
	"time"

	"github.com/zelenin/go-tdlib/client"

	"github.com/comerc/budva43/config"
	"github.com/comerc/budva43/util"
)

// TODO: logout

// Repo предоставляет методы для взаимодействия с Telegram API через TDLib
type Repo struct {
	log *util.Logger
	//
	client     *client.Client
	clientDone chan any
}

// New создает новый экземпляр репозитория Telegram
func New() *Repo {
	r := &Repo{
		log: util.NewLogger("repo.telegram"),
		//
		client:     nil,            // клиент будет создан позже, после успеха авторизатора
		clientDone: make(chan any), // закроется, когда клиент авторизован
	}

	return r
}

// Start устанавливает соединение с Telegram API
func (r *Repo) Start(_ context.Context) error {
	err := r.setupClientLog()
	if err != nil {
		// r.log.Error("setupClientLog", "err", err)
		return err
	}

	return nil
}

// CreateClient создает клиент TDLib после успешной авторизации
func (r *Repo) CreateClient(runAuthorizationStateHandler func() client.AuthorizationStateHandler) {
	for {
		// r.log.Info("Creating TDLib client")
		authorizationStateHandler := runAuthorizationStateHandler()
		tdlibClient, err := client.NewClient(authorizationStateHandler)
		if err != nil {
			// r.log.Error("client.NewClient", "err", err)
			continue
		}
		r.client = tdlibClient
		close(r.clientDone)
		// r.log.Info("TDLib client authorized")
		break
	}

	version := r.GetVersion()
	_ = version // TODO: костыль
	// r.log.Info("TDLib", "version", version)

	me := r.GetMe()
	_ = me // TODO: костыль
	// r.log.Info("Me",
	// 	"FirstName", me.FirstName,
	// 	// "LastName", me.LastName,
	// 	// "Username", func() string {
	// 	// 	if me.Usernames != nil {
	// 	// 		return me.Usernames.EditableUsername
	// 	// 	}
	// 	// 	return ""
	// 	// }(),
	// )
}

// Close закрывает клиент TDLib
func (r *Repo) Close() error {
	if r.client == nil {
		return nil
	}
	_, err := r.client.Close()
	if err != nil {
		return fmt.Errorf("ошибка закрытия TDLib клиента: %w", err)
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
	versionOption, err := r.GetClient().GetOption(&client.GetOptionRequest{
		Name: "version",
	})
	if err != nil {
		// r.log.Error("GetOption", "err", err)
		return ""
	}
	return versionOption.(*client.OptionValueString).Value
}

// GetMe выводит информацию о пользователе
func (r *Repo) GetMe() *client.User {
	me, err := r.GetClient().GetMe()
	if err != nil {
		// r.log.Error("GetMe", "err", err)
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
		return fmt.Errorf("ошибка настройки потока логов TDLib: %w", err)
	}
	_, err = client.SetLogVerbosityLevel(&client.SetLogVerbosityLevelRequest{
		NewVerbosityLevel: config.Telegram.LogVerbosityLevel,
	})
	if err != nil {
		return fmt.Errorf("ошибка настройки уровня логирования TDLib: %w", err)
	}
	return nil
}
