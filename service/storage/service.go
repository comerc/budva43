package storage

import (
	"log/slog"
	"time"

	"github.com/comerc/budva43/entity"
)

//go:generate mockery --name=storageRepo --exported
type storageRepo interface {
	Get(key []byte) ([]byte, error)
	Set(key, value []byte) error
	SetWithTTL(key, value []byte, ttl time.Duration) error
	Delete(key []byte) error
}

// Service предоставляет методы для работы с хранилищем
type Service struct {
	log *slog.Logger
	//
	repo storageRepo
}

// New создает новый экземпляр сервиса для работы с хранилищем
func New(repo storageRepo) *Service {
	return &Service{
		log: slog.With("module", "service.storage"),
		//
		repo: repo,
	}
}

func (s *Service) GetCopiedMessageIDs(fromChatMessageID string) ([]string, error) {
	// TODO: реализовать
	return nil, nil
}

func (s *Service) SetNewMessageID(chatID, tmpMessageID, newMessageID int64) error {
	// TODO: реализовать
	return nil
}

func (s *Service) SetTmpMessageID(chatID, newMessageID, tmpMessageID int64) error {
	// TODO: реализовать
	return nil
}

func (s *Service) GetRuleByID(ruleID string) (entity.ForwardRule, bool) {
	// TODO: реализовать
	return entity.ForwardRule{}, false
}
