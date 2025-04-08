package report

import (
	"errors"
	"time"

	"github.com/comerc/budva43/entity"
)

// reportService определяет интерфейс сервиса отчетов, необходимый контроллеру
type reportService interface {
	GenerateActivityReport(startDate, endDate time.Time) (*entity.ActivityReport, error)
	GenerateForwardingReport(startDate, endDate time.Time) (*entity.ForwardingReport, error)
	GenerateErrorReport(startDate, endDate time.Time) (*entity.ErrorReport, error)
}

// storageRepository определяет интерфейс репозитория хранилища, необходимый контроллеру
type storageRepository interface {
	Get(key []byte) ([]byte, error)
	Set(key, value []byte) error
}

// Controller представляет контроллер для работы с отчетами
type Controller struct {
	reportService     reportService
	storageRepository storageRepository
}

// NewController создает новый экземпляр контроллера отчетов
func NewController(
	reportService reportService,
	storageRepository storageRepository,
) *Controller {
	return &Controller{
		reportService:     reportService,
		storageRepository: storageRepository,
	}
}

// GenerateActivityReport генерирует отчет об активности за период
func (c *Controller) GenerateActivityReport(startDate, endDate time.Time) (*entity.ActivityReport, error) {
	return c.reportService.GenerateActivityReport(startDate, endDate)
}

// GenerateForwardingReport генерирует отчет о пересылке сообщений за период
func (c *Controller) GenerateForwardingReport(startDate, endDate time.Time) (*entity.ForwardingReport, error) {
	return c.reportService.GenerateForwardingReport(startDate, endDate)
}

// GenerateErrorReport генерирует отчет об ошибках за период
func (c *Controller) GenerateErrorReport(startDate, endDate time.Time) (*entity.ErrorReport, error) {
	return c.reportService.GenerateErrorReport(startDate, endDate)
}

// SaveReport сохраняет отчет в хранилище
func (c *Controller) SaveReport(report entity.Report, key string) error {
	// Здесь должна быть реализация сериализации отчета и сохранения в хранилище
	// Это заглушка для примера
	return nil
}

// GetReport получает отчет из хранилища
func (c *Controller) GetReport(key string, reportType string) (entity.Report, error) {
	// Здесь должна быть реализация получения и десериализации отчета из хранилища
	// Это заглушка для примера

	switch reportType {
	case "activity":
		return &entity.ActivityReport{}, nil
	case "forwarding":
		return &entity.ForwardingReport{}, nil
	case "error":
		return &entity.ErrorReport{}, nil
	default:
		// Возвращаем ошибку для неизвестного типа отчета
		return nil, errors.New("unknown report type")
	}
}
