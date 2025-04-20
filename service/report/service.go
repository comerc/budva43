package report

import (
	"log/slog"
	"time"

	"github.com/comerc/budva43/entity"
)

// Service предоставляет методы для работы с отчетами
type Service struct {
	log *slog.Logger
	//
	// Здесь могут быть зависимости, например, репозитории
}

// New создает новый экземпляр сервиса для работы с отчетами
func New() *Service {
	return &Service{
		log: slog.With("module", "service.report"),
		//
	}
}

// calculateReportTimeRange рассчитывает временной диапазон для отчета
func calculateReportTimeRange(period entity.ReportPeriod, now time.Time) (time.Time, time.Time) {
	switch period {
	case entity.ReportPeriodDay:
		// Отчет за день: от начала до конца текущего дня
		startTime := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		endTime := startTime.Add(24 * time.Hour).Add(-time.Second)
		return startTime, endTime
	case entity.ReportPeriodWeek:
		// Отчет за неделю: от начала текущей недели до конца текущей недели
		// Считаем, что неделя начинается с понедельника (1) и заканчивается воскресеньем (7)
		daysFromMonday := int(now.Weekday()) - 1
		if daysFromMonday < 0 {
			daysFromMonday = 6 // Если сегодня воскресенье (0), то это 6 дней от понедельника
		}
		startTime := time.Date(now.Year(), now.Month(), now.Day()-daysFromMonday, 0, 0, 0, 0, now.Location())
		endTime := startTime.Add(7 * 24 * time.Hour).Add(-time.Second)
		return startTime, endTime
	case entity.ReportPeriodMonth:
		// Отчет за месяц: от начала до конца текущего месяца
		startTime := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		endTime := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location()).Add(-time.Second)
		return startTime, endTime
	default:
		// По умолчанию - за день
		startTime := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		endTime := startTime.Add(24 * time.Hour).Add(-time.Second)
		return startTime, endTime
	}
}

// GenerateActivityReport генерирует отчет об активности за указанный период
func (s *Service) GenerateActivityReport(startDate, endDate time.Time) (*entity.ActivityReport, error) {
	// Пока просто заглушка
	id := time.Now().Format("activity_20060102_150405")

	now := time.Now()
	startTime, endTime := calculateReportTimeRange(entity.ReportPeriodCustom, now)

	report := &entity.ActivityReport{
		BaseReport: entity.BaseReport{
			ID:        id,
			Period:    entity.ReportPeriodCustom,
			Template:  "Отчет об активности",
			For:       []int64{},
			StartTime: startTime,
			EndTime:   endTime,
			Status:    entity.ReportStatusPending,
		},
		Statistics: entity.ActivityStatistics{
			UserActivity: make(map[int64]int),
			ChatActivity: make(map[int64]int),
		},
	}

	report.SetCustomTimeRange(startDate, endDate)
	report.MarkGenerated()

	return report, nil
}

// GenerateForwardingReport генерирует отчет о пересылке сообщений за указанный период
func (s *Service) GenerateForwardingReport(startDate, endDate time.Time) (*entity.ForwardingReport, error) {
	// Пока просто заглушка
	id := time.Now().Format("forwarding_20060102_150405")

	now := time.Now()
	startTime, endTime := calculateReportTimeRange(entity.ReportPeriodCustom, now)

	report := &entity.ForwardingReport{
		BaseReport: entity.BaseReport{
			ID:        id,
			Period:    entity.ReportPeriodCustom,
			Template:  "Отчет о пересылке",
			For:       []int64{},
			StartTime: startTime,
			EndTime:   endTime,
			Status:    entity.ReportStatusPending,
		},
		Statistics: entity.ForwardingStatistics{
			BySource:      make(map[int64]entity.SourceStatistics),
			ByDestination: make(map[int64]entity.DestinationStatistics),
		},
	}

	report.SetCustomTimeRange(startDate, endDate)
	report.MarkGenerated()

	return report, nil
}

// GenerateErrorReport генерирует отчет об ошибках за указанный период
func (s *Service) GenerateErrorReport(startDate, endDate time.Time) (*entity.ErrorReport, error) {
	// Пока просто заглушка
	id := time.Now().Format("error_20060102_150405")

	now := time.Now()
	startTime, endTime := calculateReportTimeRange(entity.ReportPeriodCustom, now)

	report := &entity.ErrorReport{
		BaseReport: entity.BaseReport{
			ID:        id,
			Period:    entity.ReportPeriodCustom,
			Template:  "Отчет об ошибках",
			For:       []int64{},
			StartTime: startTime,
			EndTime:   endTime,
			Status:    entity.ReportStatusPending,
		},
		Errors: make([]entity.SystemError, 0),
	}

	report.SetCustomTimeRange(startDate, endDate)
	report.MarkGenerated()

	return report, nil
}

// AddSourceStatistics добавляет статистику по источнику сообщений в отчет о пересылке
func (s *Service) AddSourceStatistics(report *entity.ForwardingReport, sourceID int64, total, forwarded int) {
	report.Statistics.BySource[sourceID] = entity.SourceStatistics{
		TotalMessages:     total,
		ForwardedMessages: forwarded,
	}
	report.Statistics.TotalMessages += total
	report.Statistics.ForwardedMessages += forwarded
}

// AddDestinationStatistics добавляет статистику по назначению сообщений в отчет о пересылке
func (s *Service) AddDestinationStatistics(report *entity.ForwardingReport, destinationID int64, received int) {
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

// AddDestinationSourceStatistics добавляет статистику по источнику для назначения в отчет о пересылке
func (s *Service) AddDestinationSourceStatistics(report *entity.ForwardingReport, destinationID, sourceID int64, count int) {
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

// AddUserActivity добавляет статистику активности пользователя в отчет об активности
func (s *Service) AddUserActivity(report *entity.ActivityReport, userID int64, messageCount int) {
	currentCount, exists := report.Statistics.UserActivity[userID]
	if exists {
		report.Statistics.UserActivity[userID] = currentCount + messageCount
	} else {
		report.Statistics.UserActivity[userID] = messageCount
	}

	report.Statistics.TotalMessages += messageCount
	report.Statistics.ActiveUsers = len(report.Statistics.UserActivity)
}

// AddChatActivity добавляет статистику активности чата в отчет об активности
func (s *Service) AddChatActivity(report *entity.ActivityReport, chatID int64, messageCount int) {
	currentCount, exists := report.Statistics.ChatActivity[chatID]
	if exists {
		report.Statistics.ChatActivity[chatID] = currentCount + messageCount
	} else {
		report.Statistics.ChatActivity[chatID] = messageCount
	}

	report.Statistics.ActiveChats = len(report.Statistics.ChatActivity)
}

// AddErrorRecord добавляет запись об ошибке в отчет об ошибках
func (s *Service) AddErrorRecord(report *entity.ErrorReport, timestamp time.Time, code, message, component string, severity entity.ErrorSeverity) {
	errorRecord := entity.SystemError{
		Timestamp: timestamp,
		Code:      code,
		Message:   message,
		Component: component,
		Severity:  severity,
	}

	report.Errors = append(report.Errors, errorRecord)
}
