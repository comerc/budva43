package telegram

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/comerc/budva43/config"
	"github.com/zelenin/go-tdlib/client"
)

// TODO: logout

// Repo предоставляет методы для взаимодействия с Telegram API через TDLib
type Repo struct {
	client         *client.Client
	initClientDone chan any
	ctx            context.Context
	cancel         context.CancelFunc
}

// New создает новый экземпляр репозитория Telegram
func New() *Repo {
	return &Repo{
		initClientDone: make(chan any),
	}
}

// Start устанавливает соединение с Telegram API
func (r *Repo) Start(ctx context.Context, cancel context.CancelFunc) error {
	// Инициализируем базовые настройки репозитория
	// Клиент будет создан позже, после установки авторизатора
	slog.Info("Telegram repository initialization started")

	r.ctx = ctx // TODO: некрасиво!
	r.cancel = cancel

	return nil
}

// CreateClient создает клиент TDLib после установки авторизатора
func (r *Repo) CreateClient(
	createAuthorizer func(
		setClient func(client *client.Client),
		cancel context.CancelFunc,
	) client.AuthorizationStateHandler,
) {
	slog.Info("Creating TDLib client")

	options := []client.Option{
		client.WithLogVerbosity(&client.SetLogVerbosityLevelRequest{
			NewVerbosityLevel: config.Telegram.LogVerbosityLevel,
		}),
	}

	// Если неудачная авторизации, то клиент закрывается, потому перезапуск цикла
	for {
		authorizationStateHandler := createAuthorizer(r.setClient, r.cancel)
		tdlibClient, err := client.NewClient(authorizationStateHandler, options...)
		if err != nil {
			slog.Error("ошибка при создании клиента TDLib", "error", err)
			select {
			case <-r.ctx.Done():
				slog.Info("ctx.Done()")
				return
			default:
				continue
			}
		}
		slog.Info("TDLib client created")
		r.setClient(tdlibClient)
		break
	}

	// Получаем информацию о версии TDLib
	versionOption, err := r.client.GetOption(&client.GetOptionRequest{
		Name: "version",
	})
	if err != nil {
		slog.Error("GetOption error", "error", err)
		return
	}

	commitOption, err := r.client.GetOption(&client.GetOptionRequest{
		Name: "commit_hash",
	})
	if err != nil {
		slog.Error("GetOption error", "error", err)
		return
	}

	slog.Info("TDLib",
		"version", versionOption.(*client.OptionValueString).Value,
		"commit", commitOption.(*client.OptionValueString).Value,
	)

	// Получаем информацию о пользователе
	me, err := r.client.GetMe()
	if err != nil {
		slog.Error("GetMe error", "error", err)
		return
	}

	slog.Info("Me", "FirstName", me.FirstName) //, "LastName", me.LastName)
}

func (r *Repo) GetClient() *client.Client {
	return r.client
}

func (r *Repo) setClient(client *client.Client) {
	slog.Info("setClient")
	r.client = client
	select {
	case _, ok := <-r.initClientDone:
		slog.Info("<-r.initClientDone", "ok", ok)
		if !ok {
			slog.Info("initClientDone closed")
			return
		}
	default:
		slog.Info("Closing initClientDone")
		close(r.initClientDone)
	}
}

// Stop закрывает соединение с Telegram API
func (r *Repo) Stop() error {
	if r.client == nil {
		return fmt.Errorf("клиент TDLib не инициализирован")
	}
	_, err := r.client.Close() // TODO: первый возвращаемый параметр client.Ok - зочем?
	if err != nil {
		return err
	}
	r.client = nil
	return nil
}

// InitClientDone возвращает канал, который будет закрыт после инициализации клиента
func (r *Repo) InitClientDone() chan any {
	return r.initClientDone
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
