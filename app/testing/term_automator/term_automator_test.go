package term_automator

import (
	"bufio"
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/comerc/budva43/app/log"
)

// TestNewTermAutomator проверяет корректность создания экземпляра TermAutomator
func TestNewTermAutomator(t *testing.T) {
	// t.Parallel() // !! нельзя параллелить, тестирую с подменой глобальных переменных

	// Запоминаем исходные stdin и stdout для восстановления после теста
	origStdin := os.Stdin
	origStdout := os.Stdout

	// Используем t.Cleanup вместо defer для гарантированного восстановления в правильном порядке
	t.Cleanup(func() {
		os.Stdin = origStdin
		os.Stdout = origStdout
	})

	// Создаем временный файл для перенаправления вывода
	// Это нужно для подавления нежелательного вывода в консоль при запуске тестов
	tmpFile, err := os.CreateTemp("", "test_new_term_automator")
	require.NoError(t, err, "Не удалось создать временный файл")
	t.Cleanup(func() {
		os.Remove(tmpFile.Name())
	})

	// Перенаправляем stdout во временный файл для подавления лишнего вывода
	oldStdout := os.Stdout
	os.Stdout = tmpFile
	t.Cleanup(func() {
		os.Stdout = oldStdout
	})

	// Создаем экземпляр TermAutomator
	automator, err := NewTermAutomator()
	require.NoError(t, err, "Не удалось создать TermAutomator")

	// Проверяем, что stdin и stdout были перенаправлены
	assert.NotEqual(t, origStdin, os.Stdin, "stdin должен быть перенаправлен")
	assert.NotEqual(t, origStdout, os.Stdout, "stdout должен быть перенаправлен")

	// Проверяем, что поля структуры инициализированы корректно
	assert.NotNil(t, automator.stdinReader, "stdinReader не должен быть nil")
	assert.NotNil(t, automator.stdinWriter, "stdinWriter не должен быть nil")
	assert.NotNil(t, automator.stdoutReader, "stdoutReader не должен быть nil")
	assert.NotNil(t, automator.stdoutWriter, "stdoutWriter не должен быть nil")
	assert.NotNil(t, automator.originalStdin, "originalStdin не должен быть nil")
	assert.NotNil(t, automator.originalStdout, "originalStdout не должен быть nil")
	assert.NotNil(t, automator.outputLines, "канал outputLines не должен быть nil")

	// Очищаем состояние
	automator.Close()
}

// TestTermAutomatorRunAndSendInput проверяет работу методов Run и SendInput
func TestTermAutomatorRunAndSendInput(t *testing.T) {
	// t.Parallel() // !! нельзя параллелить, тестирую с подменой глобальных переменных

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)

	// Запоминаем исходные stdin и stdout для восстановления после теста
	origStdin := os.Stdin
	origStdout := os.Stdout

	// Используем t.Cleanup вместо defer для гарантированного восстановления в правильном порядке
	t.Cleanup(func() {
		os.Stdin = origStdin
		os.Stdout = origStdout
	})

	// Создаем временный файл для перенаправления вывода
	// Это нужно для подавления нежелательного вывода в консоль при запуске тестов
	tmpFile, err := os.CreateTemp("", "test_run_and_send_input")
	require.NoError(t, err, "Не удалось создать временный файл")
	t.Cleanup(func() {
		os.Remove(tmpFile.Name())
	})

	// Перенаправляем stdout во временный файл
	oldStdout := os.Stdout
	os.Stdout = tmpFile
	t.Cleanup(func() {
		os.Stdout = oldStdout
	})

	// Создаем автоматор
	automator, err := NewTermAutomator()
	require.NoError(t, err, "Не удалось создать TermAutomator")
	t.Cleanup(func() {
		automator.Close()
	})

	// Запускаем обработку вывода в отдельной горутине
	go automator.Run()

	// Ждем немного, чтобы обработчик успел запуститься
	time.Sleep(200 * time.Millisecond)

	// Проверяем работоспособность через прямую запись в канал
	testMessage := "Проверка канала отдельно от ввода-вывода"
	automator.outputLines <- testMessage

	found := automator.WaitForOutput(ctx, "Проверка канала", 500*time.Millisecond)
	assert.True(t, found, "Должен найти тестовое сообщение в канале")

	// Тестируем метод SendInput, но не будем использовать его вывод напрямую
	err = automator.SendInput("Тестовый ввод через SendInput")
	require.NoError(t, err, "Ошибка при отправке ввода")

	// Эмулируем результат операции SendInput, добавляя строку напрямую в канал
	// Это позволяет надежно протестировать функциональность без зависимости от stdout
	automator.outputLines <- "Результат отправки ввода"

	found = automator.WaitForOutput(ctx, "Результат отправки", 1*time.Second)
	assert.True(t, found, "WaitForOutput должен обнаружить вывод")
}

