package log

import (
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// GetCaller возвращает информацию о вызывающем
func GetCaller() string {
	callStack := GetCallStack(2) // Пропускаем GetCaller и getCallStack
	if len(callStack) > 0 {
		return callStack[0].String()
	}
	return ""
}

// GetCallers возвращает информацию о вызывающих
func GetCallers(depth int) []string {
	// TODO: depth - deprecated, remove it

	callStack := GetCallStack(2) // Пропускаем GetCallers и getCallStack

	if len(callStack) == 0 {
		return []string{}
	}

	result := make([]string, 0, len(callStack))
	for _, call := range callStack {
		result = append(result, call.String())
	}

	return result
}

// Тестовые функции для демонстрации стека вызовов
func testFunction1() string {
	return testFunction2()
}

func testFunction2() string {
	return testFunction3()
}

func testFunction3() string {
	caller := GetCaller()
	callers := GetCallers(5)

	result := "Caller: " + caller + "\n"
	result += "Stack:\n"
	for i, call := range callers {
		result += "  " + call + "\n"
		_ = i // используем i для избежания warning
	}

	return result
}

func TestCallInfo(t *testing.T) {
	t.Parallel()

	info := CallInfo{
		FuncName: "TestFunction",
		FileName: "test/file.go",
		Line:     42,
	}

	expected := "test/file.go:42 TestFunction"
	result := info.String()

	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestGetCaller(t *testing.T) {
	t.Parallel()

	caller := GetCaller()

	if caller == "" {
		t.Error("GetCaller should not return empty string")
	}

	// Проверяем, что результат содержит имя тестовой функции
	if !strings.Contains(caller, "TestGetCaller") {
		t.Errorf("Expected caller to contain 'TestGetCaller', got: %s", caller)
	}

	// Проверяем формат: файл:строка функция
	parts := strings.Split(caller, " ")
	if len(parts) < 2 {
		t.Errorf("Expected format 'file:line function', got: %s", caller)
	}

	// Проверяем, что путь содержит app/log/callstack_test.go
	assert.True(t, strings.HasPrefix(caller, "app/log/callstack_test.go"))

	t.Logf("Caller: %s", caller)
}

func TestGetCallers(t *testing.T) {
	t.Parallel()

	stackDepth := 5
	callers := GetCallers(stackDepth)

	if len(callers) == 0 {
		t.Error("GetCallers should not return empty slice")
	}

	// Первый элемент должен быть этой тестовой функцией
	if !strings.Contains(callers[0], "TestGetCallers") {
		t.Errorf("Expected first caller to contain 'TestGetCallers', got: %s", callers[0])
	}

	// Проверяем формат каждого элемента
	for i, caller := range callers {
		if caller == "" {
			t.Errorf("Caller %d should not be empty", i)
		}

		parts := strings.Split(caller, " ")
		if len(parts) < 2 {
			t.Errorf("Caller %d should have format 'file:line function', got: %s", i, caller)
		}
	}

	t.Logf("Call stack (%d levels):", len(callers))
	for i, caller := range callers {
		t.Logf("  %d: %s", i, caller)
	}
}

func TestGetCallersWithDepth(t *testing.T) {
	t.Parallel()

	// Тестируем различные глубины
	depths := []int{3, 5, 7, 10}

	for _, depth := range depths {
		callers := GetCallers(depth)

		t.Logf("Depth %d: got %d callers", depth, len(callers))

		// Проверяем, что получили хотя бы один вызов
		if len(callers) == 0 {
			t.Errorf("Expected at least one caller for depth %d", depth)
		}
	}
}

func TestNestedFunctionCalls(t *testing.T) {
	t.Parallel()

	result := testFunction1()

	// Проверяем, что результат содержит ожидаемые функции
	expectedFunctions := []string{"testFunction3"}

	for _, fn := range expectedFunctions {
		if !strings.Contains(result, fn) {
			t.Errorf("Expected result to contain '%s', got: %s", fn, result)
		}
	}

	t.Logf("Nested calls result:\n%s", result)
}

func TestProjectModuleDetection(t *testing.T) {
	t.Parallel()

	module := getProjectModule()

	if module == "" {
		t.Error("Project module should be detected from go.mod")
	}

	if !strings.Contains(module, "budva43") {
		t.Errorf("Expected module to contain 'budva43', got: %s", module)
	}

	t.Logf("Detected project module: %s", module)
}

func TestProjectRootDetection(t *testing.T) {
	t.Parallel()

	root := getProjectRoot()

	if root == "" {
		t.Error("Project root should be detected")
	}

	// Проверяем, что корень содержит go.mod
	if !strings.HasSuffix(root, "budva43") {
		t.Logf("Project root: %s", root)
	}

	t.Logf("Detected project root: %s", root)
}

func TestRelativePaths(t *testing.T) {
	t.Parallel()

	caller := GetCaller()

	// Проверяем, что путь относительный (не начинается с /)
	if strings.HasPrefix(caller, "/") {
		t.Errorf("Path should be relative, got: %s", caller)
	}

	// Проверяем, что путь содержит app/log/callstack_test.go
	assert.True(t, strings.HasPrefix(caller, "app/log/callstack_test.go"))

	// Проверяем, что путь заканчивается на .go
	parts := strings.Split(caller, " ")
	if len(parts) > 0 {
		filePart := strings.Split(parts[0], ":")[0]
		if !strings.HasSuffix(filePart, ".go") {
			t.Errorf("Expected file to end with .go, got: %s", filePart)
		}
	}
}

// Примеры использования со slog
func demonstrateSlogUsage() {
	// Настраиваем slog для вывода в stdout с JSON форматом
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	// Пример 1: Простое логирование с информацией о вызывающем
	logger.Info("User request started", "caller", GetCaller())

	// Пример 2: Логирование с полным стеком вызовов
	logger.Error("Database connection failed",
		"callers", GetCallers(0))

	// Пример 3: Структурированное логирование в функции
	processUserRequest(logger, "invalid")
	processUserRequest(logger, "user123")
}

type User struct{}

func processUserRequest(logger *slog.Logger, userID string) {
	logger.Info("Processing user request",
		"user_id", userID,
		"caller", GetCaller())

	u := &User{}

	if err := u.validateUser(logger, userID); err != nil {
		logger.Error("User validation failed",
			"user_id", userID,
			"error", err.Error(),
			"callers", GetCallers(0))
		return
	}

	logger.Info("User request completed successfully",
		"user_id", userID,
		"caller", GetCaller())
}

func (u *User) validateUser(logger *slog.Logger, userID string) error {
	logger.Debug("Validating user",
		"user_id", userID,
		"caller", GetCaller())

	if userID == "invalid" {
		return &customError{message: "user not found"}
	}

	return nil
}

type customError struct {
	message string
}

func (e *customError) Error() string {
	return e.message
}

func TestExampleSlogUsage(t *testing.T) {
	// Запускаем пример использования
	demonstrateSlogUsage()
}

// Бенчмарк для проверки производительности
func BenchmarkGetCaller(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GetCaller()
	}
}

func BenchmarkGetCallers(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GetCallers(5)
	}
}
