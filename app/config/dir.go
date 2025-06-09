package config

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// не используем slog, т.к. он инициализируется после конфига

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

func getAllDirs() []string {
	return []string{
		LogOptions.Directory,
		Storage.LogDirectory,
		Storage.DatabaseDirectory,
		Storage.BackupDirectory,
		Telegram.LogDirectory,
		Telegram.DatabaseDirectory,
		Telegram.FilesDirectory,
	}
}

var projectRoot string

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
