package test

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zelenin/go-tdlib/client"

	config "github.com/comerc/budva43/config"
	authTelegramController "github.com/comerc/budva43/controller/auth_telegram"
	telegramRepo "github.com/comerc/budva43/repo/telegram"
	authTelegramService "github.com/comerc/budva43/service/auth_telegram"
	cliTransport "github.com/comerc/budva43/transport/cli"
)

func TestMain(m *testing.M) {
	config.Telegram.UseTestDc = true
	config.Telegram.DatabaseDirectory = "./test/.data/telegram/database/"
	config.Telegram.FilesDirectory = "./test/.data/telegram/files/"
	var dirs = []string{
		config.Telegram.DatabaseDirectory,
		config.Telegram.FilesDirectory,
	}
	config.MakeDirs(dirs...)
	code := m.Run()
	config.RemoveDirs(dirs...)
	os.Exit(code)
}

func TestAuthTelegram_InvalidTdlibParameters(t *testing.T) {
	tests := []struct {
		name  string
		setup func()
	}{
		{
			name: "invalid api id",
			setup: func() {
				config.Telegram.ApiId = 0
			},
		},
		{
			name: "invalid api hash",
			setup: func() {
				config.Telegram.ApiHash = ""
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setup()

			ctx, cancel := context.WithCancel(context.Background())

			telegramRepo := telegramRepo.New()
			err := telegramRepo.Start(ctx, cancel)
			require.NoError(t, err)
			defer telegramRepo.Stop()

			authTelegramService := authTelegramService.New(telegramRepo)
			require.NotNil(t, authTelegramService)

			select {
			case <-ctx.Done():
				// OK, контекст отменен
			case <-time.After(2 * time.Second):
				assert.Fail(t, "контекст не был отменен")
			}
		})
	}
}

// OutputLine представляет строку вывода CLI
type OutputLine struct {
	Line    string
	IsError bool
}

// CLIAutomator - структура для эмуляции ввода/вывода при тестировании CLI
type CLIAutomator struct {
	t              *testing.T
	originalStdin  *os.File
	originalStdout *os.File
	originalStderr *os.File
	stdinReader    *os.File
	stdinWriter    *os.File
	stdoutReader   *os.File
	stdoutWriter   *os.File
	stderrReader   *os.File
	stderrWriter   *os.File
	outputLines    chan OutputLine
	wg             sync.WaitGroup
	stdinPty       *os.File // Псевдо-терминал для эмуляции терминального ввода
}

// NewCLIAutomator создает экземпляр эмулятора CLI для интеграционного тестирования
func NewCLIAutomator(t *testing.T) *CLIAutomator {
	// Сохраняем оригинальные потоки ввода-вывода
	originalStdin := os.Stdin
	originalStdout := os.Stdout
	originalStderr := os.Stderr

	// Создаем пайпы для stdin
	stdinReader, stdinWriter, err := os.Pipe()
	require.NoError(t, err)

	// Создаем пайпы для stdout
	stdoutReader, stdoutWriter, err := os.Pipe()
	require.NoError(t, err)

	// Создаем пайпы для stderr
	stderrReader, stderrWriter, err := os.Pipe()
	require.NoError(t, err)

	// Перенаправляем стандартные потоки ввода-вывода
	os.Stdin = stdinReader
	os.Stdout = stdoutWriter
	os.Stderr = stderrWriter

	automator := &CLIAutomator{
		t:              t,
		originalStdin:  originalStdin,
		originalStdout: originalStdout,
		originalStderr: originalStderr,
		stdinReader:    stdinReader,
		stdinWriter:    stdinWriter,
		stdoutReader:   stdoutReader,
		stdoutWriter:   stdoutWriter,
		stderrReader:   stderrReader,
		stderrWriter:   stderrWriter,
		outputLines:    make(chan OutputLine, 100), // Буферизованный канал для строк вывода
	}

	// Запускаем горутины для чтения вывода
	automator.wg.Add(2)

	// Горутина для чтения stdout
	go func() {
		defer automator.wg.Done()
		scanner := bufio.NewScanner(stdoutReader)
		for scanner.Scan() {
			line := scanner.Text()
			select {
			case automator.outputLines <- OutputLine{Line: line, IsError: false}:
				// Строка добавлена в канал
			default:
				// Канал заполнен, пропускаем строку (или можно расширить буфер)
				t.Logf("Канал вывода заполнен, пропускаем строку: %s", line)
			}
		}
		if err := scanner.Err(); err != nil {
			t.Logf("Ошибка чтения stdout: %v", err)
		}
		close(automator.outputLines) // Закрываем канал после завершения чтения
	}()

	// Горутина для чтения stderr
	go func() {
		defer automator.wg.Done()
		scanner := bufio.NewScanner(stderrReader)
		for scanner.Scan() {
			line := scanner.Text()
			select {
			case automator.outputLines <- OutputLine{Line: line, IsError: true}:
				// Строка добавлена в канал
			default:
				// Канал заполнен, пропускаем строку
				t.Logf("Канал вывода заполнен, пропускаем строку ошибки: %s", line)
			}
		}
		if err := scanner.Err(); err != nil {
			t.Logf("Ошибка чтения stderr: %v", err)
		}
	}()

	// Эмуляция PTY не работает напрямую в Go, но мы пометим здесь, что это возможно реализовать
	// с использованием внешних библиотек, например https://github.com/creack/pty
	// automator.stdinPty, err = pty.Open()
	// require.NoError(t, err)

	return automator
}