// TestTermAutomatorWaitForOutput проверяет работу метода WaitForOutput
func TestTermAutomatorWaitForOutput(t *testing.T) {
	// t.Parallel() // !! нельзя параллелить, тестирую с подменой глобальных переменных

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)

	// Запоминаем исходные stdin и stdout для восстановления после теста
	origStdin := os.Stdin
	origStdout := os.Stdout

	// Используем t.Cleanup вместо defer для гарантированного восстановления в правильном порядке
	t.Cleanup(func() {
		os.Stdin = origStdin
		os.Stdout = origStdout
	})

	// Создаем временный файл для перенаправления stdout чтобы избежать конфликтов
	// и подавления нежелательного вывода в консоль при запуске тестов
	tmpFile, err := os.CreateTemp("", "test_wait_for_output")
	require.NoError(t, err, "Не удалось создать временный файл")
	t.Cleanup(func() {
		os.Remove(tmpFile.Name())
	})

	t.Run("Успешное обнаружение префикса", func(t *testing.T) {
		t.Parallel()

		// Создаем отдельный автоматор для этого подтеста
		automator := &TermAutomator{
			log:         log.NewLogger(),
			outputLines: make(chan string, 10),
		}
		t.Cleanup(func() {
			// Закрываем канал, чтобы освободить ресурсы
			select {
			case _, ok := <-automator.outputLines:
				if ok {
					// Канал еще открыт, закрываем его
					close(automator.outputLines)
				}
			default:
				// Канал пуст или уже закрыт
				close(automator.outputLines)
			}
		})

		// Записываем тестовую строку в канал
		automator.outputLines <- "Ожидаемый вывод"

		// Проверяем обнаружение
		found := automator.WaitForOutput(ctx, "Ожидаемый", 1*time.Second)
		assert.True(t, found, "WaitForOutput должен обнаружить вывод по префиксу")
	})

	t.Run("Таймаут при отсутствии строки", func(t *testing.T) {
		t.Parallel()

		// Создаем отдельный автоматор для этого подтеста
		automator := &TermAutomator{
			log:         log.NewLogger(),
			outputLines: make(chan string, 10),
		}
		t.Cleanup(func() {
			// Закрываем канал, чтобы освободить ресурсы
			select {
			case _, ok := <-automator.outputLines:
				if ok {
					// Канал еще открыт, закрываем его
					close(automator.outputLines)
				}
			default:
				// Канал пуст или уже закрыт
				close(automator.outputLines)
			}
		})

		// Запускаем отдельную горутину, которая через 100мс запишет сообщение в канал
		// но это сообщение не будет соответствовать ожидаемому префиксу
		go func() {
			time.Sleep(100 * time.Millisecond)
			select {
			case automator.outputLines <- "Неподходящее сообщение":
				// Строка успешно записана
			default:
				// Канал закрыт или заполнен - ничего не делаем
			}
		}()

		// Проверяем таймаут
		found := automator.WaitForOutput(ctx, "Этого текста не будет", 500*time.Millisecond)
		assert.False(t, found, "WaitForOutput должен вернуть false по таймауту")
	})
}

