package notification

import (
	"sync"
	"time"
)

// Notifier интерфейс для отправки уведомлений
type Notifier interface {
	SendNotification(chatID int64, message string) error
}

// Service предоставляет методы для работы с уведомлениями
type Service struct {
	notifier      Notifier
	notifyHistory map[string]time.Time
	mutex         sync.RWMutex
}

// New создает новый экземпляр сервиса для уведомлений
func New(notifier Notifier) *Service {
	return &Service{
		notifier:      notifier,
		notifyHistory: make(map[string]time.Time),
	}
}

// SendNotification отправляет уведомление в чат
func (s *Service) SendNotification(chatID int64, message string) error {
	if s.notifier == nil {
		return nil // Тихое игнорирование, если нет notifier
	}
	return s.notifier.SendNotification(chatID, message)
}

// SendThrottledNotification отправляет уведомление не чаще указанного интервала
func (s *Service) SendThrottledNotification(key string, chatID int64, message string, interval time.Duration) error {
	s.mutex.RLock()
	lastTime, exists := s.notifyHistory[key]
	s.mutex.RUnlock()

	now := time.Now()

	// Если уведомление уже отправлялось и не прошло достаточно времени
	if exists && now.Sub(lastTime) < interval {
		return nil
	}

	// Отправляем уведомление
	err := s.SendNotification(chatID, message)
	if err != nil {
		return err
	}

	// Обновляем историю отправки
	s.mutex.Lock()
	s.notifyHistory[key] = now
	s.mutex.Unlock()

	return nil
}

// SendStatusNotification отправляет уведомление о статусе операции
func (s *Service) SendStatusNotification(chatID int64, operation string, success bool, details string) error {
	var message string
	if success {
		message = "✅ " + operation + " успешно выполнена"
	} else {
		message = "❌ " + operation + " завершилась с ошибкой"
	}

	if details != "" {
		message += "\nПодробности: " + details
	}

	return s.SendNotification(chatID, message)
}

// SendErrorNotification отправляет уведомление об ошибке
func (s *Service) SendErrorNotification(chatID int64, errorType string, details string) error {
	message := "❌ Ошибка: " + errorType
	if details != "" {
		message += "\nПодробности: " + details
	}

	return s.SendNotification(chatID, message)
}

// SendBatchNotifications отправляет набор уведомлений
func (s *Service) SendBatchNotifications(notifications map[int64]string) (map[int64]error, error) {
	results := make(map[int64]error)

	for chatID, message := range notifications {
		err := s.SendNotification(chatID, message)
		results[chatID] = err
	}

	return results, nil
}

// ClearNotificationHistory очищает историю отправленных уведомлений
func (s *Service) ClearNotificationHistory() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.notifyHistory = make(map[string]time.Time)
}

// SendProgressNotification отправляет уведомление о прогрессе операции
func (s *Service) SendProgressNotification(chatID int64, operation string, progress int, total int) error {
	message := "⏳ " + operation + ": " +
		"выполнено " + string(rune(progress)) + " из " + string(rune(total))

	return s.SendNotification(chatID, message)
}