// SendInput отправляет ввод в stdin CLI
func (c *CLIAutomator) SendInput(input string) {
	_, err := fmt.Fprintln(c.stdinWriter, input)
	require.NoError(c.t, err)
}

// PatternInfo содержит информацию о совпадении паттерна
type PatternInfo struct {
	Line  string
	Found bool
}

// WaitForOutput ожидает указанный вывод в течение таймаута
// Возвращает информацию о найденной строке или таймауте
func (c *CLIAutomator) WaitForOutput(pattern string, timeout time.Duration) PatternInfo {
	// deadline := time.Now().Add(timeout)

	// Создаем контекст с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			// Таймаут истек
			return PatternInfo{Found: false}
		case line, ok := <-c.outputLines:
			if !ok {
				// Канал закрыт, значит чтение завершено
				return PatternInfo{Found: false}
			}

			println("*****************")
			println(line.Line)
			println("*****************")
			// Проверяем совпадение с началом строки
			if strings.HasPrefix(line.Line, pattern) {
				return PatternInfo{Line: line.Line, Found: true}
			}

			// Выводим строку в лог тестов, если это ошибка
			if line.IsError {
				c.t.Logf("STDERR: %s", line.Line)
			}
		}

		// Дополнительная проверка таймаута
		// if time.Now().After(deadline) {
		// 	return PatternInfo{Found: false}
		// }
	}
}

// Cleanup восстанавливает стандартные потоки ввода-вывода
func (c *CLIAutomator) Cleanup() {
	// Закрываем пайпы записи для корректного завершения горутин чтения
	c.stdinWriter.Close()
	c.stdoutWriter.Close()
	c.stderrWriter.Close()

	// Ждем завершения горутин чтения вывода
	c.wg.Wait()

	// Восстанавливаем оригинальные потоки ввода-вывода
	os.Stdin = c.originalStdin
	os.Stdout = c.originalStdout
	os.Stderr = c.originalStderr

	// Закрываем оставшиеся дескрипторы
	c.stdinReader.Close()
	c.stdoutReader.Close()
	c.stderrReader.Close()

	// Закрываем PTY, если он был открыт
	if c.stdinPty != nil {
		c.stdinPty.Close()
	}
}