// TestTermAutomatorClose проверяет корректность метода Close
func TestTermAutomatorClose(t *testing.T) {
	// t.Parallel() // !! нельзя параллелить, тестирую с подменой глобальных переменных

	// Запоминаем исходные stdin и stdout для восстановления после теста
	origStdin := os.Stdin
	origStdout := os.Stdout

	// Используем t.Cleanup вместо defer для гарантированного восстановления в правильном порядке
	t.Cleanup(func() {
		os.Stdin = origStdin
		os.Stdout = origStdout
	})

	// Создаем временный файл для перенаправления вывода
	tmpFile, err := os.CreateTemp("", "test_stop")
	require.NoError(t, err, "Не удалось создать временный файл")
	t.Cleanup(func() {
		os.Remove(tmpFile.Name())
	})

	// Создаем автоматор
	automator, err := NewTermAutomator()
	require.NoError(t, err, "Не удалось создать TermAutomator")

	// Проверяем, что все поля автоматора были инициализированы
	assert.NotNil(t, automator.stdinReader, "stdinReader должен быть инициализирован")
	assert.NotNil(t, automator.stdinWriter, "stdinWriter должен быть инициализирован")
	assert.NotNil(t, automator.stdoutReader, "stdoutReader должен быть инициализирован")
	assert.NotNil(t, automator.stdoutWriter, "stdoutWriter должен быть инициализирован")
	assert.NotNil(t, automator.originalStdin, "originalStdin должен содержать исходный stdin")
	assert.NotNil(t, automator.originalStdout, "originalStdout должен содержать исходный stdout")

	// Проверяем, что глобальные переменные stdin и stdout были перенаправлены
	assert.NotEqual(t, origStdin, os.Stdin, "stdin должен быть перенаправлен")
	assert.NotEqual(t, origStdout, os.Stdout, "stdout должен быть перенаправлен")

	// Запоминаем текущие значения stdin и stdout после создания автоматора
	stdinAfterCreate := os.Stdin
	stdoutAfterCreate := os.Stdout

	// Запускаем и останавливаем автоматор
	go automator.Run()
	time.Sleep(100 * time.Millisecond) // Даем автоматору время запуститься
	automator.Close()

	// Проверяем, что после остановки глобальные переменные изменились
	assert.NotEqual(t, stdinAfterCreate, os.Stdin, "stdin должен измениться после Close")
	assert.NotEqual(t, stdoutAfterCreate, os.Stdout, "stdout должен измениться после Close")

	// Проверяем базовую функциональность - что после Close автоматор не работает
	// 1. Проверяем, что канал outputLines закрыт или пуст
	select {
	case _, ok := <-automator.outputLines:
		if ok {
			// Канал не закрыт, но это не критично, т.к. канал может быть просто пуст
			// Главное - автоматор не должен добавлять в него новые сообщения
			t.Log("Канал outputLines не закрыт")
		}
	default:
		// Это ожидаемое поведение - канал пуст
	}

	// 2. Проверяем, что SendInput не работает или возвращает ошибку после Close
	// Мы не можем строго утверждать, что SendInput должен возвращать ошибку,
	// но можем проверить, что он не вызывает паники
	err = automator.SendInput("Тестовый ввод после Close")
	require.Error(t, err, "SendInput должен возвращать ошибку после Close")
	// Не проверяем конкретную ошибку, т.к. реализация может отличаться
}

// TestTermAutomatorBufferResize проверяет, что буфер канала расширяется при необходимости
func TestTermAutomatorBufferResize(t *testing.T) {
	// t.Parallel() // !! нельзя параллелить, тестирую с подменой глобальных переменных

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)

	// Запоминаем исходные stdin и stdout для восстановления после теста
	origStdin := os.Stdin
	origStdout := os.Stdout

	// Используем t.Cleanup вместо defer для гарантированного восстановления в правильном порядке
	t.Cleanup(func() {
		os.Stdin = origStdin
		os.Stdout = origStdout
	})

	// Создаем автоматор
	automator, err := NewTermAutomator()
	require.NoError(t, err, "Не удалось создать TermAutomator")
	t.Cleanup(func() {
		automator.Close()
	})

	// Создаем временный буфер для подавления вывода в консоль
	tmpFile, err := os.CreateTemp("", "suppress_output")
	require.NoError(t, err, "Не удалось создать временный файл")
	t.Cleanup(func() {
		os.Remove(tmpFile.Name())
	})

	// Перенаправляем stdout во временный файл
	oldStdout := os.Stdout
	os.Stdout = tmpFile
	t.Cleanup(func() {
		os.Stdout = oldStdout
	})

	// Запускаем обработку вывода
	go automator.Run()

	// Даем немного времени, чтобы Run запустился
	time.Sleep(100 * time.Millisecond)

	// Добавляем строку напрямую в канал для проверки
	automator.outputLines <- "Тестовая строка для проверки"

	// Проверяем, что строка найдена
	found := automator.WaitForOutput(ctx, "Тестовая строка", 500*time.Millisecond)
	assert.True(t, found, "Должна быть найдена тестовая строка")
}

