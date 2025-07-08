package facade_gql

import (
	"github.com/zelenin/go-tdlib/client"

	"github.com/comerc/budva43/app/dto/gql/dto"
	"github.com/comerc/budva43/app/log"
	"github.com/comerc/budva43/app/util"
)

//go:generate mockery --name=telegramRepo --exported
type telegramRepo interface {
	GetClientDone() <-chan any
	// tdlibClient methods
	GetOption(*client.GetOptionRequest) (client.OptionValue, error)
	GetMe() (*client.User, error)
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

// GetStatus возвращает статус авторизации
func (s *Service) GetStatus() (*dto.Status, error) {
	var err error
	defer s.log.ErrorOrDebug(&err, "")

	// TODO: отображать аргументы структурированной ошибки в GraphQL

	var versionOption client.OptionValue
	versionOption, err = s.telegramRepo.GetOption(&client.GetOptionRequest{
		Name: "version",
	})
	if err != nil {
		return nil, err
	}
	version := versionOption.(*client.OptionValueString).Value

	var me *client.User
	me, err = s.telegramRepo.GetMe()
	if err != nil {
		return nil, err
	}

	return &dto.Status{
		ReleaseVersion: util.GetReleaseVersion(),
		TdlibVersion:   version,
		UserId:         me.Id,
	}, nil
}
