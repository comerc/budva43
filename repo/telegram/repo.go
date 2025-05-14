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
	client     *client.Client
	initDone   chan any
	clientDone chan any
	authState  client.AuthorizationState
	inputChan  chan string
}

// New создает новый экземпляр репозитория Telegram
func New() *Repo {
	return &Repo{
		log: slog.With("module", "repo.telegram"),
		//
		client:     nil,            // клиент будет создан позже, после успеха авторизатора
		initDone:   make(chan any), // закроется, когда инициализация завершена
		clientDone: make(chan any), // закроется, когда клиент авторизован
		authState:  nil,            // nil, потому что авторизатор ещё не запущен
		inputChan:  make(chan string, 1),
	}
}

// Start устанавливает соединение с Telegram API
func (r *Repo) Start(ctx context.Context) error {
	r.log.Info("Creating TDLib client")

	err := r.setupClientLog()
	if err != nil {
		r.log.Error("setupClientLog", "err", err)
		return err
	}

	tdlibParameters := &client.SetTdlibParametersRequest{
		UseTestDc:           config.Telegram.UseTestDc,
		DatabaseDirectory:   config.Telegram.DatabaseDirectory,
		FilesDirectory:      config.Telegram.FilesDirectory,
		UseFileDatabase:     config.Telegram.UseFileDatabase,
		UseChatInfoDatabase: config.Telegram.UseChatInfoDatabase,
		UseMessageDatabase:  config.Telegram.UseMessageDatabase,
		UseSecretChats:      config.Telegram.UseSecretChats,
		ApiId:               config.Telegram.ApiId,
		ApiHash:             config.Telegram.ApiHash,
		SystemLanguageCode:  config.Telegram.SystemLanguageCode,
		DeviceModel:         config.Telegram.DeviceModel,
		SystemVersion:       config.Telegram.SystemVersion,
		ApplicationVersion:  config.Telegram.ApplicationVersion,
	}
	authorizer := client.ClientAuthorizer(tdlibParameters)

	go func() {
		initDoneFlag := false
		for {
			select {
			case <-ctx.Done():
				r.log.Info("ctx.Done()")
				return
			case r.authState = <-authorizer.State:
				if !initDoneFlag {
					close(r.initDone)
					initDoneFlag = true
				}
				switch r.authState.(type) {
				case *client.AuthorizationStateWaitPhoneNumber:
					s := <-r.inputChan
					authorizer.PhoneNumber <- s
				case *client.AuthorizationStateWaitCode:
					s := <-r.inputChan
					authorizer.Code <- s
				case *client.AuthorizationStateWaitPassword:
					s := <-r.inputChan
					authorizer.Password <- s
				case *client.AuthorizationStateReady:
					break
				}
			}
		}
	}()

	go func() {
		for {
			tdlibClient, err := client.NewClient(authorizer)
			if err != nil {
				r.log.Error("client.NewClient", "err", err)
				select {
				case <-ctx.Done():
					r.log.Info("ctx.Done()")
					return
				default:
					continue
				}
			}
			r.client = tdlibClient
			close(r.clientDone)
			r.log.Info("TDLib client authorized")
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

		// r.listenerChan <- r.client.GetListener()
	}()

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

// GetClient возвращает клиент TDLib, если он авторизован
func (r *Repo) GetClient() *client.Client {
	<-r.clientDone
	return r.client
}

// GetInitDone возвращает канал, который будет закрыт после инициализации клиента
func (r *Repo) GetInitDone() <-chan any {
	return r.initDone
}

// GetClientDone возвращает канал, который будет закрыт после авторизации клиента
func (r *Repo) GetClientDone() <-chan any {
	return r.clientDone
}

// GetAuthState возвращает состояние авторизации
func (r *Repo) GetAuthState() client.AuthorizationState {
	return r.authState
}

// GetInputChan возвращает канал для ввода данных
func (r *Repo) GetInputChan() chan string {
	return r.inputChan
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
