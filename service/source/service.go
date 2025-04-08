package source

import (
	"github.com/comerc/budva43/entity"
)

// SourceService предоставляет методы для работы с источниками сообщений
type SourceService struct {
	// Здесь могут быть зависимости, например, репозитории
}

// NewSourceService создает новый экземпляр сервиса для работы с источниками сообщений
func NewSourceService() *SourceService {
	return &SourceService{}
}

// ShouldAddSign проверяет, нужно ли добавлять подпись к сообщению для указанного чата
func (s *SourceService) ShouldAddSign(source *entity.Source, chatID int64) bool {
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
func (s *SourceService) ShouldAddLink(source *entity.Source, chatID int64) bool {
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
