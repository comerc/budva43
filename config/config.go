package config

import (
	"log"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

type (
	// Настройки приложения
	config struct {
		General    general
		LogOptions logOptions
		Telegram   telegram
		Forwarding forwarding
		Reports    reports
		Storage    storage
		Web        web
		Bot        bot
	}

	// Общие настройки
	general struct {
		AutoStart     bool
		NotifyOnStart bool
		Language      string
		Theme         string
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
		LogVerbosityLevel   int32
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

var (
	once        sync.Once
	cfg         = &config{}
	projectRoot string
	General     = &cfg.General
	LogOptions  = &cfg.LogOptions
	Telegram    = &cfg.Telegram
	Forwarding  = &cfg.Forwarding
	Reports     = &cfg.Reports
	Storage     = &cfg.Storage
	Web         = &cfg.Web
	Bot         = &cfg.Bot
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

var (
	allDirs = []string{
		Storage.DatabaseDirectory,
		Storage.BackupDirectory,
		Telegram.DatabaseDirectory,
		Telegram.FilesDirectory,
	}
)

func RemoveDirs(dirs ...string) {
	if len(dirs) == 0 {
		dirs = allDirs
	}
	for _, dir := range dirs {
		if !filepath.IsAbs(dir) {
			dir = filepath.Join(projectRoot, dir)
		}
		err := os.RemoveAll(dir)
		if err != nil && !os.IsNotExist(err) {
			log.Fatalf("ошибка удаления директории %s: %v", dir, err)
		}
	}
}

func MakeDirs(dirs ...string) {
	if len(dirs) == 0 {
		dirs = allDirs
	}
	for _, dir := range dirs {
		// Устанавливаем директории относительно корня проекта, если они не абсолютные
		if !filepath.IsAbs(dir) {
			dir = filepath.Join(projectRoot, dir)
		}
		_, err := os.Stat(dir)
		if os.IsNotExist(err) {
			if err := os.MkdirAll(dir, os.ModePerm); err != nil {
				log.Fatalf("ошибка создания директории %s: %v", dir, err)
			}
		} else if err != nil {
			log.Fatalf("ошибка проверки директории %s: %v", dir, err)
		}
		// Если директория существует, то ничего не делаем
	}
}

// init - это зло https://habr.com/ru/articles/771858/
// но подходит для реализации синглтона
func init() {
	once.Do(func() {
		projectRoot = findProjectRoot()
		loadedCfg, err := load()
		if err != nil {
			log.Fatalf("ошибка загрузки конфигурации: %s", err)
		}
		*cfg = *loadedCfg
		MakeDirs()
	})
}

func Watch(cb func(e fsnotify.Event)) {
	viper.OnConfigChange(cb)
	viper.WatchConfig()
}
