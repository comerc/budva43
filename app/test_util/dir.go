package test_util

import (
	"log"
	"os"
	"path/filepath"

	"github.com/comerc/budva43/app/config"
)

// не используем slog, т.к. он инициализируется после конфига

func InitConfigDirs() {
	currDir, err := os.Getwd()
	if err != nil {
		log.Panicf("ошибка получения текущей директории: %v", err)
	}
	logDir := filepath.Join(currDir, ".data", "log")
	config.LogOptions.Directory = logDir
	config.Storage.LogDirectory = logDir
	config.Storage.DatabaseDirectory = filepath.Join(currDir, ".data", "badger", "db")
	config.Storage.BackupDirectory = filepath.Join(currDir, ".data", "badger", "backup")
	config.Telegram.LogDirectory = logDir
	config.Telegram.DatabaseDirectory = filepath.Join(currDir, ".data", "telegram", "db")
	config.Telegram.FilesDirectory = filepath.Join(currDir, ".data", "telegram", "files")
	var dirs = []string{
		config.LogOptions.Directory,
		config.Storage.LogDirectory,
		config.Storage.DatabaseDirectory,
		config.Storage.BackupDirectory,
		config.Telegram.LogDirectory,
		config.Telegram.DatabaseDirectory,
		config.Telegram.FilesDirectory,
	}
	if config.LenAllDirs() != len(dirs) {
		log.Panic("количество директорий не совпадает")
	}
	config.RemoveDirs(dirs...)
	config.MakeDirs(dirs...)
}
