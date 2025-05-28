package util

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

var (
	// Кеш для корня проекта и имени модуля
	projectRoot     string
	projectModule   string
	projectRootOnce sync.Once
)

// CallInfo представляет информацию о вызове функции
type CallInfo struct {
	Function string // Название функции
	File     string // Относительный путь к файлу
	Line     int    // Номер строки
}

// String возвращает компактное строковое представление информации о вызове
func (c CallInfo) String() string {
	return fmt.Sprintf("%s:%d %s", c.File, c.Line, c.Function)
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

// isProjectPath проверяет, принадлежит ли путь к текущему проекту
func isProjectPath(funcName string) bool {
	module := getProjectModule()
	if module == "" {
		return false
	}
	return strings.HasPrefix(funcName, module)
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
func getCallStack(skip int, depth int) []CallInfo {
	var callStack []CallInfo

	// Если depth не указан, возвращаем максимум 10 фреймов
	if depth <= 0 {
		depth = 10
	}

	for i := skip; i < skip+depth; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}

		// Получаем информацию о функции
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}

		funcName := fn.Name()
		isProject := isProjectPath(funcName)

		// Если это не код проекта, прерываем поиск
		if !isProject {
			break
		}

		// Разделяем название функции на пакет и имя функции
		parts := strings.Split(funcName, ".")
		var functionName string

		if len(parts) >= 2 {
			functionName = parts[len(parts)-1]
		} else {
			functionName = funcName
		}

		// Получаем относительный путь к файлу от корня проекта
		fileName := getRelativePath(file)

		callStack = append(callStack, CallInfo{
			Function: functionName,
			File:     fileName,
			Line:     line,
		})
	}

	return callStack
}

// GetCaller возвращает информацию о вызывающем
func GetCaller() string {
	stack := getCallStack(2, 1) // Пропускаем GetCaller и getCallStack
	if len(stack) > 0 {
		return stack[0].String()
	}
	return ""
}

// GetCallers возвращает информацию о вызывающих
func GetCallers(stackDepth int) []string {
	stack := getCallStack(2, stackDepth) // Пропускаем GetCallers и getCallStack

	if len(stack) == 0 {
		return []string{}
	}

	var stackInfo []string
	for _, call := range stack {
		stackInfo = append(stackInfo, call.String())
	}

	return stackInfo
}
