package entity

import "encoding/json"

// Settings представляет настройки приложения
type Settings struct {
	// Общие настройки приложения
	General struct {
		AutoStart     bool   `json:"auto_start"`
		NotifyOnStart bool   `json:"notify_on_start"`
		Language      string `json:"language"`
		Theme         string `json:"theme"`
		LogLevel      string `json:"log_level"`
	} `json:"general"`

	// Настройки Telegram
	Telegram struct {
		ApiId                      int64  `json:"api_id"`
		ApiHash                    string `json:"api_hash"`
		PhoneNumber                string `json:"phone_number,omitempty"`
		DatabaseDirectory          string `json:"database_directory"`
		FilesDirectory             string `json:"files_directory"`
		UseTestDc                  bool   `json:"use_test_dc"`
		UseChatInfoDatabase        bool   `json:"use_chat_info_database"`
		UseFileDatabase            bool   `json:"use_file_database"`
		UseMessageDatabase         bool   `json:"use_message_database"`
		DisableIntegrityProtection bool   `json:"disable_integrity_protection"`
		IgnoreFileNames            bool   `json:"ignore_file_names"`
	} `json:"telegram"`

	// Настройки для пересылки сообщений
	Forwarding struct {
		DefaultDelay         int  `json:"default_delay"`
		MaxMessagesPerMinute int  `json:"max_messages_per_minute"`
		PreserveFormatting   bool `json:"preserve_formatting"`
		KeepMediaOriginal    bool `json:"keep_media_original"`
		AutoSign             bool `json:"auto_sign"`
		AddSourceLink        bool `json:"add_source_link"`
		AddForwardedTag      bool `json:"add_forwarded_tag"`
	} `json:"forwarding"`

	// Настройки для отчетов
	Reports struct {
		DefaultPeriod     string `json:"default_period"`
		AutoGenerate      bool   `json:"auto_generate"`
		SendToAdmin       bool   `json:"send_to_admin"`
		IncludeStatistics bool   `json:"include_statistics"`
		StatFormat        string `json:"stat_format"`
		TemplatePath      string `json:"template_path"`
	} `json:"reports"`

	// Настройки хранилища данных
	Storage struct {
		DatabasePath      string `json:"database_path"`
		MaxCacheSize      int64  `json:"max_cache_size"`
		DataRetentionDays int    `json:"data_retention_days"`
		AutoCleanup       bool   `json:"auto_cleanup"`
		BackupEnabled     bool   `json:"backup_enabled"`
		BackupDirectory   string `json:"backup_directory,omitempty"`
		BackupFrequency   string `json:"backup_frequency,omitempty"`
	} `json:"storage"`

	// Настройки веб-интерфейса
	Web struct {
		Enabled        bool   `json:"enabled"`
		Port           int    `json:"port"`
		Host           string `json:"host"`
		EnableTLS      bool   `json:"enable_tls"`
		CertFile       string `json:"cert_file,omitempty"`
		KeyFile        string `json:"key_file,omitempty"`
		RequireAuth    bool   `json:"require_auth"`
		SessionTimeout int    `json:"session_timeout"`
		AdminUsername  string `json:"admin_username,omitempty"`
	} `json:"web"`
}

// SettingsUpdate представляет частичное обновление настроек
type SettingsUpdate struct {
	General    *GeneralSettingsUpdate    `json:"general,omitempty"`
	Telegram   *TelegramSettingsUpdate   `json:"telegram,omitempty"`
	Forwarding *ForwardingSettingsUpdate `json:"forwarding,omitempty"`
	Reports    *ReportsSettingsUpdate    `json:"reports,omitempty"`
	Storage    *StorageSettingsUpdate    `json:"storage,omitempty"`
	Web        *WebSettingsUpdate        `json:"web,omitempty"`
}

// Вспомогательные структуры для частичного обновления

