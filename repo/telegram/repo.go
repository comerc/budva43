package telegram

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/comerc/budva43/config"
	"github.com/zelenin/go-tdlib/client"
	"golang.org/x/term"
)

// Repo предоставляет методы для взаимодействия с Telegram API через TDLib
type Repo struct {
	client *client.Client
}

// New создает новый экземпляр репозитория Telegram
func New() (*Repo, error) {
	return &Repo{}, nil
}

// Start устанавливает соединение с Telegram API
func (r *Repo) Start(ctx context.Context) error {
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
	// client authorizer
	clientAuthorizer := client.ClientAuthorizer(tdlibParameters)
	// go client.CliInteractor(authorizer)

	// TODO: ввод данных авторизации tdlib сейчас реализован тут с привязкой к консоли,
	// хочется вынести в транспортные слои и сделать реализации для разных транспортов
	go func() {
		for {
			select {
			case state, ok := <-clientAuthorizer.State:
				if !ok {
					return
				}

				switch state.AuthorizationStateType() {
				case client.TypeAuthorizationStateWaitPhoneNumber:
					if config.Telegram.PhoneNumber != "" {
						fmt.Println("Используется номер телефона из конфигурации")
						time.Sleep(2 * time.Second)
						clientAuthorizer.PhoneNumber <- config.Telegram.PhoneNumber
						maskedPhone := maskPhoneNumber(config.Telegram.PhoneNumber)
						fmt.Println("Номер телефона:", maskedPhone)
					} else {
						fmt.Print("Введите номер телефона: ")
						var phoneNumber string
						fmt.Scanln(&phoneNumber)
						clientAuthorizer.PhoneNumber <- phoneNumber
						maskedPhone := maskPhoneNumber(phoneNumber)
						fmt.Println("Используется номер:", maskedPhone)
					}

				case client.TypeAuthorizationStateWaitCode:
					// var code string
					// fmt.Println("Enter code: ")
					// fmt.Scanln(&code)

					fmt.Print("Введите код подтверждения: ")
					code, err := term.ReadPassword(int(os.Stdin.Fd()))
					if err != nil {
						fmt.Println("\nОшибка при чтении кода:", err)
						continue
					}
					fmt.Println()

					clientAuthorizer.Code <- string(code)

				case client.TypeAuthorizationStateWaitPassword:
					// fmt.Println("Enter password: ")
					// var password string
					// fmt.Scanln(&password)

					fmt.Print("Введите пароль: ")
					password, err := term.ReadPassword(int(os.Stdin.Fd()))
					if err != nil {
						fmt.Println("\nОшибка при чтении пароля:", err)
						continue
					}
					fmt.Println()

					clientAuthorizer.Password <- string(password)

				case client.TypeAuthorizationStateReady:
					return
				}
			}
		}
	}()

	_, err := client.SetLogVerbosityLevel(&client.SetLogVerbosityLevelRequest{
		NewVerbosityLevel: 1,
	})
	if err != nil {
		return fmt.Errorf("SetLogVerbosityLevel error: %w", err)
	}

	tdlibClient, err := client.NewClient(clientAuthorizer)
	if err != nil {
		return fmt.Errorf("NewClient error: %w", err)
	}
	r.client = tdlibClient

	versionOption, err := client.GetOption(&client.GetOptionRequest{
		Name: "version",
	})
	if err != nil {
		return fmt.Errorf("GetOption error: %w", err)
	}

	commitOption, err := client.GetOption(&client.GetOptionRequest{
		Name: "commit_hash",
	})
	if err != nil {
		return fmt.Errorf("GetOption error: %w", err)
	}
	slog.Info("TDLib",
		"version", versionOption.(*client.OptionValueString).Value,
		"commit", commitOption.(*client.OptionValueString).Value,
	)

	// TODO: отсутствует client.TDLIB_VERSION
	// if commitOption.(*client.OptionValueString).Value != client.TDLIB_VERSION {
	// 	log.Printf("TDLib version supported by the library (%s) is not the same as TDLib version (%s)", client.TDLIB_VERSION, commitOption.(*client.OptionValueString).Value)
	// }

	// me, err := r.client.GetMe()
	// if err != nil {
	// 	return fmt.Errorf("GetMe error: %w", err)
	// }
	// slog.Info("Me", "FirstName", me.FirstName, "LastName", me.LastName)

	return nil
}

// Stop закрывает соединение с Telegram API
func (r *Repo) Stop() error {
	if r.client != nil {
		_, err := r.client.Close() // TODO: ok - зачем?
		if err != nil {
			return err
		}
		return nil
	}
	return nil
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

// GetChats получает список чатов
func (r *Repo) GetChats(limit int) ([]*client.Chat, error) {
	// Реализация будет добавлена позже
	return []*client.Chat{}, nil
}

// maskPhoneNumber маскирует номер телефона, заменяя 5 цифр перед последними двумя
// Например, +7 926 111 22 33 становится +7926*****33
func maskPhoneNumber(phone string) string {
	// Удаляем возможные пробелы и другие разделители
	cleanPhone := strings.ReplaceAll(strings.ReplaceAll(phone, " ", ""), "-", "")

	// Кол-во символов для маскирования
	const maskedCount = 5
	// Кол-во видимых символов в конце
	const visibleSuffixCount = 2

	// Проверяем минимальную длину номера
	if len(cleanPhone) <= maskedCount+visibleSuffixCount {
		// Если номер слишком короткий, показываем только последние 2 символа
		if len(cleanPhone) <= visibleSuffixCount {
			return "**" // Слишком короткий номер
		}
		// Маскируем все, кроме последних двух
		visibleSuffix := cleanPhone[len(cleanPhone)-visibleSuffixCount:]
		maskLength := len(cleanPhone) - visibleSuffixCount
		mask := strings.Repeat("*", maskLength)
		return mask + visibleSuffix
	}

	// Видимый префикс (всё, кроме последних 7 символов)
	prefixLength := len(cleanPhone) - maskedCount - visibleSuffixCount
	visiblePrefix := cleanPhone[:prefixLength]

	// Последние 2 символа
	visibleSuffix := cleanPhone[len(cleanPhone)-visibleSuffixCount:]

	// Маскированная часть (5 символов)
	mask := strings.Repeat("*", maskedCount)

	return visiblePrefix + mask + visibleSuffix
}
