package config

import (
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/comerc/budva43/app/util"
)

func setDefaultConfig(config *config) {
	subproject := strings.TrimSpace(os.Getenv("SUBPROJECT"))
	if subproject == "" {
		subproject = "engine"
	}
	if !slices.Contains([]string{"engine", "facade"}, subproject) {
		log.Panic("invalid subproject: " + subproject + " (valid: engine, facade)")
	}
	logDir := filepath.Join(util.ProjectRoot, ".data", subproject, "log")
	testing := os.Getenv("GOEXPERIMENT") == "synctest"
	logLevel := slog.LevelDebug
	if testing {
		logLevel = slog.LevelError // !! в тестах проверяем только ошибки
	}

	config.General.EngineConfigFile = "engine.yml"

	config.General.Log.Level = logLevel
	config.General.Log.Directory = logDir
	config.General.Log.MaxFileSize = 10 // MB

	config.LogSource.RelativePath = true

	config.Telegram.UseTestDc = testing
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
	config.Telegram.DatabaseDirectory = filepath.Join(util.ProjectRoot, ".data", subproject, "telegram", "db")
	config.Telegram.FilesDirectory = filepath.Join(util.ProjectRoot, ".data", subproject, "telegram", "files")

	config.Storage.Log.Level = slog.LevelInfo
	config.Storage.Log.Directory = logDir
	config.Storage.Log.MaxFileSize = 10 // MB
	config.Storage.DatabaseDirectory = filepath.Join(util.ProjectRoot, ".data", subproject, "badger", "db")
	// config.Storage.BackupEnabled = false
	// config.Storage.BackupDirectory = filepath.Join(projectRoot, ".data", subproject, "badger", "backup")
	// config.Storage.BackupFrequency = "daily"

	config.Web.Port = "7070"
	config.Web.Host = "localhost" // V6 supported
	config.Web.ReadTimeout = 15 * time.Second
	config.Web.WriteTimeout = 15 * time.Second
	config.Web.ShutdownTimeout = 5 * time.Second
}
