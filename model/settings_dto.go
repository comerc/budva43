package model

// SettingsDTO представляет собой объект передачи данных для настроек приложения
// Пример DTO, который используется для API и сохранения/загрузки конфигурации
type SettingsDTO struct {
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

// SettingsUpdateDTO представляет частичное обновление настроек
// Пример DTO для частичного обновления через API (PATCH)
type SettingsUpdateDTO struct {
	General    *GeneralSettingsUpdateDTO    `json:"general,omitempty"`
	Telegram   *TelegramSettingsUpdateDTO   `json:"telegram,omitempty"`
	Forwarding *ForwardingSettingsUpdateDTO `json:"forwarding,omitempty"`
	Reports    *ReportsSettingsUpdateDTO    `json:"reports,omitempty"`
	Storage    *StorageSettingsUpdateDTO    `json:"storage,omitempty"`
	Web        *WebSettingsUpdateDTO        `json:"web,omitempty"`
}

// Вспомогательные структуры для частичного обновления

type GeneralSettingsUpdateDTO struct {
	AutoStart     *bool   `json:"auto_start,omitempty"`
	NotifyOnStart *bool   `json:"notify_on_start,omitempty"`
	Language      *string `json:"language,omitempty"`
	Theme         *string `json:"theme,omitempty"`
	LogLevel      *string `json:"log_level,omitempty"`
}

type TelegramSettingsUpdateDTO struct {
	DatabaseDirectory          *string `json:"database_directory,omitempty"`
	FilesDirectory             *string `json:"files_directory,omitempty"`
	UseTestDc                  *bool   `json:"use_test_dc,omitempty"`
	UseChatInfoDatabase        *bool   `json:"use_chat_info_database,omitempty"`
	UseFileDatabase            *bool   `json:"use_file_database,omitempty"`
	UseMessageDatabase         *bool   `json:"use_message_database,omitempty"`
	DisableIntegrityProtection *bool   `json:"disable_integrity_protection,omitempty"`
	IgnoreFileNames            *bool   `json:"ignore_file_names,omitempty"`
}

type ForwardingSettingsUpdateDTO struct {
	DefaultDelay         *int  `json:"default_delay,omitempty"`
	MaxMessagesPerMinute *int  `json:"max_messages_per_minute,omitempty"`
	PreserveFormatting   *bool `json:"preserve_formatting,omitempty"`
	KeepMediaOriginal    *bool `json:"keep_media_original,omitempty"`
	AutoSign             *bool `json:"auto_sign,omitempty"`
	AddSourceLink        *bool `json:"add_source_link,omitempty"`
	AddForwardedTag      *bool `json:"add_forwarded_tag,omitempty"`
}

type ReportsSettingsUpdateDTO struct {
	DefaultPeriod     *string `json:"default_period,omitempty"`
	AutoGenerate      *bool   `json:"auto_generate,omitempty"`
	SendToAdmin       *bool   `json:"send_to_admin,omitempty"`
	IncludeStatistics *bool   `json:"include_statistics,omitempty"`
	StatFormat        *string `json:"stat_format,omitempty"`
	TemplatePath      *string `json:"template_path,omitempty"`
}

type StorageSettingsUpdateDTO struct {
	MaxCacheSize      *int64  `json:"max_cache_size,omitempty"`
	DataRetentionDays *int    `json:"data_retention_days,omitempty"`
	AutoCleanup       *bool   `json:"auto_cleanup,omitempty"`
	BackupEnabled     *bool   `json:"backup_enabled,omitempty"`
	BackupDirectory   *string `json:"backup_directory,omitempty"`
	BackupFrequency   *string `json:"backup_frequency,omitempty"`
}

type WebSettingsUpdateDTO struct {
	Enabled        *bool   `json:"enabled,omitempty"`
	Port           *int    `json:"port,omitempty"`
	Host           *string `json:"host,omitempty"`
	EnableTLS      *bool   `json:"enable_tls,omitempty"`
	CertFile       *string `json:"cert_file,omitempty"`
	KeyFile        *string `json:"key_file,omitempty"`
	RequireAuth    *bool   `json:"require_auth,omitempty"`
	SessionTimeout *int    `json:"session_timeout,omitempty"`
}

// NewDefaultSettingsDTO создает новый экземпляр DTO настроек с значениями по умолчанию
func NewDefaultSettingsDTO() *SettingsDTO {
	settings := &SettingsDTO{}

	// Установка значений по умолчанию для общих настроек
	settings.General.AutoStart = true
	settings.General.NotifyOnStart = true
	settings.General.Language = "en"
	settings.General.Theme = "light"
	settings.General.LogLevel = "info"

	// Значения по умолчанию для Telegram
	settings.Telegram.UseTestDc = false
	settings.Telegram.UseChatInfoDatabase = true
	settings.Telegram.UseFileDatabase = true
	settings.Telegram.UseMessageDatabase = true
	settings.Telegram.DisableIntegrityProtection = false
	settings.Telegram.IgnoreFileNames = false

	// Значения по умолчанию для пересылки
	settings.Forwarding.DefaultDelay = 3
	settings.Forwarding.MaxMessagesPerMinute = 20
	settings.Forwarding.PreserveFormatting = true
	settings.Forwarding.KeepMediaOriginal = true
	settings.Forwarding.AutoSign = false
	settings.Forwarding.AddSourceLink = true
	settings.Forwarding.AddForwardedTag = true

	// Значения по умолчанию для отчетов
	settings.Reports.DefaultPeriod = "daily"
	settings.Reports.AutoGenerate = false
	settings.Reports.SendToAdmin = false
	settings.Reports.IncludeStatistics = true
	settings.Reports.StatFormat = "text"

	// Значения по умолчанию для хранилища
	settings.Storage.MaxCacheSize = 1024 * 1024 * 100 // 100 MB
	settings.Storage.DataRetentionDays = 30
	settings.Storage.AutoCleanup = true
	settings.Storage.BackupEnabled = false

	// Значения по умолчанию для веб-интерфейса
	settings.Web.Enabled = true
	settings.Web.Port = 8080
	settings.Web.Host = "localhost"
	settings.Web.EnableTLS = false
	settings.Web.RequireAuth = true
	settings.Web.SessionTimeout = 60 // 60 минут

	return settings
}

// ApplyUpdate применяет частичное обновление к настройкам
func (s *SettingsDTO) ApplyUpdate(update *SettingsUpdateDTO) {
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

	// Аналогичное обновление для других секций
	// ...

	// В реальном приложении здесь был бы код для обновления
	// всех остальных секций настроек
}
