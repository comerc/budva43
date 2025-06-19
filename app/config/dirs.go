package config

import (
	"path/filepath"

	"github.com/comerc/budva43/app/util"
)

// не используем slog, т.к. он инициализируется после конфига

func transformDirs() {
	for _, dirPtr := range dirPtrs {
		dir := *dirPtr
		// Устанавливаем директории относительно корня проекта, если они не абсолютные
		if !filepath.IsAbs(dir) {
			dir = filepath.Join(util.ProjectRoot, dir)
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
