package config

import (
	"fmt"
	"log"
	"log/slog"
	"os"
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
	// - BUDVA43__TELEGRAM__API_ID - из переменной окружения
	// - viper.GetString("telegram.api-id") - из конфигурационного файла

	// Автоматическое чтение из переменных окружения
	// viper.AutomaticEnv()
	// требуется костыль - ключ в config.yml,
	// иначе не читается из .env и выдает дефолтное значение

	// Задаём ключи для переопределения через .env
	viper.BindEnv("telegram.api-id")
	viper.BindEnv("telegram.api-hash")
	viper.BindEnv("telegram.phone-number")

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

	// Преобразуем относительные пути в абсолютные
	transformDirs()

	return config
}

// func getFlag(name string) *string {
// 	prefix := fmt.Sprintf("-%s=", name) // только для флагов через "="
// 	var result *string
// 	for _, arg := range os.Args {
// 		if strings.HasPrefix(arg, prefix) {
// 			v := arg[len(prefix):]
// 			result = &v
// 		}
// 	}
// 	return result
// }

func hasFlag(name string) bool {
	prefix := fmt.Sprintf("-%s", name)
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, prefix) {
			return true
		}
	}
	return false
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
	logDir := filepath.Join(projectRoot, ".data", "log")

	// var testVerbose *string // TODO: отказаться совсем от TestVerbose?
	// testVerbose = getFlag("test.v") // TODO: не работает для debug-сессии
	// config.General.TestVerbose = testVerbose

	config.General.LogLevel = slog.LevelDebug
	config.General.LogDirectory = logDir
	config.General.LogMaxFileSize = 10 // MB

	config.Telegram.UseTestDc = hasFlag("test.run")
	config.Telegram.UseFileDatabase = true
	config.Telegram.UseChatInfoDatabase = true
	config.Telegram.UseMessageDatabase = true
	config.Telegram.UseSecretChats = false
	config.Telegram.SystemLanguageCode = "en"
	config.Telegram.DeviceModel = "Server"
	config.Telegram.SystemVersion = "1.0.0"
	config.Telegram.ApplicationVersion = "1.0.0"
	config.Telegram.LogVerbosityLevel = 0
	config.Telegram.LogMaxFileSize = 10 // MB

	config.Telegram.LogDirectory = logDir
	config.Telegram.DatabaseDirectory = filepath.Join(projectRoot, ".data", "telegram", "db")
	config.Telegram.FilesDirectory = filepath.Join(projectRoot, ".data", "telegram", "files")

	config.Storage.LogLevel = slog.LevelInfo
	config.Storage.LogDirectory = logDir
	config.Storage.LogMaxFileSize = 10 // MB
	config.Storage.DatabaseDirectory = filepath.Join(projectRoot, ".data", "badger", "db")
	// config.Storage.BackupEnabled = false
	// config.Storage.BackupDirectory = filepath.Join(projectRoot, ".data", "badger", "backup")
	// config.Storage.BackupFrequency = "daily"

	config.Web.Port = 7070
	config.Web.Host = "localhost"
	config.Web.ReadTimeout = 15 * time.Second
	config.Web.WriteTimeout = 15 * time.Second
	config.Web.ShutdownTimeout = 5 * time.Second
}

// TODO: реализовать перезагрузку конфига при изменении файла
// func Watch(cb func(e fsnotify.Event)) {
// 	viper.OnConfigChange(cb)
// 	viper.WatchConfig()
// }
