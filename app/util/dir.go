package util

import (
	"log"
	"os"
)

// применяется в config до инициализации slog - поэтому не используем slog

func GetCurrDir() string {
	currDir, err := os.Getwd()
	if err != nil {
		log.Panicf("ошибка получения текущей директории %v", err)
	}
	return currDir
}

func RemoveDir(dir string) {
	err := os.RemoveAll(dir)
	if err != nil && !os.IsNotExist(err) {
		log.Panicf("ошибка удаления директории %s: %v", dir, err)
	}
}

func MakeDir(dir string) {
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
