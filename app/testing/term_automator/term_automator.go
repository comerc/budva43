package term_automator

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/comerc/budva43/app/log"
)

// TODO: переделать на использование goroutine-local stdin/stdout

// TermAutomator - структура для эмуляции ввода/вывода при тестировании терминала
type TermAutomator struct {
	log *log.Logger
	//
	originalStdin  *os.File
	originalStdout *os.File
	stdinReader    *os.File
	stdinWriter    *os.File
	stdoutReader   *os.File
	stdoutWriter   *os.File
	outputLines    chan string // Канал для вывода строк
}

// NewTermAutomator создает экземпляр эмулятора терминала для интеграционного тестирования
func NewTermAutomator() (*TermAutomator, error) {
	// Сохраняем оригинальные потоки ввода-вывода
	originalStdin := os.Stdin
	originalStdout := os.Stdout

	// Создаем пайпы для stdin
	stdinReader, stdinWriter, err := os.Pipe()
	if err != nil {
		return nil, err
	}

	// Создаем пайпы для stdout
	stdoutReader, stdoutWriter, err := os.Pipe()
	if err != nil {
		return nil, err
	}

	automator := &TermAutomator{
		log: log.NewLogger(),
		//
		originalStdin:  originalStdin,
		originalStdout: originalStdout,
		stdinReader:    stdinReader,
		stdinWriter:    stdinWriter,
		stdoutReader:   stdoutReader,
		stdoutWriter:   stdoutWriter,
		outputLines:    make(chan string, 100),
	}

	// Перенаправляем стандартные потоки ввода-вывода
	os.Stdin = automator.stdinReader
	os.Stdout = automator.stdoutWriter

	return automator, nil
}

// Run запускает обработку вывода
func (a *TermAutomator) Run() {
	var err error
	defer a.log.ErrorOrDebug(&err, "")

	scanner := bufio.NewScanner(a.stdoutReader)
	for scanner.Scan() {
		line := scanner.Text()
		// Безопасно добавляем строку в канал с расширением буфера при необходимости
		select {
		case a.outputLines <- line:
			// Строка добавлена в канал
		default:
			// Канал заполнен, расширяем буфер
			newBuffer := make(chan string, cap(a.outputLines)*2)
			close(a.outputLines) // Закрываем старый канал

			// Копируем все из старого канала в новый
			for oldLine := range a.outputLines {
				newBuffer <- oldLine
			}

			// Добавляем текущую строку
			newBuffer <- line
			a.outputLines = newBuffer
		}
	}

	if err = scanner.Err(); err != nil {
		if err != io.EOF && !errors.Is(err, os.ErrClosed) {
			err = log.WrapError(err) // внешняя ошибка
		} else {
			err = nil
		}
	}

	// Закрываем канал после завершения чтения
	close(a.outputLines)
}

// SendInput отправляет ввод в stdin
func (a *TermAutomator) SendInput(input string) error {
	_, err := fmt.Fprintln(a.stdinWriter, input)
	return err
}

// WaitForOutput ожидает указанный вывод в течение таймаута
func (a *TermAutomator) WaitForOutput(ctx context.Context, pattern string, timeout time.Duration) bool {
	// Создаем контекст с таймаутом
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			// Таймаут истек
			return false
		case line, ok := <-a.outputLines:
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

// Close останавливает работу и восстанавливает стандартные потоки ввода-вывода
func (a *TermAutomator) Close() {
	// Восстанавливаем оригинальные стандартные потоки ввода-вывода
	os.Stdin = a.originalStdin
	os.Stdout = a.originalStdout

	// Закрываем потоки ввода-вывода в правильном порядке
	// Сначала закрываем writer'ы чтобы прекратить запись и дать signal читателям
	if a.stdoutWriter != nil {
		a.stdoutWriter.Close()
		a.stdoutWriter = nil
	}
	if a.stdinWriter != nil {
		a.stdinWriter.Close()
		a.stdinWriter = nil
	}

	// Затем закрываем reader'ы (это должно разблокировать горутину Run)
	if a.stdoutReader != nil {
		a.stdoutReader.Close()
		a.stdoutReader = nil
	}
	if a.stdinReader != nil {
		a.stdinReader.Close()
		a.stdinReader = nil
	}
}
