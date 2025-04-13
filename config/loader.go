package config

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/mitchellh/mapstructure"
	"github.com/samber/lo"
	"github.com/spf13/viper"
)

func setDefaultConfig(config *config) {
	config.general.AutoStart = true
	config.general.NotifyOnStart = true
	config.general.Language = "en"
	config.general.Theme = "light"

	config.logOptions.Level = slog.LevelDebug
	config.logOptions.AddSource = false

	config.telegram.UseTestDc = testing.Testing()
	config.telegram.UseFileDatabase = true
	config.telegram.UseChatInfoDatabase = true
	config.telegram.UseMessageDatabase = true
	config.telegram.UseSecretChats = false
	config.telegram.SystemLanguageCode = "en"
	config.telegram.DeviceModel = "Server"
	config.telegram.SystemVersion = "1.0.0"
	config.telegram.ApplicationVersion = "1.0.0"
	config.telegram.LogVerbosityLevel = 0

	config.telegram.DatabaseDirectory = filepath.Join(projectRoot, "data", "tdlib")
	config.telegram.FilesDirectory = filepath.Join(projectRoot, "data", "tdlib_files")

	config.storage.DatabaseDirectory = filepath.Join(projectRoot, "data", "storage")
	config.storage.BackupDirectory = filepath.Join(projectRoot, "data", "backups")

	config.forwarding.DefaultDelay = 3
	config.forwarding.MaxMessagesPerMinute = 20
	config.forwarding.PreserveFormatting = true
	config.forwarding.KeepMediaOriginal = true
	config.forwarding.AutoSign = false
	config.forwarding.AddSourceLink = true
	config.forwarding.AddForwardedTag = true

	config.reports.DefaultPeriod = "daily"
	config.reports.AutoGenerate = false
	config.reports.SendToAdmin = false
	config.reports.IncludeStatistics = true
	config.reports.StatFormat = "text"

	config.storage.MaxCacheSize = 1024 * 1024 * 100 // 100 MB
	config.storage.DataRetentionDays = 30
	config.storage.AutoCleanup = true
	config.storage.BackupEnabled = false

	config.web.Enabled = true
	config.web.Port = 8080
	config.web.Host = "localhost"
	config.web.ReadTimeout = 15 * time.Second
	config.web.WriteTimeout = 15 * time.Second
	config.web.ShutdownTimeout = 5 * time.Second
	config.web.EnableTLS = false
	config.web.RequireAuth = true
	config.web.SessionTimeout = 60 * time.Minute
}

func kebabCaseKeyHookFunc() mapstructure.DecodeHookFunc {
	return func(from reflect.Type, to reflect.Type, data interface{}) (interface{}, error) {
		if from.Kind() != reflect.Map {
			return data, nil
		}

		m, ok := data.(map[string]interface{})
		if !ok {
			return data, nil
		}

		// Создаем новую карту с преобразованными ключами
		out := make(map[string]interface{})
		for k, v := range m {
			// Преобразуем ключ из kebab-case в PascalCase
			pascalKey := lo.PascalCase(k)
			out[pascalKey] = v
		}
		return out, nil
	}
}

func load() (*config, error) {
	// flag.Parse() // TODO: пока отказался от флагов, проблема с тестами

	envPath := filepath.Join(projectRoot, ".env")
	if err := godotenv.Load(envPath); err != nil {
		// log.Print("Не удалось загрузить .env файл %w", err)
		// Продолжаем выполнение
	}

	// Настройка Viper для чтения конфигурации из файла
	viper.SetConfigName("config") // имя конфигурационного файла без расширения
	viper.SetConfigType("yml")    // расширение файла конфигурации
	viper.AddConfigPath(projectRoot)

	// Настраиваем Viper для правильной обработки имен полей и секций
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "__", "-", "_"))
	viper.SetEnvPrefix("BUDVA43_") // Префикс для переменных окружения
	// одинаково работает:
	// - BUDVA43__GENERAL__TELEGRAM__API_ID - из переменной окружения
	// - viper.GetString("general.telegram.api-id") - из конфигурационного файла

	// Автоматическое чтение из переменных окружения
	viper.AutomaticEnv()

	// Читаем конфигурацию из файла
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("ошибка чтения конфигурации: %w", err)
	}

	options := viper.DecodeHook(
		mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
			kebabCaseKeyHookFunc(),
		),
	)

	// Создаем конфигурацию со значениями по умолчанию
	config := &config{}
	setDefaultConfig(config)

	// Переопределяем значения из конфигурационного файла и переменных окружения
	if err := viper.Unmarshal(config, options); err != nil {
		return nil, fmt.Errorf("ошибка разбора конфигурации: %w", err)
	}

	return config, nil
}
