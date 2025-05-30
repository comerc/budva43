package config

import (
	"log"
	"os"
	"path/filepath"
)

func getAllDirs() []string {
	return []string{
		Storage.DatabaseDirectory,
		Storage.BackupDirectory,
		Telegram.LogDirectory,
		Telegram.DatabaseDirectory,
		Telegram.FilesDirectory,
	}
}

func RemoveDirs(dirs ...string) {
	if len(dirs) == 0 {
		dirs = getAllDirs()
	}
	for _, dir := range dirs {
		if !filepath.IsAbs(dir) {
			dir = filepath.Join(projectRoot, dir)
		}
		err := os.RemoveAll(dir)
		if err != nil && !os.IsNotExist(err) {
			log.Panicf("ошибка удаления директории %s: %v", dir, err)
		}
	}
}

func MakeDirs(dirs ...string) {
	if len(dirs) == 0 {
		dirs = getAllDirs()
	}
	for _, dir := range dirs {
		// Устанавливаем директории относительно корня проекта, если они не абсолютные
		if !filepath.IsAbs(dir) {
			dir = filepath.Join(projectRoot, dir)
		}
		_, err := os.Stat(dir)
		if os.IsNotExist(err) {
			if err := os.MkdirAll(dir, os.ModePerm); err != nil {
				log.Panicf("ошибка создания директории %s: %v", dir, err)
			}
		} else if err != nil {
			log.Panicf("ошибка проверки директории %s: %v", dir, err)
		}
		// Если директория существует, то ничего не делаем
	}
}
