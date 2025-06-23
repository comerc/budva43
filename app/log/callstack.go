package log

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/comerc/budva43/app/config"
	"github.com/comerc/budva43/app/entity"
)

var (
	// Кеш для корня проекта и имени модуля
	projectRoot     string
	projectModule   string
	projectRootOnce sync.Once
)

// CallInfo представляет информацию о вызове функции
type CallInfo struct {
	FuncName string // Название функции
	FileName string // Относительный путь к файлу
	Line     int    // Номер строки
}

// String возвращает компактное строковое представление информации о вызове
func (c CallInfo) String() string {
	return fmt.Sprintf("%s:%d %s", c.FileName, c.Line, c.FuncName)
}

// findProjectRootAndModule ищет корень проекта и читает имя модуля из go.mod
func findProjectRootAndModule() (string, string) {
	// Начинаем с текущей директории
	dir, err := os.Getwd()
	if err != nil {
		return "", ""
	}

	// Поднимаемся по директориям в поисках go.mod
	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			// Найден go.mod, читаем имя модуля
			moduleName := readModuleName(goModPath)
			return dir, moduleName
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Достигли корня файловой системы
			break
		}
		dir = parent
	}

	return "", ""
}

// readModuleName читает имя модуля из go.mod файла
func readModuleName(goModPath string) string {
	file, err := os.Open(goModPath)
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				return parts[1]
			}
		}
	}

	return ""
}

// initProjectInfo инициализирует информацию о проекте (кешируется)
func initProjectInfo() {
	projectRootOnce.Do(func() {
		projectRoot, projectModule = findProjectRootAndModule()
	})
}

// getProjectRoot возвращает корень проекта
func getProjectRoot() string {
	initProjectInfo()
	return projectRoot
}

// getProjectModule возвращает имя модуля проекта
func getProjectModule() string {
	initProjectInfo()
	return projectModule
}

// getFuncName извлекает короткое имя функции, убирая путь модуля
func getFuncName(fullFnName string) string {
	module := getProjectModule()
	if module == "" {
		return fullFnName
	}

	// Убираем префикс модуля (например, "github.com/comerc/budva43/")
	if strings.HasPrefix(fullFnName, module+"/") {
		funcName := strings.TrimPrefix(fullFnName, module+"/")

		// Убираем путь до пакета, оставляя только пакет.функция
		// Например: "app/log.(*SomeObject).NestedMethod" -> "log.(*SomeObject).NestedMethod"
		lastSlashIndex := strings.LastIndex(funcName, "/")
		if lastSlashIndex != -1 {
			funcName = funcName[lastSlashIndex+1:]
		}

		return funcName
	}

	// Если это прямо модуль без подпакетов
	if strings.HasPrefix(fullFnName, module+".") {
		return strings.TrimPrefix(fullFnName, module+".")
	}

	// Для функций main пакета возвращаем как есть
	if strings.HasPrefix(fullFnName, "main.") {
		return fullFnName
	}

	return fullFnName
}

// isProjectPath проверяет, принадлежит ли путь к текущему проекту
func isProjectPath(fullFnName string) bool {
	module := getProjectModule()
	if module == "" {
		return false
	}

	// Проверяем функции из подпакетов модуля
	if strings.HasPrefix(fullFnName, module) {
		return true
	}

	// Проверяем функции из main пакета (они имеют вид "main.FuncName")
	if strings.HasPrefix(fullFnName, "main.") {
		return true
	}

	return false
}

// getRelativePath возвращает относительный путь к файлу относительно корня проекта
func getRelativePath(fullPath string) string {
	root := getProjectRoot()
	if root == "" {
		// Если не нашли корень проекта, возвращаем только имя файла
		return filepath.Base(fullPath)
	}

	// Пытаемся получить относительный путь от корня проекта
	if rel, err := filepath.Rel(root, fullPath); err == nil {
		return rel
	}

	// Если не получилось, возвращаем только имя файла
	return filepath.Base(fullPath)
}

// getCallStack возвращает стек вызовов для логирования (только из текущего проекта)
// skip - количество фреймов для пропуска (обычно 1, чтобы пропустить саму эту функцию)
// depth - количество фреймов для сбора после пропуска (0 = все доступные фреймы проекта)
func GetCallStack(skip int) []*CallInfo {
	var result []*CallInfo

	depth := 10
	if config.ErrorSource.Type == entity.TypeErrorSourceOne {
		depth = 1
	}

	for i := skip; i < skip+depth; i++ {
		pc, fullPath, line, ok := runtime.Caller(i)
		if !ok {
			break
		}

		// Получаем информацию о функции
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}

		fullFnName := fn.Name()
		isProject := isProjectPath(fullFnName)

		// Если это не код проекта, прерываем поиск
		if !isProject {
			break
		}

		// Получаем короткое имя функции без полного пути модуля
		funcName := getFuncName(fullFnName)

		fileName := fullPath
		if config.ErrorSource.RelativePath {
			// Получаем относительный путь к файлу от корня проекта
			fileName = getRelativePath(fullPath)
		}

		result = append(result, &CallInfo{
			FuncName: funcName,
			FileName: fileName,
			Line:     line,
		})
	}

	return result
}