type GeneralSettingsUpdate struct {
	AutoStart     *bool   `json:"auto_start,omitempty"`
	NotifyOnStart *bool   `json:"notify_on_start,omitempty"`
	Language      *string `json:"language,omitempty"`
	Theme         *string `json:"theme,omitempty"`
	LogLevel      *string `json:"log_level,omitempty"`
}

type TelegramSettingsUpdate struct {
	DatabaseDirectory          *string `json:"database_directory,omitempty"`
	FilesDirectory             *string `json:"files_directory,omitempty"`
	UseTestDc                  *bool   `json:"use_test_dc,omitempty"`
	UseChatInfoDatabase        *bool   `json:"use_chat_info_database,omitempty"`
	UseFileDatabase            *bool   `json:"use_file_database,omitempty"`
	UseMessageDatabase         *bool   `json:"use_message_database,omitempty"`
	DisableIntegrityProtection *bool   `json:"disable_integrity_protection,omitempty"`
	IgnoreFileNames            *bool   `json:"ignore_file_names,omitempty"`
}

type ForwardingSettingsUpdate struct {
	DefaultDelay         *int  `json:"default_delay,omitempty"`
	MaxMessagesPerMinute *int  `json:"max_messages_per_minute,omitempty"`
	PreserveFormatting   *bool `json:"preserve_formatting,omitempty"`
	KeepMediaOriginal    *bool `json:"keep_media_original,omitempty"`
	AutoSign             *bool `json:"auto_sign,omitempty"`
	AddSourceLink        *bool `json:"add_source_link,omitempty"`
	AddForwardedTag      *bool `json:"add_forwarded_tag,omitempty"`
}

type ReportsSettingsUpdate struct {
	DefaultPeriod     *string `json:"default_period,omitempty"`
	AutoGenerate      *bool   `json:"auto_generate,omitempty"`
	SendToAdmin       *bool   `json:"send_to_admin,omitempty"`
	IncludeStatistics *bool   `json:"include_statistics,omitempty"`
	StatFormat        *string `json:"stat_format,omitempty"`
	TemplatePath      *string `json:"template_path,omitempty"`
}

type StorageSettingsUpdate struct {
	MaxCacheSize      *int64  `json:"max_cache_size,omitempty"`
	DataRetentionDays *int    `json:"data_retention_days,omitempty"`
	AutoCleanup       *bool   `json:"auto_cleanup,omitempty"`
	BackupEnabled     *bool   `json:"backup_enabled,omitempty"`
	BackupDirectory   *string `json:"backup_directory,omitempty"`
	BackupFrequency   *string `json:"backup_frequency,omitempty"`
}

type WebSettingsUpdate struct {
	Enabled        *bool   `json:"enabled,omitempty"`
	Port           *int    `json:"port,omitempty"`
	Host           *string `json:"host,omitempty"`
	EnableTLS      *bool   `json:"enable_tls,omitempty"`
	CertFile       *string `json:"cert_file,omitempty"`
	KeyFile        *string `json:"key_file,omitempty"`
	RequireAuth    *bool   `json:"require_auth,omitempty"`
	SessionTimeout *int    `json:"session_timeout,omitempty"`
}

// MarshalJSON реализует интерфейс json.Marshaler для Settings
func (s Settings) MarshalJSON() ([]byte, error) {
	type SettingsAlias Settings
	return json.Marshal((*SettingsAlias)(&s))
}

// UnmarshalJSON реализует интерфейс json.Unmarshaler для Settings
func (s *Settings) UnmarshalJSON(data []byte) error {
	type SettingsAlias Settings
	alias := (*SettingsAlias)(s)

	if err := json.Unmarshal(data, alias); err != nil {
		return err
	}

	return nil
}

// // NewDefaultSettings создает новый экземпляр настроек с значениями по умолчанию
// func NewDefaultSettings() *Settings {
// 	settings := &Settings{}

// 	// Установка значений по умолчанию для общих настроек
// 	settings.General.AutoStart = true
// 	settings.General.NotifyOnStart = true
// 	settings.General.Language = "en"
// 	settings.General.Theme = "light"
// 	settings.General.LogLevel = "info"

