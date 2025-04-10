package config

import (
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// TODO: log.Fatalf в этом файле используются много раз, нужно переделать на return error?

type (
	// Настройки приложения
	config struct {
		General    general
		Telegram   telegram
		Forwarding forwarding
		Reports    reports
		Storage    storage
		Web        web // TODO: Site || WebSite || Http
		Bot        bot
	}

	// Общие настройки
	general struct {
		LogOptions    logOptions
		AutoStart     bool
		NotifyOnStart bool
		Language      string
		Theme         string
		LogLevel      string
	}

	// Настройки логгера
	logOptions struct {
		Level     slog.Level
		AddSource bool
	}

	// Настройки Telegram
	telegram struct {
		ApiId               int32
		ApiHash             string
		PhoneNumber         string
		DatabaseDirectory   string
		FilesDirectory      string
		SystemLanguageCode  string
		DeviceModel         string
		SystemVersion       string
		ApplicationVersion  string
		UseTestDc           bool
		UseChatInfoDatabase bool
		UseFileDatabase     bool
		UseMessageDatabase  bool
		UseSecretChats      bool
	}

	// Настройки для бота
	bot struct {
		BotToken    string
		AdminChatId int64
	}

	// Настройки для пересылки сообщений
	forwarding struct {
		DefaultDelay         int
		MaxMessagesPerMinute int
		PreserveFormatting   bool
		KeepMediaOriginal    bool
		AutoSign             bool
		AddSourceLink        bool
		AddForwardedTag      bool
	}

	// Настройки для отчетов
	reports struct {
		DefaultPeriod     string
		AutoGenerate      bool
		SendToAdmin       bool
		IncludeStatistics bool
		StatFormat        string
		TemplateDirectory string
	}

	// Настройки хранилища данных
	storage struct {
		DatabaseDirectory string
		MaxCacheSize      int64
		DataRetentionDays int
		AutoCleanup       bool
		BackupEnabled     bool
		BackupDirectory   string
		BackupFrequency   string
	}

	// Настройки веб-интерфейса
	web struct {
		Enabled         bool
		Host            string
		Port            int
		ReadTimeout     time.Duration
		WriteTimeout    time.Duration
		ShutdownTimeout time.Duration
		EnableTLS       bool
		CertFile        string
		KeyFile         string
		RequireAuth     bool
		SessionTimeout  time.Duration
		AdminUsername   string
	}
)

var cfg *config
var General = cfg.General
var Telegram = cfg.Telegram
var Forwarding = cfg.Forwarding
var Reports = cfg.Reports
var Storage = cfg.Storage
var Web = cfg.Web
var Bot = cfg.Bot

func new() *config {
	result, err := load()
	if err != nil {
		log.Fatalf("ошибка загрузки конфигурации: %s", err)
	}
	return result
}

// TODO: куда бы пенести этот код? или тут ему место, т.к. тут же мы определили директории в конфиге
func makeDirs() {
	var dirs = []string{
		cfg.Storage.DatabaseDirectory,
		cfg.Storage.BackupDirectory,
		cfg.Telegram.DatabaseDirectory,
		cfg.Telegram.FilesDirectory,
	}
	for _, dir := range dirs {
		_, err := os.Stat(dir)
		if os.IsNotExist(err) {
			if err := os.Mkdir(dir, os.ModePerm); err != nil {
				log.Fatalf("ошибка создания директории: %s", err)
			}
		} else {
			log.Fatalf("ошибка проверки директории: %s", err)
		}
	}
}

func Init() {
	cfg = new()
	makeDirs()
}

func Watch(cb func(e fsnotify.Event)) {
	viper.OnConfigChange(cb)
	viper.WatchConfig()
}
