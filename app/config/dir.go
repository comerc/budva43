package config

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/comerc/budva43/app/util"
)

// не используем slog, т.к. он инициализируется после конфига

func transformDirs() {
	for _, dirPtr := range dirPtrs {
		dir := *dirPtr
		// Устанавливаем директории относительно корня проекта, если они не абсолютные
		if !filepath.IsAbs(dir) {
			dir = filepath.Join(projectRoot, dir)
			*dirPtr = dir
		}
	}
}

func makeDirs() {
	for _, dirPtr := range dirPtrs {
		dir := *dirPtr
		util.MakeDir(dir)
	}
}

var dirPtrs = []*string{
	&General.LogDirectory,
	&Storage.LogDirectory,
	&Storage.DatabaseDirectory,
	// &Storage.BackupDirectory,
	&Telegram.LogDirectory,
	&Telegram.DatabaseDirectory,
	&Telegram.FilesDirectory,
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
