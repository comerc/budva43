package user

import (
	"time"

	"github.com/zelenin/go-tdlib/client"
)

// Service предоставляет методы для работы с пользователями
type Service struct {
	// Здесь могут быть зависимости, например, репозитории
}

// New создает новый экземпляр сервиса для работы с пользователями
func New() *Service {
	return &Service{}
}

// GetFullName возвращает полное имя пользователя
func (s *Service) GetFullName(user *client.User) string {
	if user == nil {
		return ""
	}

	if user.LastName == "" {
		return user.FirstName
	}

	return user.FirstName + " " + user.LastName
}

// IsBot проверяет, является ли пользователь ботом
func (s *Service) IsBot(user *client.User) bool {
	if user == nil || user.Type == nil {
		return false
	}

	_, ok := user.Type.(*client.UserTypeBot)
	return ok
}

// GetStatusText возвращает текстовое представление статуса пользователя
func (s *Service) GetStatusText(user *client.User) string {
	if user == nil || user.Status == nil {
		return "unknown"
	}

	switch status := user.Status.(type) {
	case *client.UserStatusOnline:
		return "online"
	case *client.UserStatusOffline:
		// Время последнего посещения
		lastOnlineTime := time.Unix(int64(status.WasOnline), 0)
		now := time.Now()

		diff := now.Sub(lastOnlineTime)
		if diff < 24*time.Hour {
			return "last seen today"
		} else if diff < 48*time.Hour {
			return "last seen yesterday"
		} else if diff < 7*24*time.Hour {
			return "last seen this week"
		} else if diff < 30*24*time.Hour {
			return "last seen this month"
		}
		return "offline"
	case *client.UserStatusRecently:
		return "last seen recently"
	case *client.UserStatusLastWeek:
		return "last seen this week"
	case *client.UserStatusLastMonth:
		return "last seen this month"
	default:
		return "last seen a long time ago"
	}
}
