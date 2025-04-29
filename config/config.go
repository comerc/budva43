package config

import (
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"

	"github.com/comerc/budva43/entity"
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
		Level     slog.Level
		AddSource bool
	}

	// Настройки хранилища данных
	storage struct {
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
		// Настройки для замены ссылок на себя
		ReplaceMyselfLinks map[entity.ChatId]*entity.ReplaceMyselfLink
		// Настройки для замены фрагментов текста
		ReplaceFragments map[entity.ChatId]*entity.ReplaceFragment
		// Настройки источников
		Sources map[entity.ChatId]*entity.Source
		// Настройки отчетов
		Reports struct {
			Template string
			For      []entity.ChatId
		}
		// Правила форвардинга
		Forwards map[entity.ForwardRuleId]*entity.ForwardRule
		// Настройки автоответов
		// Answers map[int64]entity.Answer
		// Удаление системных сообщений
		DeleteSystemMessages map[entity.ChatId]bool
	}
)

var (
	once        sync.Once
	cfg         = &config{}
	projectRoot string
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

// не используем slog, т.к. он инициализируется в main.go

// findProjectRoot находит корень проекта на основе файла go.mod
func findProjectRoot() string {
	// Запускаем команду "go env GOMOD" чтобы найти путь к go.mod
	cmd := exec.Command("go", "env", "GOMOD")
	output, err := cmd.Output()
	if err != nil {
		// log.Print("Не удалось определить путь к go.mod: %w", err)
		// Если не удалось, пробуем взять текущую директорию
		currentDir, err := os.Getwd()
		if err != nil {
			// log.Print("Не удалось получить текущую директорию: %w", err)
			return "."
		}
		return currentDir
	}

	// Удаляем символ новой строки из вывода
	goModPath := strings.TrimSpace(string(output))
	// Получаем директорию go.mod - это и есть корень проекта
	projectRoot := filepath.Dir(goModPath)

	return projectRoot
}

// init - это зло https://habr.com/ru/articles/771858/
// но подходит для реализации синглтона
func init() {
	once.Do(func() {
		projectRoot = findProjectRoot()
		*cfg = *load()
		MakeDirs()
	})
}

func Watch(cb func(e fsnotify.Event)) {
	viper.OnConfigChange(cb)
	viper.WatchConfig()
}
