package config

import "time"

type (
	// Config представляет настройки приложения
	Config struct {
		General    General
		Telegram   Telegram
		Forwarding Forwarding
		Reports    Reports
		Storage    Storage
		Web        Web
	}

	// Общие настройки приложения
	General struct {
		AutoStart     bool
		NotifyOnStart bool
		Language      string
		Theme         string
		LogLevel      string
	}

	// Настройки Telegram
	Telegram struct {
		ApiID                      string
		ApiHash                    string
		BotToken                   string
		PhoneNumber                string
		DatabaseDirectory          string
		FilesDirectory             string
		UseTestDC                  bool
		UseChatInfoDatabase        bool
		UseFileDatabase            bool
		UseMessageDatabase         bool
		DisableIntegrityProtection bool
		IgnoreFileNames            bool
		AdminChatID                int64
	}

	// Настройки для пересылки сообщений
	Forwarding struct {
		DefaultDelay         int
		MaxMessagesPerMinute int
		PreserveFormatting   bool
		KeepMediaOriginal    bool
		AutoSign             bool
		AddSourceLink        bool
		AddForwardedTag      bool
	}

	// Настройки для отчетов
	Reports struct {
		DefaultPeriod     string
		AutoGenerate      bool
		SendToAdmin       bool
		IncludeStatistics bool
		StatFormat        string
		TemplatePath      string
	}

	// Настройки хранилища данных
	Storage struct {
		DatabasePath      string
		MaxCacheSize      int64
		DataRetentionDays int
		AutoCleanup       bool
		BackupEnabled     bool
		BackupDirectory   string
		BackupFrequency   string
	}

	// Настройки веб-интерфейса
	Web struct {
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
