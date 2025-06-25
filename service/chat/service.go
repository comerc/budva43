package chat

import (
	"github.com/comerc/budva43/app/log"
	"github.com/zelenin/go-tdlib/client"
)

//go:generate mockery --name=telegramRepo --exported
type telegramRepo interface {
	GetChat(*client.GetChatRequest) (*client.Chat, error)
}

type Service struct {
	log *log.Logger
	//
	telegramRepo telegramRepo
}

func New(telegramRepo telegramRepo) *Service {
	return &Service{
		log: log.NewLogger("service.chat"),
		//
		telegramRepo: telegramRepo,
	}
}

// IsBasicGroup проверяет, является ли чат основным чатом
func (s *Service) IsBasicGroup(chatId int64) (bool, error) {
	chat, err := s.telegramRepo.GetChat(&client.GetChatRequest{
		ChatId: chatId,
	})
	if err != nil {
		return false, err
	}
	_, ok := chat.Type.(*client.ChatTypeBasicGroup)
	return ok, nil
}
