package config

import (
	"log"
	"log/slog"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/mitchellh/mapstructure"
	"github.com/samber/lo"
	"github.com/spf13/viper"
)

func load() *config {
	// flag.Parse() // TODO: пока отказался от флагов, проблема с тестами - cobra?

	envPath := filepath.Join(projectRoot, ".env")
	if err := godotenv.Load(envPath); err != nil {
		log.Panic("не удалось загрузить .env файл: ", err)
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
		log.Panic("ошибка чтения конфигурации: ", err)
	}

	// Создаем конфигурацию с дефолтными значениями
	config := &config{}
	setDefaultConfig(config)

	// Настраиваем декодирование
	options := viper.DecodeHook(
		mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
			kebabCaseKeyHookFunc(),
		),
	)

	// Переопределяем значения из конфигурационного файла и переменных окружения
	if err := viper.Unmarshal(config, options); err != nil {
		log.Panic("ошибка разбора конфигурации: ", err)
	}

	return config
}

func kebabCaseKeyHookFunc() mapstructure.DecodeHookFunc {
	return func(from reflect.Type, _ reflect.Type, data any) (any, error) {
		if from.Kind() != reflect.Map {
			return data, nil
		}

		m, ok := data.(map[string]any)
		if !ok {
			return data, nil
		}

		// Создаем новую карту с преобразованными ключами
		out := make(map[string]any)
		for k, v := range m {
			// Преобразуем ключ из kebab-case в PascalCase
			pascalKey := lo.PascalCase(k)
			out[pascalKey] = v
		}
		return out, nil
	}
}

func setDefaultConfig(config *config) {
	// config.General.AutoStart = true
	// config.General.NotifyOnStart = true
	// config.General.Language = "en"
	// config.General.Theme = "light"

	config.LogOptions.Level = slog.LevelDebug

	config.Telegram.UseTestDc = false
	config.Telegram.UseFileDatabase = true
	config.Telegram.UseChatInfoDatabase = true
	config.Telegram.UseMessageDatabase = true
	config.Telegram.UseSecretChats = false
	config.Telegram.SystemLanguageCode = "en"
	config.Telegram.DeviceModel = "Server"
	config.Telegram.SystemVersion = "1.0.0"
	config.Telegram.ApplicationVersion = "1.0.0"
	config.Telegram.LogVerbosityLevel = 0
	config.Telegram.LogMaxFileSize = 10485760

	config.Telegram.LogDirectory = filepath.Join(projectRoot, ".data", "telegram", "log")
	config.Telegram.DatabaseDirectory = filepath.Join(projectRoot, ".data", "telegram", "db")
	config.Telegram.FilesDirectory = filepath.Join(projectRoot, ".data", "telegram", "files")

	config.Storage.DatabaseDirectory = filepath.Join(projectRoot, ".data", "storage")
	config.Storage.BackupDirectory = filepath.Join(projectRoot, ".data", "backups")

	// config.Forwarding.DefaultDelay = 3
	// config.Forwarding.MaxMessagesPerMinute = 20
	// config.Forwarding.PreserveFormatting = true
	// config.Forwarding.KeepMediaOriginal = true
	// config.Forwarding.AutoSign = false
	// config.Forwarding.AddSourceLink = true
	// config.Forwarding.AddForwardedTag = true

	// config.Reports.DefaultPeriod = "daily"
	// config.Reports.AutoGenerate = false
	// config.Reports.SendToAdmin = false
	// config.Reports.IncludeStatistics = true
	// config.Reports.StatFormat = "text"

	// config.Storage.MaxCacheSize = 1024 * 1024 * 100 // 100 MB
	// config.Storage.DataRetentionDays = 30
	// config.Storage.AutoCleanup = true
	// config.Storage.BackupEnabled = false

	// config.Web.Enabled = true
	config.Web.Port = 8080
	config.Web.Host = "localhost"
	config.Web.ReadTimeout = 15 * time.Second
	config.Web.WriteTimeout = 15 * time.Second
	config.Web.ShutdownTimeout = 5 * time.Second
	// config.Web.EnableTLS = false
	// config.Web.RequireAuth = true
	// config.Web.SessionTimeout = 60 * time.Minute
}
