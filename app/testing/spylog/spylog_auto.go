package spylog

import (
	"github.com/comerc/budva43/app/log"
	// TODO: зависимость от app/log - вынести в отдельный пакет?
)

// GetHandler автоматически определяет loggerName из стека вызовов
func GetHandler(testName string, init func()) *PackageLogHandler {
	loggerName := log.GetPackageFileNameWithLine()
	// Устанавливаем loggerName для использования в NewLogger()
	log.SetLoggerName(loggerName)

	return getHandler(loggerName, testName, init)
}