// SetupTermReadPassword настраивает обходное решение для проблемы с term.ReadPassword
// В реальном коде нужно либо модифицировать transport/cli/transport.go для поддержки
// тестового режима, либо использовать внешнюю библиотеку для эмуляции PTY
func (c *CLIAutomator) SetupTermReadPassword() {
	// В реальном коде здесь нужно использовать библиотеку для создания PTY
	// например, github.com/creack/pty

	// Примерный код:
	// 1. Создать PTY (pseudo-terminal)
	// 2. Перенаправить os.Stdin на PTY master
	// 3. Использовать PTY slave для чтения/записи данных

	// Для полной реализации требуется изменение кода transport/cli/transport.go
	// чтобы он принимал io.Reader/io.Writer вместо прямого использования term.ReadPassword
}

func TestAuthTelegram(t *testing.T) {
	if testing.Short() {
		t.Skip("Пропускаем интеграционный тест в коротком режиме")
	}

	automator := NewCLIAutomator(t)
	defer automator.Cleanup()

	// Настраиваем обходное решение для term.ReadPassword
	// automator.SetupTermReadPassword()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	telegramRepo := telegramRepo.New()
	err := telegramRepo.Start(ctx, cancel)
	require.NoError(t, err)
	defer telegramRepo.Stop()

	authTelegramService := authTelegramService.New(telegramRepo)
	require.NotNil(t, authTelegramService)

	time.Sleep(1 * time.Second) // TODO: dirty hack

	authTelegramController := authTelegramController.New(authTelegramService)
	require.NotNil(t, authTelegramController)

	state, err := authTelegramController.GetAuthorizationState()
	require.NoError(t, err)
	assert.Equal(t, client.TypeAuthorizationStateWaitPhoneNumber, state.AuthorizationStateType())

	cliTransport := cliTransport.New(
		nil, // messageController,
		nil, // forwardController,
		nil, // reportController,
		authTelegramController,
	)
	err = cliTransport.Start(ctx, cancel)
	require.NoError(t, err)
	defer cliTransport.Stop()

	result := automator.WaitForOutput("Запуск CLI интерфейса", 5*time.Second)
	require.True(t, result.Found, "CLI транспорт не запустился")
	// result = automator.WaitForOutput("> ", 3*time.Second)

	time.Sleep(2 * time.Second)

	automator.SendInput("auth")
	result = automator.WaitForOutput("Введите номер телефона: ", 3*time.Second)
	assert.True(t, result.Found, "Команда auth не выдала запрос на ввод номера телефона")

	// phoneNumber := "+70000000000"
	// automator.SendInput(phoneNumber)
	// time.Sleep(2 * time.Second)

	// automator.SendInput("auth")
	// result = automator.WaitForOutput("Введите номер телефона: ", 2*time.Second)
	// assert.True(t, result.Found, "2Команда auth не выдала запрос на ввод номера телефона")

	// result = automator.WaitForOutput("Введите код подтверждения: ", 2*time.Second)
	// assert.True(t, result.Found, "Команда auth не выдала запрос на ввод кода подтверждения")

	// code := "123456"
	// automator.SendInput(code)

	// automator.SendInput("help")
	// result = automator.WaitForOutput("Доступные команды", 2*time.Second)
	// assert.True(t, result.Found, "Команда help не выдала список команд")

	// Отправляем команду exit для завершения CLI
	// automator.SendInput("exit")
	// result = automator.WaitForOutput("Выход из программы", 2*time.Second)
	// assert.True(t, result.Found, "Команда exit не сработала")

	// select {
	// case <-ctx.Done():
	// 	// OK, контекст отменен
	// case <-time.After(2 * time.Second):
	// 	assert.Fail(t, "CLI не завершился после команды exit")
	// }
}

// Обратите внимание: term.ReadPassword не будет работать в тестах
// из-за отсутствия реального терминала, но мы все равно отправляем ввод
// phoneNumber := "+70000000000"
// automator.SendInput(phoneNumber)

// t.Log("Ожидаем запроса кода. Этот шаг может не сработать в тестовом режиме.")
// // TODO:

// // Завершаем тест
// t.Log("Завершение теста авторизации")
// automator.SendCommand("exit")
// found = automator.WaitForOutput("Выход из программы", 2*time.Second)
// assert.True(t, found, "Не получили сообщение о выходе")

// Выводим полный лог для отладки
// t.Logf("Полный вывод CLI:\n%s", automator.GetOutput())