// 	// Значения по умолчанию для Telegram
// 	settings.Telegram.UseTestDc = false
// 	settings.Telegram.UseChatInfoDatabase = true
// 	settings.Telegram.UseFileDatabase = true
// 	settings.Telegram.UseMessageDatabase = true
// 	settings.Telegram.DisableIntegrityProtection = false
// 	settings.Telegram.IgnoreFileNames = false

// 	// Значения по умолчанию для пересылки
// 	settings.Forwarding.DefaultDelay = 3
// 	settings.Forwarding.MaxMessagesPerMinute = 20
// 	settings.Forwarding.PreserveFormatting = true
// 	settings.Forwarding.KeepMediaOriginal = true
// 	settings.Forwarding.AutoSign = false
// 	settings.Forwarding.AddSourceLink = true
// 	settings.Forwarding.AddForwardedTag = true

// 	// Значения по умолчанию для отчетов
// 	settings.Reports.DefaultPeriod = "daily"
// 	settings.Reports.AutoGenerate = false
// 	settings.Reports.SendToAdmin = false
// 	settings.Reports.IncludeStatistics = true
// 	settings.Reports.StatFormat = "text"

// 	// Значения по умолчанию для хранилища
// 	settings.Storage.MaxCacheSize = 1024 * 1024 * 100 // 100 MB
// 	settings.Storage.DataRetentionDays = 30
// 	settings.Storage.AutoCleanup = true
// 	settings.Storage.BackupEnabled = false

// 	// Значения по умолчанию для веб-интерфейса
// 	settings.Web.Enabled = true
// 	settings.Web.Port = 8080
// 	settings.Web.Host = "localhost"
// 	settings.Web.EnableTLS = false
// 	settings.Web.RequireAuth = true
// 	settings.Web.SessionTimeout = 60 // 60 минут

// 	return settings
// }

