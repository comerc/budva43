package config

import (
	"log/slog"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"

	"github.com/comerc/budva43/app/entity"
)

var (
	cfg = &config{}
	// General     = &cfg.General
	LogOptions = &cfg.LogOptions
	Storage    = &cfg.Storage
	Telegram   = &cfg.Telegram
	Bot        = &cfg.Bot
	Web        = &cfg.Web
	// Forwarding = &cfg.Forwarding
	// Reports = &cfg.Reports
	Engine = &cfg.Engine
)

type (
	// Настройки приложения
	config struct {
		// General    general
		LogOptions logOptions
		Storage    storage
		Telegram   telegram
		Bot        bot
		Web        web
		// Forwarding forwarding
		// Reports reports
		Engine engine
	}

	// Общие настройки
	// general struct {
	// 	AutoStart     bool
	// 	NotifyOnStart bool
	// 	Language      string
	// 	Theme         string
	// }

	// Настройки логгера
	logOptions struct {
		Level       slog.Level
		Directory   string
		MaxFileSize int // MB
	}

	// Настройки хранилища данных
	storage struct {
		LogLevel          slog.Level
		LogDirectory      string
		LogMaxFileSize    int // MB
		DatabaseDirectory string
		// MaxCacheSize      int64
		// DataRetentionDays int
		// AutoCleanup       bool
		// BackupEnabled     bool
		BackupDirectory string
		BackupFrequency string
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
		LogVerbosityLevel   int32
		LogDirectory        string
		LogMaxFileSize      int // MB
	}

	// Настройки для бота
	bot struct {
		BotToken    string
		AdminChatId int64
	}

	// Настройки веб-интерфейса
	web struct {
		// Enabled         bool
		Host            string
		Port            int
		ReadTimeout     time.Duration
		WriteTimeout    time.Duration
		ShutdownTimeout time.Duration
		// EnableTLS       bool
		// CertFile        string
		// KeyFile         string
		// RequireAuth     bool
		// SessionTimeout  time.Duration
		// AdminUsername   string
	}

	// Настройки для пересылки сообщений
	// forwarding struct {
	// 	DefaultDelay         int
	// 	MaxMessagesPerMinute int
	// 	PreserveFormatting   bool
	// 	KeepMediaOriginal    bool
	// 	AutoSign             bool
	// 	AddSourceLink        bool
	// 	AddForwardedTag      bool
	// }

	// Настройки для отчетов
	// reports struct {
	// 	DefaultPeriod     string
	// 	AutoGenerate      bool
	// 	SendToAdmin       bool
	// 	IncludeStatistics bool
	// 	StatFormat        string
	// 	TemplateDirectory string
	// }

	// Настройки движка форвардинга из budva32
	engine struct {
		// Настройки получателей
		Destinations map[entity.ChatId]*entity.Destination
		// Настройки источников
		Sources map[entity.ChatId]*entity.Source
		// Правила форвардинга
		ForwardRules map[entity.ForwardRuleId]*entity.ForwardRule
		// Настройки отчетов
		// Report struct {
		// 	Template string
		// 	For      []entity.ChatId
		// }
		// Уникальные источники
		UniqueSources map[entity.ChatId]struct{} `mapstructure:"-"`
		// Порядок форвардинга
		OrderedForwardRules []entity.ForwardRuleId `mapstructure:"-"`
	}
)

func Watch(cb func(e fsnotify.Event)) {
	viper.OnConfigChange(cb)
	viper.WatchConfig()
}
