package util

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/creack/pty"
)

// CLIAutomator - структура для эмуляции ввода/вывода при тестировании CLI
type CLIAutomator struct {
	log            *slog.Logger
	originalStdin  *os.File
	originalStdout *os.File
	stdinReader    *os.File
	stdinWriter    *os.File
	stdoutReader   *os.File
	stdoutWriter   *os.File
	outputLines    chan string // Канал для вывода строк
}

// NewCLIAutomator создает экземпляр эмулятора CLI для интеграционного тестирования
func NewCLIAutomator() (*CLIAutomator, error) {
	// Сохраняем оригинальные потоки ввода-вывода
	originalStdin := os.Stdin
	originalStdout := os.Stdout

	// Создаем псевдо-терминал для term.ReadPassword
	stdinWriter, stdinReader, err := pty.Open()
	if err != nil {
		return nil, err
	}

	// Создаем пайпы для stdout
	stdoutReader, stdoutWriter, err := os.Pipe()
	if err != nil {
		return nil, err
	}

	automator := &CLIAutomator{
		log:            slog.With("module", "util.cli_automator"),
		originalStdin:  originalStdin,
		originalStdout: originalStdout,
		stdinReader:    stdinReader, // Подчиненная часть PTY (pseudo-terminal slave)
		stdinWriter:    stdinWriter, // Главная часть PTY (pseudo-terminal master)
		stdoutReader:   stdoutReader,
		stdoutWriter:   stdoutWriter,
		outputLines:    make(chan string, 100),
	}

	// Перенаправляем стандартные потоки ввода-вывода
	os.Stdin = automator.stdinReader
	os.Stdout = automator.stdoutWriter

	return automator, nil
}

// Start запускает обработку вывода CLI
func (c *CLIAutomator) Run() {
	scanner := bufio.NewScanner(c.stdoutReader)
	for scanner.Scan() {
		line := scanner.Text()
		// Безопасно добавляем строку в канал с расширением буфера при необходимости
		select {
		case c.outputLines <- line:
			// Строка добавлена в канал
		default:
			// Канал заполнен, расширяем буфер
			newBuffer := make(chan string, cap(c.outputLines)*2)
			close(c.outputLines) // Закрываем старый канал

			// Копируем все из старого канала в новый
			for oldLine := range c.outputLines {
				newBuffer <- oldLine
			}

			// Добавляем текущую строку
			newBuffer <- line
			c.outputLines = newBuffer
		}
	}

	if err := scanner.Err(); err != nil {
		if err != io.EOF && !errors.Is(err, os.ErrClosed) {
			c.log.Error("Ошибка чтения stdout", "error", err)
		}
	}

	// Закрываем канал после завершения чтения
	close(c.outputLines)
}

// SendInput отправляет ввод в stdin CLI
func (c *CLIAutomator) SendInput(input string) error {
	_, err := fmt.Fprintln(c.stdinWriter, input)
	return err
}

// WaitForOutput ожидает указанный вывод в течение таймаута
func (c *CLIAutomator) WaitForOutput(pattern string, timeout time.Duration) bool {
	// Создаем контекст с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			// Таймаут истек
			return false
		case line, ok := <-c.outputLines:
			if !ok {
				// Канал закрыт, значит чтение завершено
				return false
			}

			// Сначала проверяем начало строки
			if strings.HasPrefix(line, pattern) {
				return true
			}

			// Затем проверяем содержимое (deprecated)
			// if strings.Contains(line, pattern) {
			// 	return true
			// }
		}
	}
}

// Stop останавливает работу CLIAutomator и восстанавливает стандартные потоки ввода-вывода
func (c *CLIAutomator) Stop() {
	// Восстанавливаем оригинальные стандартные потоки ввода-вывода
	os.Stdin = c.originalStdin
	os.Stdout = c.originalStdout

	// Закрываем потоки ввода-вывода
	c.stdinReader.Close()
	c.stdinWriter.Close()
	c.stdoutWriter.Close()
	c.stdoutReader.Close()
}
