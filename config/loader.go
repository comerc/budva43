package config

import (
	"flag"
	"fmt"
	"log/slog"
	"reflect"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/mitchellh/mapstructure"
	"github.com/samber/lo"
	"github.com/spf13/viper"
)

func setDefaultConfig(config *config) {
	config.General.AutoStart = true
	config.General.NotifyOnStart = true
	config.General.Language = "en"
	config.General.Theme = "light"
	config.General.LogLevel = "info"

	config.General.LogOptions.Level = slog.LevelInfo
	config.General.LogOptions.AddSource = false

	config.Telegram.UseTestDc = false
	config.Telegram.UseFileDatabase = true
	config.Telegram.UseChatInfoDatabase = true
	config.Telegram.UseMessageDatabase = true
	config.Telegram.UseSecretChats = false
	config.Telegram.SystemLanguageCode = "en"
	config.Telegram.DeviceModel = "Server"
	config.Telegram.SystemVersion = "1.0.0"
	config.Telegram.ApplicationVersion = "1.0.0"

	config.Forwarding.DefaultDelay = 3
	config.Forwarding.MaxMessagesPerMinute = 20
	config.Forwarding.PreserveFormatting = true
	config.Forwarding.KeepMediaOriginal = true
	config.Forwarding.AutoSign = false
	config.Forwarding.AddSourceLink = true
	config.Forwarding.AddForwardedTag = true

	config.Reports.DefaultPeriod = "daily"
	config.Reports.AutoGenerate = false
	config.Reports.SendToAdmin = false
	config.Reports.IncludeStatistics = true
	config.Reports.StatFormat = "text"

	config.Storage.MaxCacheSize = 1024 * 1024 * 100 // 100 MB
	config.Storage.DataRetentionDays = 30
	config.Storage.AutoCleanup = true
	config.Storage.BackupEnabled = false

	config.Web.Enabled = true
	config.Web.Port = 8080
	config.Web.Host = "localhost"
	config.Web.ReadTimeout = 15 * time.Second
	config.Web.WriteTimeout = 15 * time.Second
	config.Web.ShutdownTimeout = 5 * time.Second
	config.Web.EnableTLS = false
	config.Web.RequireAuth = true
	config.Web.SessionTimeout = 60 * time.Minute
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
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("ошибка загрузки переменных окружения: %w", err)
	}

	var configPath = flag.String("config", ".", "config path")

	flag.Parse()

	// Настройка Viper для чтения конфигурации из файла
	viper.SetConfigName("config") // имя конфигурационного файла без расширения
	viper.SetConfigType("yml")    // расширение файла конфигурации
	viper.AddConfigPath(*configPath)

	// Настраиваем Viper для правильной обработки имен полей и секций
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "__", "-", "_"))
	viper.SetEnvPrefix("BUDVA43_") // Префикс для переменных окружения
	// BUDVA43__GENERAL__TELEGRAM__API_ID == viper.GetString("general.telegram.api-id")

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
