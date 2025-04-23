package telegram

import (
	"context"
	"log/slog"

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
	authClientDone chan any
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
		authClientDone: make(chan any),
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

	options := []client.Option{
		client.WithLogVerbosity(&client.SetLogVerbosityLevelRequest{
			NewVerbosityLevel: config.Telegram.LogVerbosityLevel,
		}),
	}

	// Если неудачная авторизации, то клиент закрывается, потому перезапуск цикла
	for {
		authorizationStateHandler := createAuthorizer(r.setClient, r.shutdown)
		_, err := client.NewClient(authorizationStateHandler, options...)
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
		close(r.authClientDone)
		break
	}

	// Получаем информацию о версии TDLib
	versionOption, err := r.client.GetOption(&client.GetOptionRequest{
		Name: "version",
	})
	if err != nil {
		r.log.Error("GetOption error", "err", err)
		return
	}

	commitOption, err := r.client.GetOption(&client.GetOptionRequest{
		Name: "commit_hash",
	})
	if err != nil {
		r.log.Error("GetOption error", "err", err)
		return
	}

	r.log.Info("TDLib",
		"version", versionOption.(*client.OptionValueString).Value,
		"commit", commitOption.(*client.OptionValueString).Value,
	)

	// Получаем информацию о пользователе
	me, err := r.client.GetMe()
	if err != nil {
		r.log.Error("GetMe error", "err", err)
		return
	}

	r.log.Info("Me", "FirstName", me.FirstName) // , "LastName", me.LastName)
}

func (r *Repo) GetClient() *client.Client {
	return r.client
}

func (r *Repo) setClient(tdlibClient *client.Client) {
	// r.log.Info("setClient")
	if r.client != nil {
		return
	}
	r.client = tdlibClient
	close(r.initClientDone)
	// select {
	// case _, ok := <-r.setClientDone:
	// 	// r.log.Info("<-r.setClientDone", "ok", ok)
	// 	if !ok {
	// 		// r.log.Info("setClientDone closed")
	// 		return
	// 	}
	// default:
	// 	// r.log.Info("Closing setClientDone")
	// 	close(r.setClientDone)
	// }
}

// Stop закрывает соединение с Telegram API
func (r *Repo) Stop() error {
	if r.client == nil {
		return nil
	}
	_, err := r.client.Close()
	if err != nil {
		return err
	}
	r.client = nil
	return nil
}

// setClientDone возвращает канал, который будет закрыт после инициализации клиента
func (r *Repo) InitClientDone() chan any {
	return r.initClientDone
}

// NewClientDone возвращает канал, который будет закрыт после авторизации клиента
func (r *Repo) AuthClientDone() chan any {
	return r.authClientDone
}

// // GetMessage получает сообщение по идентификатору
// func (r *Repo) GetMessage(chatID, messageID int64) (*client.Message, error) {
// 	// Реализация будет добавлена позже
// 	return &client.Message{}, nil
// }

// // SendTextMessage отправляет текстовое сообщение
// func (r *Repo) SendMessage(chatID int64, text string) (*client.Message, error) {
// 	// Реализация будет добавлена позже
// 	return &client.Message{}, nil
// }

// // ForwardMessage пересылает сообщение
// func (r *Repo) ForwardMessage(fromChatID, messageID int64, toChatID int64) (*client.Message, error) {
// 	// Реализация будет добавлена позже
// 	return &client.Message{}, nil
// }

// // DeleteMessage удаляет сообщение
// func (r *Repo) DeleteMessage(chatID, messageID int64) error {
// 	// Реализация будет добавлена позже
// 	return nil
// }

// // EditMessage редактирует сообщение
// func (r *Repo) EditMessage(chatID, messageID int64, text string) (*client.Message, error) {
// 	// Реализация будет добавлена позже
// 	return &client.Message{}, nil
// }
