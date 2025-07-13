package facade

import (
	"context"

	"github.com/comerc/budva43/app/log"
)

//go:generate mockery --name=telegramRepo --exported
type telegramRepo interface {
	GetClientDone() <-chan any
}

//go:generate mockery --name=engineLoaderService --exported
type engineLoaderService interface {
	LoadConfig()
}

// Service предоставляет функциональность фасада
type Service struct {
	log *log.Logger
	//
	telegramRepo        telegramRepo
	engineLoaderService engineLoaderService
}

// New создает новый экземпляр сервиса фасада
func New(
	telegramRepo telegramRepo,
	engineLoaderService engineLoaderService,
) *Service {
	return &Service{
		log: log.NewLogger(),
		//
		telegramRepo:        telegramRepo,
		engineLoaderService: engineLoaderService,
	}
}

// StartContext запускает сервис
func (s *Service) StartContext(ctx context.Context) error {

	go s.run(ctx)

	return nil
}

// Close останавливает сервис
func (s *Service) Close() error {
	return nil
}

func (s *Service) run(ctx context.Context) {
	// Ждём авторизации клиента
	select {
	case <-ctx.Done():
		return
	case <-s.telegramRepo.GetClientDone():
		s.engineLoaderService.LoadConfig()
	}
}