// ApplyUpdate применяет частичное обновление к настройкам
func (s *Settings) ApplyUpdate(update *SettingsUpdate) {
	if update == nil {
		return
	}

	// Обновление общих настроек
	if update.General != nil {
		if update.General.AutoStart != nil {
			s.General.AutoStart = *update.General.AutoStart
		}
		if update.General.NotifyOnStart != nil {
			s.General.NotifyOnStart = *update.General.NotifyOnStart
		}
		if update.General.Language != nil {
			s.General.Language = *update.General.Language
		}
		if update.General.Theme != nil {
			s.General.Theme = *update.General.Theme
		}
		if update.General.LogLevel != nil {
			s.General.LogLevel = *update.General.LogLevel
		}
	}

	// Обновление настроек Telegram
	if update.Telegram != nil {
		if update.Telegram.DatabaseDirectory != nil {
			s.Telegram.DatabaseDirectory = *update.Telegram.DatabaseDirectory
		}
		if update.Telegram.FilesDirectory != nil {
			s.Telegram.FilesDirectory = *update.Telegram.FilesDirectory
		}
		if update.Telegram.UseTestDc != nil {
			s.Telegram.UseTestDc = *update.Telegram.UseTestDc
		}
		if update.Telegram.UseChatInfoDatabase != nil {
			s.Telegram.UseChatInfoDatabase = *update.Telegram.UseChatInfoDatabase
		}
		if update.Telegram.UseFileDatabase != nil {
			s.Telegram.UseFileDatabase = *update.Telegram.UseFileDatabase
		}
		if update.Telegram.UseMessageDatabase != nil {
			s.Telegram.UseMessageDatabase = *update.Telegram.UseMessageDatabase
		}
		if update.Telegram.DisableIntegrityProtection != nil {
			s.Telegram.DisableIntegrityProtection = *update.Telegram.DisableIntegrityProtection
		}
		if update.Telegram.IgnoreFileNames != nil {
			s.Telegram.IgnoreFileNames = *update.Telegram.IgnoreFileNames
		}
	}

	// Обновление настроек пересылки
	if update.Forwarding != nil {
		if update.Forwarding.DefaultDelay != nil {
			s.Forwarding.DefaultDelay = *update.Forwarding.DefaultDelay
		}
		if update.Forwarding.MaxMessagesPerMinute != nil {
			s.Forwarding.MaxMessagesPerMinute = *update.Forwarding.MaxMessagesPerMinute
		}
		if update.Forwarding.PreserveFormatting != nil {
			s.Forwarding.PreserveFormatting = *update.Forwarding.PreserveFormatting
		}
		if update.Forwarding.KeepMediaOriginal != nil {
			s.Forwarding.KeepMediaOriginal = *update.Forwarding.KeepMediaOriginal
		}
		if update.Forwarding.AutoSign != nil {
			s.Forwarding.AutoSign = *update.Forwarding.AutoSign
		}
		if update.Forwarding.AddSourceLink != nil {
			s.Forwarding.AddSourceLink = *update.Forwarding.AddSourceLink
		}
		if update.Forwarding.AddForwardedTag != nil {
			s.Forwarding.AddForwardedTag = *update.Forwarding.AddForwardedTag
		}
	}

	// Обновление настроек отчетов
	if update.Reports != nil {
		if update.Reports.DefaultPeriod != nil {
			s.Reports.DefaultPeriod = *update.Reports.DefaultPeriod
		}
		if update.Reports.AutoGenerate != nil {
			s.Reports.AutoGenerate = *update.Reports.AutoGenerate
		}
		if update.Reports.SendToAdmin != nil {
			s.Reports.SendToAdmin = *update.Reports.SendToAdmin
		}
		if update.Reports.IncludeStatistics != nil {
			s.Reports.IncludeStatistics = *update.Reports.IncludeStatistics
		}
		if update.Reports.StatFormat != nil {
			s.Reports.StatFormat = *update.Reports.StatFormat
		}
		if update.Reports.TemplatePath != nil {
			s.Reports.TemplatePath = *update.Reports.TemplatePath
		}
	}

	// Обновление настроек хранилища
	if update.Storage != nil {
		if update.Storage.MaxCacheSize != nil {
			s.Storage.MaxCacheSize = *update.Storage.MaxCacheSize
		}
		if update.Storage.DataRetentionDays != nil {
			s.Storage.DataRetentionDays = *update.Storage.DataRetentionDays
		}
		if update.Storage.AutoCleanup != nil {
			s.Storage.AutoCleanup = *update.Storage.AutoCleanup
		}
		if update.Storage.BackupEnabled != nil {
			s.Storage.BackupEnabled = *update.Storage.BackupEnabled
		}
		if update.Storage.BackupDirectory != nil {
			s.Storage.BackupDirectory = *update.Storage.BackupDirectory
		}
		if update.Storage.BackupFrequency != nil {
			s.Storage.BackupFrequency = *update.Storage.BackupFrequency
		}
	}

	// Обновление настроек веб-интерфейса
	if update.Web != nil {
		if update.Web.Enabled != nil {
			s.Web.Enabled = *update.Web.Enabled
		}
		if update.Web.Port != nil {
			s.Web.Port = *update.Web.Port
		}
		if update.Web.Host != nil {
			s.Web.Host = *update.Web.Host
		}
		if update.Web.EnableTLS != nil {
			s.Web.EnableTLS = *update.Web.EnableTLS
		}
		if update.Web.CertFile != nil {
			s.Web.CertFile = *update.Web.CertFile
		}
		if update.Web.KeyFile != nil {
			s.Web.KeyFile = *update.Web.KeyFile
		}
		if update.Web.RequireAuth != nil {
			s.Web.RequireAuth = *update.Web.RequireAuth
		}
		if update.Web.SessionTimeout != nil {
			s.Web.SessionTimeout = *update.Web.SessionTimeout
		}
	}
}
