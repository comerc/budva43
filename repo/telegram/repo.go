package telegram

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/comerc/budva43/config"
	"github.com/zelenin/go-tdlib/client"
)

// TODO: logout
// TODO: login - это CreateClient?
// TODO: выпилить devcontainer
// TODO: заменить сборку tdlib на готовый image
// TODO: повторить авторизацию для web-транспорта

// Repo предоставляет методы для взаимодействия с Telegram API через TDLib
type Repo struct {
	client                    *client.Client
	authorizationStateHandler *clientAuthorizer
	initClientDone            chan any
}

// New создает новый экземпляр репозитория Telegram
func New() *Repo {
	return &Repo{
		initClientDone: make(chan any),
	}
}

// Start устанавливает соединение с Telegram API
func (r *Repo) Start(ctx context.Context, cancel context.CancelFunc) error {
	// Создаем авторизатор только для инициализации
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

	go func() {
		slog.Info("Creating TDLib client")

		options := []client.Option{
			client.WithLogVerbosity(&client.SetLogVerbosityLevelRequest{
				NewVerbosityLevel: config.Telegram.LogVerbosityLevel,
			}),
		}

		for {
			r.authorizationStateHandler = NewClientAuthorizer(tdlibParameters, r.SetClient)
			tdlibClient, err := client.NewClient(r.authorizationStateHandler, options...)
			slog.Info("TDLib client created?")
			if err != nil {
				slog.Error("ошибка при создании клиента TDLib", "error", err)
				// return fmt.Errorf("ошибка при создании клиента TDLib: %w", err)
			}
			if tdlibClient != nil {
				r.SetClient(tdlibClient)
				break
			}
		}

		// Получаем информацию о версии TDLib
		versionOption, err := r.client.GetOption(&client.GetOptionRequest{
			Name: "version",
		})
		if err != nil {
			slog.Error("GetOption error", "error", err)
			// return fmt.Errorf("GetOption error: %w", err)
		}

		commitOption, err := r.client.GetOption(&client.GetOptionRequest{
			Name: "commit_hash",
		})
		if err != nil {
			slog.Error("GetOption error", "error", err)
			// return fmt.Errorf("GetOption error: %w", err)
		}

		slog.Info("TDLib",
			"version", versionOption.(*client.OptionValueString).Value,
			"commit", commitOption.(*client.OptionValueString).Value,
		)

		// Получаем информацию о пользователе
		me, err := r.client.GetMe()
		if err != nil {
			slog.Error("GetMe error", "error", err)
			// return fmt.Errorf("GetMe error: %w", err)
		}

		slog.Info("Me", "FirstName", me.FirstName) //, "LastName", me.LastName)

		// return nil
	}()

	return nil
}

func (r *Repo) GetPhoneNumber() chan string {
	return r.authorizationStateHandler.PhoneNumber
}

func (r *Repo) GetCode() chan string {
	return r.authorizationStateHandler.Code
}

func (r *Repo) GetStateChan() chan client.AuthorizationState {
	return r.authorizationStateHandler.State
}

func (r *Repo) GetPassword() chan string {
	return r.authorizationStateHandler.Password
}

// GetClient возвращает клиент TDLib
func (r *Repo) GetClient() *client.Client {
	return r.client
}

func (r *Repo) SetClient(client *client.Client) {
	slog.Info("SetClient")
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
	return nil
}

// InitClientDone возвращает канал, который будет закрыт после инициализации клиента
func (r *Repo) InitClientDone() chan any {
	return r.initClientDone
}

// GetMessage получает сообщение по идентификатору
func (r *Repo) GetMessage(chatID, messageID int64) (*client.Message, error) {
	// Реализация будет добавлена позже
	return &client.Message{}, nil
}

// SendTextMessage отправляет текстовое сообщение
func (r *Repo) SendMessage(chatID int64, text string) (*client.Message, error) {
	// Реализация будет добавлена позже
	return &client.Message{}, nil
}

