package source

import (
	"log/slog"

	"github.com/comerc/budva43/entity"
)

// Service предоставляет методы для работы с источниками сообщений
type Service struct {
	log *slog.Logger
	//
	// Здесь могут быть зависимости, например, репозитории
}

// New создает новый экземпляр сервиса для работы с источниками сообщений
func New() *Service {
	return &Service{
		log: slog.With("module", "service.source"),
		//
	}
}

// ShouldAddSign проверяет, нужно ли добавлять подпись к сообщению для указанного чата
func (s *Service) ShouldAddSign(source *entity.Source, chatID int64) bool {
	if source.Sign == nil {
		return false
	}

	for _, id := range source.Sign.For {
		if id == chatID {
			return true
		}
	}

	return false
}

// ShouldAddLink проверяет, нужно ли добавлять ссылку к сообщению для указанного чата
func (s *Service) ShouldAddLink(source *entity.Source, chatID int64) bool {
	if source.Link == nil {
		return false
	}

	for _, id := range source.Link.For {
		if id == chatID {
			return true
		}
	}

	return false
}
