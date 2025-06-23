package config

import (
	"log/slog"
	"time"

	"github.com/comerc/budva43/app/entity"
)

var (
	cfg         = &config{}
	General     = &cfg.General
	ErrorSource = &cfg.ErrorSource
	Storage     = &cfg.Storage
	Telegram    = &cfg.Telegram
	Web         = &cfg.Web
	Engine      = &cfg.Engine
	// Reports = &cfg.Reports
)

type (
	// Настройки приложения
	config struct {
		General     general
		ErrorSource entity.ErrorSource
		Storage     storage
		Telegram    telegram
		Web         web
		Engine      entity.EngineConfig
		// Reports reports
	}

	// Общие настройки приложения
	general struct {
		TestVerbose      *string // nil - production mode & debug mode
		EngineConfigFile string
		Log              generalLog
	}

	// Настройки логирования приложения
	generalLog struct {
		Level       slog.Level
		Directory   string
		MaxFileSize int // MB
	}

	// Настройки хранилища данных
	storage struct {
		Log               storageLog
		DatabaseDirectory string
		// BackupEnabled     bool
		// BackupDirectory string
		// BackupFrequency string
	}

	// Настройки логирования хранилища данных
	storageLog struct {
		Level       slog.Level
		Directory   string
		MaxFileSize int // MB
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

	// Настройки веб-интерфейса
	web struct {
		Host            string
		Port            string
		ReadTimeout     time.Duration
		WriteTimeout    time.Duration
		ShutdownTimeout time.Duration
	}

	// Настройки отчетов
	// report struct {
	// 	Template string
	// 	For      []entity.ChatId
	// }
)
