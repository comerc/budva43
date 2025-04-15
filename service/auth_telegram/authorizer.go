package auth_telegram

import (
	"log/slog"
	"time"

	"github.com/zelenin/go-tdlib/client"

	"github.com/comerc/budva43/config"
)

// Authorizer реализует интерфейс client.AuthorizationStateHandler
// для управления процессом авторизации в Telegram
type Authorizer struct {
	setClient       func(client *client.Client)
	shutdown        func()
	tdlibParameters *client.SetTdlibParametersRequest
	phoneNumber     chan string
	code            chan string
	state           chan client.AuthorizationState
	password        chan string
}

func NewAuthorizer(setClient func(*client.Client), shutdown func()) *Authorizer {

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
		shutdown:        shutdown,
		tdlibParameters: tdlibParameters,
		phoneNumber:     make(chan string, 1),
		code:            make(chan string, 1),
		state:           make(chan client.AuthorizationState, 10),
		password:        make(chan string, 1),
	}
}

func (a *Authorizer) Handle(tdlibClient *client.Client, state client.AuthorizationState) error {
	a.setClient(tdlibClient) // dirty hack - чтобы получить клиент до завершения client.NewClient()

	stateType := state.AuthorizationStateType()

	slog.Info("State send", "stateType", stateType)
	select {
	case a.state <- state:
		slog.Debug("State sent to channel")
	default:
		slog.Warn("State channel full, unable to send state")
	}

	switch stateType {
	case client.TypeAuthorizationStateWaitTdlibParameters:
		_, err := tdlibClient.SetTdlibParameters(a.tdlibParameters)
		if err != nil {
			slog.Error("ошибка при установке параметров TDLib", "error", err)

			a.shutdown()

			time.Sleep(1 * time.Second) // dirty hack
			return err
		}
		return nil

	case client.TypeAuthorizationStateWaitPhoneNumber:
		_, err := tdlibClient.SetAuthenticationPhoneNumber(&client.SetAuthenticationPhoneNumberRequest{
			PhoneNumber: <-a.phoneNumber,
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
			Code: <-a.code,
		})
		return err

	case client.TypeAuthorizationStateWaitOtherDeviceConfirmation:
		return client.NotSupportedAuthorizationState(state)

	case client.TypeAuthorizationStateWaitRegistration:
		return client.NotSupportedAuthorizationState(state)

	case client.TypeAuthorizationStateWaitPassword:
		_, err := tdlibClient.CheckAuthenticationPassword(&client.CheckAuthenticationPasswordRequest{
			Password: <-a.password,
		})
		return err

	case client.TypeAuthorizationStateReady:
		return nil

	case client.TypeAuthorizationStateLoggingOut:
		return client.NotSupportedAuthorizationState(state)

	case client.TypeAuthorizationStateClosing:
		<-a.state // не допускает переполнение (и блокировки чтения) канала
		return nil

	case client.TypeAuthorizationStateClosed:
		return nil
	}

	return client.NotSupportedAuthorizationState(state)
}

func (a *Authorizer) Close() {
	slog.Debug("Closing Authorizer")
	close(a.phoneNumber)
	close(a.code)
	close(a.state)
	close(a.password)
}