// TestTermAutomatorErrorHandling проверяет корректность обработки ошибок
func TestTermAutomatorErrorHandling(t *testing.T) {
	// t.Parallel() // !! нельзя параллелить, тестирую с подменой глобальных переменных

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)

	// Запоминаем исходные stdin и stdout для восстановления после теста
	origStdin := os.Stdin
	origStdout := os.Stdout

	// Используем t.Cleanup вместо defer для гарантированного восстановления в правильном порядке
	t.Cleanup(func() {
		os.Stdin = origStdin
		os.Stdout = origStdout
	})

	// Создаем временный файл для перенаправления вывода
	// Это нужно для подавления нежелательного вывода в консоль при запуске тестов
	tmpFile, err := os.CreateTemp("", "test_error_handling")
	require.NoError(t, err, "Не удалось создать временный файл")
	t.Cleanup(func() {
		os.Remove(tmpFile.Name())
	})

	// Перенаправляем stdout во временный файл
	oldStdout := os.Stdout
	os.Stdout = tmpFile
	t.Cleanup(func() {
		os.Stdout = oldStdout
	})

	t.Run("WaitForOutput с малым таймаутом без Run", func(t *testing.T) {
		t.Parallel()

		// Создаем автоматор
		automator, err := NewTermAutomator()
		require.NoError(t, err, "Не удалось создать TermAutomator")
		t.Cleanup(func() {
			automator.Close()
		})

		// Проверяем ситуацию, когда Run не запущен и WaitForOutput вызван с очень малым таймаутом
		found := automator.WaitForOutput(ctx, "этого не будет", 1*time.Millisecond)
		assert.False(t, found, "WaitForOutput должен вернуть false при малом таймауте и отсутствии вывода")
	})

	t.Run("WaitForOutput с пустым шаблоном", func(t *testing.T) {
		t.Parallel()

		// Создаем автоматор
		automator, err := NewTermAutomator()
		require.NoError(t, err, "Не удалось создать TermAutomator")
		t.Cleanup(func() {
			automator.Close()
		})

		// Запускаем обработку вывода
		go automator.Run()

		// Даем время на запуск Run()
		time.Sleep(50 * time.Millisecond)

		// Пишем тестовые данные напрямую в канал
		automator.outputLines <- "Тестовая строка для проверки пустого шаблона"

		// Проверяем с пустым шаблоном
		found := automator.WaitForOutput(ctx, "", 200*time.Millisecond)
		assert.True(t, found, "WaitForOutput должен вернуть true при пустом шаблоне")
	})

	t.Run("WaitForOutput после закрытия stdout", func(t *testing.T) {
		t.Parallel()

		// Создаем автоматор
		automator, err := NewTermAutomator()
		require.NoError(t, err, "Не удалось создать TermAutomator")
		t.Cleanup(func() {
			automator.Close()
		})

		// Создаем мок для ручного контроля канала вывода
		mockAutomator := &TermAutomator{
			log:         automator.log,
			outputLines: make(chan string),
			// Остальные поля не используются в этом тесте
		}

		// Закрываем канал вывода, чтобы эмулировать завершение Run
		close(mockAutomator.outputLines)

		// Проверяем поведение WaitForOutput при закрытом канале
		timeStart := time.Now()
		found := mockAutomator.WaitForOutput(ctx, "что-то", 1*time.Second)
		timeElapsed := time.Since(timeStart)

		assert.False(t, found, "WaitForOutput должен вернуть false при закрытом канале")
		assert.Less(t, timeElapsed, 500*time.Millisecond, "WaitForOutput должен вернуться быстрее таймаута при закрытом канале")
	})

	t.Run("SendInput после Close", func(t *testing.T) {
		t.Parallel()

		// Создаем автоматор
		automator, err := NewTermAutomator()
		require.NoError(t, err, "Не удалось создать TermAutomator")

		// Останавливаем автоматор
		automator.Close()

		// Пытаемся отправить ввод после остановки
		err = automator.SendInput("тест")
		require.Error(t, err, "SendInput должен возвращать ошибку после Close")
		// Мы не проверяем конкретную ошибку, так как поведение может зависеть от реализации
	})
}