// ForwardMessage пересылает сообщение
func (r *Repo) ForwardMessage(fromChatID, messageID int64, toChatID int64) (*client.Message, error) {
	// Реализация будет добавлена позже
	return &client.Message{}, nil
}

// DeleteMessage удаляет сообщение
func (r *Repo) DeleteMessage(chatID, messageID int64) error {
	// Реализация будет добавлена позже
	return nil
}

// EditMessage редактирует сообщение
func (r *Repo) EditMessage(chatID, messageID int64, text string) (*client.Message, error) {
	// Реализация будет добавлена позже
	return &client.Message{}, nil
}

// TODO: реализация ClientAuthorizer - это сервисный слой?

type setClient func(*client.Client)

type clientAuthorizer struct {
	SetClient       func(*client.Client)
	TdlibParameters *client.SetTdlibParametersRequest
	PhoneNumber     chan string
	Code            chan string
	State           chan client.AuthorizationState
	Password        chan string
}

func NewClientAuthorizer(tdlibParameters *client.SetTdlibParametersRequest, setClient func(*client.Client)) *clientAuthorizer {
	return &clientAuthorizer{
		SetClient:       setClient,
		TdlibParameters: tdlibParameters,
		PhoneNumber:     make(chan string, 1),
		Code:            make(chan string, 1),
		State:           make(chan client.AuthorizationState, 10),
		Password:        make(chan string, 1),
	}
}

func (stateHandler *clientAuthorizer) Handle(tdlibClient *client.Client, state client.AuthorizationState) error {
	slog.Info("Handle", "state", state.AuthorizationStateType())

	stateHandler.SetClient(tdlibClient) // dirty hack

	slog.Info("State send")
	stateHandler.State <- state
	slog.Info("State sent")

	switch state.AuthorizationStateType() {
	case client.TypeAuthorizationStateWaitTdlibParameters:
		_, err := tdlibClient.SetTdlibParameters(stateHandler.TdlibParameters)
		return err

	case client.TypeAuthorizationStateWaitPhoneNumber:
		_, err := tdlibClient.SetAuthenticationPhoneNumber(&client.SetAuthenticationPhoneNumberRequest{
			PhoneNumber: <-stateHandler.PhoneNumber,
			Settings: &client.PhoneNumberAuthenticationSettings{
				AllowFlashCall:       false,
				IsCurrentPhoneNumber: false,
				AllowSmsRetrieverApi: false,
			},
		})
		return err

	case client.TypeAuthorizationStateWaitEmailAddress:
		return client.NotSupportedAuthorizationState(state)

	case client.TypeAuthorizationStateWaitEmailCode:
		return client.NotSupportedAuthorizationState(state)

	case client.TypeAuthorizationStateWaitCode:
		_, err := tdlibClient.CheckAuthenticationCode(&client.CheckAuthenticationCodeRequest{
			Code: <-stateHandler.Code,
		})
		return err

	case client.TypeAuthorizationStateWaitOtherDeviceConfirmation:
		return client.NotSupportedAuthorizationState(state)

	case client.TypeAuthorizationStateWaitRegistration:
		return client.NotSupportedAuthorizationState(state)

	case client.TypeAuthorizationStateWaitPassword:
		_, err := tdlibClient.CheckAuthenticationPassword(&client.CheckAuthenticationPasswordRequest{
			Password: <-stateHandler.Password,
		})
		return err

	case client.TypeAuthorizationStateReady:
		return nil

	case client.TypeAuthorizationStateLoggingOut:
		return client.NotSupportedAuthorizationState(state)

	case client.TypeAuthorizationStateClosing:
		<-stateHandler.State // не допускаем переполнение (и блокировки чтения) канала
		return nil

	case client.TypeAuthorizationStateClosed:
		return nil
	}

	return client.NotSupportedAuthorizationState(state)
}

func (stateHandler *clientAuthorizer) Close() {
	slog.Info("Closing stateHandler")
	close(stateHandler.PhoneNumber)
	close(stateHandler.Code)
	close(stateHandler.State)
	close(stateHandler.Password)
}
