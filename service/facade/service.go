package facade

import (
	"context"

	"github.com/comerc/budva43/app/log"
)

//go:generate mockery --name=telegramRepo --exported
type telegramRepo interface {
	GetClientDone() <-chan any
}

//go:generate mockery --name=loaderService --exported
type loaderService interface {
	Run()
}

// Service предоставляет функциональность фасада
type Service struct {
	log *log.Logger
	//
	telegramRepo  telegramRepo
	loaderService loaderService
}

// New создает новый экземпляр сервиса фасада
func New(
	telegramRepo telegramRepo,
	loaderService loaderService,
) *Service {
	return &Service{
		log: log.NewLogger(),
		//
		telegramRepo:  telegramRepo,
		loaderService: loaderService,
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
		s.loaderService.Run()
	}
}
