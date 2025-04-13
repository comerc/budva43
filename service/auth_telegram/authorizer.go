package auth_telegram

import (
	"log/slog"

	"github.com/comerc/budva43/config"
	"github.com/zelenin/go-tdlib/client"
)

// Authorizer реализует интерфейс client.AuthorizationStateHandler
// для управления процессом авторизации в Telegram
type Authorizer struct {
	setClient       func(client *client.Client)
	TdlibParameters *client.SetTdlibParametersRequest
	PhoneNumber     chan string
	Code            chan string
	State           chan client.AuthorizationState
	Password        chan string
}

func NewAuthorizer(setClient func(*client.Client)) *Authorizer {
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

	return &Authorizer{
		setClient:       setClient,
		TdlibParameters: tdlibParameters,
		PhoneNumber:     make(chan string, 1),
		Code:            make(chan string, 1),
		State:           make(chan client.AuthorizationState, 10),
		Password:        make(chan string, 1),
	}
}

func (a *Authorizer) Handle(tdlibClient *client.Client, state client.AuthorizationState) error {
	slog.Info("Authorizer.Handle", "state", state.AuthorizationStateType())

	a.setClient(tdlibClient) // dirty hack

	slog.Info("State send")
	select {
	case a.State <- state:
		slog.Debug("State sent to channel")
	default:
		slog.Warn("State channel full, unable to send state")
	}

	switch state.AuthorizationStateType() {
	case client.TypeAuthorizationStateWaitTdlibParameters:
		_, err := tdlibClient.SetTdlibParameters(a.TdlibParameters)
		return err

	case client.TypeAuthorizationStateWaitPhoneNumber:
		_, err := tdlibClient.SetAuthenticationPhoneNumber(&client.SetAuthenticationPhoneNumberRequest{
			PhoneNumber: <-a.PhoneNumber,
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
			Code: <-a.Code,
		})
		return err

	case client.TypeAuthorizationStateWaitOtherDeviceConfirmation:
		return client.NotSupportedAuthorizationState(state)

	case client.TypeAuthorizationStateWaitRegistration:
		return client.NotSupportedAuthorizationState(state)

	case client.TypeAuthorizationStateWaitPassword:
		_, err := tdlibClient.CheckAuthenticationPassword(&client.CheckAuthenticationPasswordRequest{
			Password: <-a.Password,
		})
		return err

	case client.TypeAuthorizationStateReady:
		return nil

	case client.TypeAuthorizationStateLoggingOut:
		return client.NotSupportedAuthorizationState(state)

	case client.TypeAuthorizationStateClosing:
		<-a.State // не допускает переполнение (и блокировки чтения) канала
		return nil

	case client.TypeAuthorizationStateClosed:
		return nil
	}

	return client.NotSupportedAuthorizationState(state)
}

func (a *Authorizer) Close() {
	slog.Debug("Closing stateHandler")
	close(a.PhoneNumber)
	close(a.Code)
	close(a.State)
	close(a.Password)
}
