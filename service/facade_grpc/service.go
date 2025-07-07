package facade_grpc

import (
	"github.com/zelenin/go-tdlib/client"

	"github.com/comerc/budva43/app/dto/grpc"
	"github.com/comerc/budva43/app/log"
)

//go:generate mockery --name=telegramRepo --exported
type telegramRepo interface {
	GetClientDone() <-chan any
	// tdlibClient methods
	GetMessage(*client.GetMessageRequest) (*client.Message, error)
	SendMessage(*client.SendMessageRequest) (*client.Message, error)
	SendMessageAlbum(*client.SendMessageAlbumRequest) (*client.Messages, error)
	EditMessageText(*client.EditMessageTextRequest) (*client.Message, error)
	EditMessageCaption(*client.EditMessageCaptionRequest) (*client.Message, error)
	DeleteMessages(*client.DeleteMessagesRequest) (*client.Ok, error)
	GetMessages(*client.GetMessagesRequest) (*client.Messages, error)
}

type Service struct {
	log *log.Logger
	//
	telegramRepo telegramRepo
}

func New(telegramRepo telegramRepo) *Service {
	return &Service{
		log: log.NewLogger(),
		//
		telegramRepo: telegramRepo,
	}
}

func (s *Service) Start() error {
	return nil
}

func (s *Service) Close() error {
	return nil
}

func (s *Service) GetClientDone() <-chan any {
	return s.telegramRepo.GetClientDone()
}

func (s *Service) GetMessages(chatId int64) ([]*grpc.Message, error) {
	return nil, nil
}

func (s *Service) CreateMessage(message *grpc.NewMessage) (*grpc.Message, error) {
	return nil, nil
}

func (s *Service) GetMessage(messageId int64) (*grpc.Message, error) {
	return nil, nil
}

func (s *Service) UpdateMessage(message *grpc.Message) (*grpc.Message, error) {
	return nil, nil
}

func (s *Service) DeleteMessage(message *grpc.Message) (*grpc.Message, error) {
	return nil, nil
}