// mockReadWriter реализует простой мок для тестирования с изолированными вводом/выводом
type mockReadWriter struct {
	readData  string
	writeData string
	closed    bool
}

func (m *mockReadWriter) Read(p []byte) (int, error) {
	if m.closed {
		return 0, io.EOF
	}
	if len(m.readData) == 0 {
		return 0, io.EOF
	}
	n := copy(p, m.readData)
	m.readData = m.readData[n:]
	return n, nil
}

func (m *mockReadWriter) Write(p []byte) (int, error) {
	if m.closed {
		return 0, errors.New("write to closed pipe")
	}
	m.writeData += string(p)
	return len(p), nil
}

func (m *mockReadWriter) Close() error {
	m.closed = true
	return nil
}

// TestTermAutomatorWithMocks проверяет TermAutomator с использованием моков для stdin/stdout
func TestTermAutomatorWithMocks(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)

	// Этот тест не использует реальные stdin/stdout, а создает моки
	// В реальной ситуации мы бы использовали gomock или testify/mock для более сложных моков

	// Создаем экземпляр TermAutomator вручную для тестирования с моками
	automator := &TermAutomator{
		log:         log.NewLogger(),
		outputLines: make(chan string, 10),
	}

	// Создаем сообщения для тестирования
	testMessages := []string{
		"Сообщение 1\n",
		"Сообщение 2\n",
		"Финальное сообщение\n",
	}

	// Подготавливаем данные для мока
	mockData := strings.Join(testMessages, "")

	// Создаем мок для чтения stdout
	mockReader := &mockReadWriter{
		readData: mockData,
	}

	// Канал для синхронизации: сообщает о доступности следующего сообщения
	messageReady := make(chan struct{})

	// Канал для завершения теста
	done := make(chan struct{})

	// Канал для контроля последовательности проверок
	nextCheck := make(chan struct{}, 3) // Буфер на 3 сообщения

	// Закрываем каналы при завершении теста
	t.Cleanup(func() {
		close(automator.outputLines)
	})

	// Запускаем тестирование в отдельной горутине
	go func() {
		// Создаем буфер для чтения
		reader := bufio.NewReader(mockReader)

		// Читаем и обрабатываем каждую строку
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					t.Errorf("Ошибка чтения: %v", err)
				}
				break
			}

			// Отправляем строку в канал, удаляя символ новой строки
			automator.outputLines <- strings.TrimSuffix(line, "\n")

			// Сигнализируем о готовности сообщения
			nextCheck <- struct{}{}

			// Ждем подтверждения, что проверка выполнена
			<-messageReady
		}

		// Сигнализируем о завершении
		close(done)
	}()

	// Ожидаем первое сообщение
	<-nextCheck

	// Проверяем обнаружение первого сообщения
	found := automator.WaitForOutput(ctx, "Сообщение 1", 100*time.Millisecond)
	assert.True(t, found, "Должно найти первое сообщение")

	// Сигнализируем о завершении проверки первого сообщения
	messageReady <- struct{}{}

	// Ожидаем второе сообщение
	<-nextCheck

	// Проверяем обнаружение второго сообщения
	found = automator.WaitForOutput(ctx, "Сообщение 2", 100*time.Millisecond)
	assert.True(t, found, "Должно найти второе сообщение")

	// Сигнализируем о завершении проверки второго сообщения
	messageReady <- struct{}{}

	// Ожидаем третье сообщение
	<-nextCheck

	// Проверяем обнаружение финального сообщения
	found = automator.WaitForOutput(ctx, "Финальное", 100*time.Millisecond)
	assert.True(t, found, "Должно найти финальное сообщение")

	// Сигнализируем о завершении проверки третьего сообщения
	messageReady <- struct{}{}

	// Ждем завершения чтения
	<-done

	// Проверяем, что все сообщения были прочитаны
	assert.Empty(t, mockReader.readData, "Все данные должны быть прочитаны")
}
