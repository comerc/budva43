package config

import (
	"log/slog"
	"path/filepath"
	"time"

	"github.com/comerc/budva43/app/util"
)

func setDefaultConfig(config *config) {
	logDir := filepath.Join(util.ProjectRoot, ".data", "log")

	var testVerbose *string
	testVerbose = util.GetFlag("test.v") // не работает для debug-сессии!
	config.General.TestVerbose = testVerbose

	config.General.LogLevel = slog.LevelDebug
	config.General.LogDirectory = logDir
	config.General.LogMaxFileSize = 10 // MB

	config.Telegram.UseTestDc = util.HasFlag("test.run")
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
	config.Telegram.DatabaseDirectory = filepath.Join(util.ProjectRoot, ".data", "telegram", "db")
	config.Telegram.FilesDirectory = filepath.Join(util.ProjectRoot, ".data", "telegram", "files")

	config.Storage.LogLevel = slog.LevelInfo
	config.Storage.LogDirectory = logDir
	config.Storage.LogMaxFileSize = 10 // MB
	config.Storage.DatabaseDirectory = filepath.Join(util.ProjectRoot, ".data", "badger", "db")
	// config.Storage.BackupEnabled = false
	// config.Storage.BackupDirectory = filepath.Join(projectRoot, ".data", "badger", "backup")
	// config.Storage.BackupFrequency = "daily"

	config.Web.Port = 7070
	config.Web.Host = "localhost"
	config.Web.ReadTimeout = 15 * time.Second
	config.Web.WriteTimeout = 15 * time.Second
	config.Web.ShutdownTimeout = 5 * time.Second
}
