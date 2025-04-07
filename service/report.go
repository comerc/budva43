package service

import (
	"github.com/comerc/budva43/entity"
)

// ReportService предоставляет методы для работы с отчетами
type ReportService struct {
	// Здесь могут быть зависимости, например, репозитории
}

// NewReportService создает новый экземпляр сервиса для работы с отчетами
func NewReportService() *ReportService {
	return &ReportService{}
}

// AddSourceStatistics добавляет статистику по источнику сообщений
func (s *ReportService) AddSourceStatistics(report *entity.Report, sourceID int64, total, forwarded int) {
	report.Statistics.BySource[sourceID] = entity.SourceStatistics{
		TotalMessages:     total,
		ForwardedMessages: forwarded,
	}
	report.Statistics.TotalMessages += total
	report.Statistics.ForwardedMessages += forwarded
}

// AddDestinationStatistics добавляет статистику по назначению сообщений
func (s *ReportService) AddDestinationStatistics(report *entity.Report, destinationID int64, received int) {
	// Получаем текущую статистику или создаем новую
	stats, exists := report.Statistics.ByDestination[destinationID]
	if !exists {
		stats = entity.DestinationStatistics{
			BySource: make(map[int64]int),
		}
	}

	// Обновляем количество полученных сообщений
	stats.ReceivedMessages += received

	// Сохраняем обновленную статистику обратно в карту
	report.Statistics.ByDestination[destinationID] = stats
}

// AddDestinationSourceStatistics добавляет статистику по источнику для назначения
func (s *ReportService) AddDestinationSourceStatistics(report *entity.Report, destinationID, sourceID int64, count int) {
	// Получаем текущую статистику или создаем новую
	stats, exists := report.Statistics.ByDestination[destinationID]
	if !exists {
		stats = entity.DestinationStatistics{
			BySource: make(map[int64]int),
		}
	}

	// Обновляем количество сообщений от источника
	currentCount, _ := stats.BySource[sourceID]
	stats.BySource[sourceID] = currentCount + count

	// Сохраняем обновленную статистику обратно в карту
	report.Statistics.ByDestination[destinationID] = stats
}
