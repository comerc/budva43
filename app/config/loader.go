package config

import (
	"log"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"

	"github.com/comerc/budva43/app/util"
)

func load() *config {
	envPath := filepath.Join(util.ProjectRoot, ".config", ".private", ".env")
	if err := godotenv.Load(envPath); err != nil {
		log.Panic("не удалось загрузить .env файл: ", err)
	}

	// Настройка Viper для чтения конфигурации из файла
	viper.SetConfigName("app") // имя конфигурационного файла без расширения
	viper.SetConfigType("yml") // расширение файла конфигурации
	viper.AddConfigPath(filepath.Join(util.ProjectRoot, ".config"))

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
	_ = viper.BindEnv("telegram.api-id")
	_ = viper.BindEnv("telegram.api-hash")
	_ = viper.BindEnv("telegram.phone-number")
	_ = viper.BindEnv("general.engine-config-file")

	// Читаем конфигурацию из файла
	if err := viper.ReadInConfig(); err != nil {
		log.Panic("ошибка чтения конфигурации: ", err)
	}

	// Создаем конфигурацию с дефолтными значениями
	config := &config{}
	setDefaultConfig(config)

	// Переопределяем значения из конфигурационного файла и переменных окружения
	if err := viper.Unmarshal(config, util.GetConfigOptions()); err != nil {
		log.Panic("ошибка разбора конфигурации: ", err)
	}

	// Преобразуем относительные пути в абсолютные
	transformDirs()

	return config
}
