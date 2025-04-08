package entity

import "time"

// Report представляет отчет о работе системы
type Report struct {
	// ID уникальный идентификатор отчета
	ID string
	// Period период времени, за который составляется отчет
	Period ReportPeriod
	// StartTime время начала периода отчета
	StartTime time.Time
	// EndTime время окончания периода отчета
	EndTime time.Time
	// Template шаблон текста отчета (с поддержкой разметки)
	Template string
	// For список идентификаторов чатов, для которых генерируется отчет
	For []int64
	// Statistics статистика отчета
	Statistics ReportStatistics
	// Status статус отчета
	Status ReportStatus
}

// ReportPeriod представляет период отчета
type ReportPeriod string

const (
	// ReportPeriodDay отчет за день
	ReportPeriodDay ReportPeriod = "day"
	// ReportPeriodWeek отчет за неделю
	ReportPeriodWeek ReportPeriod = "week"
	// ReportPeriodMonth отчет за месяц
	ReportPeriodMonth ReportPeriod = "month"
	// ReportPeriodCustom отчет за произвольный период
	ReportPeriodCustom ReportPeriod = "custom"
)

// ReportStatistics представляет статистику в отчете
type ReportStatistics struct {
	// TotalMessages общее количество сообщений, обработанных системой
	TotalMessages int
	// ForwardedMessages количество пересланных сообщений
	ForwardedMessages int
	// FilteredMessages количество отфильтрованных сообщений
	FilteredMessages int
	// BySource статистика по источникам сообщений
	BySource map[int64]SourceStatistics
	// ByDestination статистика по назначениям сообщений
	ByDestination map[int64]DestinationStatistics
}

// SourceStatistics представляет статистику по источнику сообщений
type SourceStatistics struct {
	// TotalMessages общее количество сообщений из этого источника
	TotalMessages int
	// ForwardedMessages количество пересланных сообщений из этого источника
	ForwardedMessages int
}

// DestinationStatistics представляет статистику по назначению сообщений
type DestinationStatistics struct {
	// ReceivedMessages количество полученных сообщений этим назначением
	ReceivedMessages int
	// BySource статистика по источникам сообщений для этого назначения
	BySource map[int64]int
}

// ReportStatus представляет статус отчета
type ReportStatus string

const (
	// ReportStatusPending отчет ожидает генерации
	ReportStatusPending ReportStatus = "pending"
	// ReportStatusGenerated отчет сгенерирован
	ReportStatusGenerated ReportStatus = "generated"
	// ReportStatusSent отчет отправлен
	ReportStatusSent ReportStatus = "sent"
	// ReportStatusFailed ошибка при генерации отчета
	ReportStatusFailed ReportStatus = "failed"
)

// NewReport создает новый экземпляр отчета
func NewReport(id string, period ReportPeriod, template string, for_ []int64) *Report {
	now := time.Now()
	startTime, endTime := calculateReportTimeRange(period, now)

	return &Report{
		ID:        id,
		Period:    period,
		Template:  template,
		For:       for_,
		StartTime: startTime,
		EndTime:   endTime,
		Status:    ReportStatusPending,
		Statistics: ReportStatistics{
			BySource:      make(map[int64]SourceStatistics),
			ByDestination: make(map[int64]DestinationStatistics),
		},
	}
}

// SetCustomTimeRange устанавливает произвольный временной диапазон для отчета
func (r *Report) SetCustomTimeRange(startTime, endTime time.Time) {
	r.Period = ReportPeriodCustom
	r.StartTime = startTime
	r.EndTime = endTime
}

// MarkGenerated помечает отчет как сгенерированный
func (r *Report) MarkGenerated() {
	r.Status = ReportStatusGenerated
}

// MarkSent помечает отчет как отправленный
func (r *Report) MarkSent() {
	r.Status = ReportStatusSent
}

// MarkFailed помечает отчет как неудачный
func (r *Report) MarkFailed() {
	r.Status = ReportStatusFailed
}

// calculateReportTimeRange рассчитывает временной диапазон для отчета
func calculateReportTimeRange(period ReportPeriod, now time.Time) (time.Time, time.Time) {
	switch period {
	case ReportPeriodDay:
		// Отчет за день: от начала до конца текущего дня
		startTime := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		endTime := startTime.Add(24 * time.Hour).Add(-time.Second)
		return startTime, endTime
	case ReportPeriodWeek:
		// Отчет за неделю: от начала текущей недели до конца текущей недели
		// Считаем, что неделя начинается с понедельника (1) и заканчивается воскресеньем (7)
		daysFromMonday := int(now.Weekday()) - 1
		if daysFromMonday < 0 {
			daysFromMonday = 6 // Если сегодня воскресенье (0), то это 6 дней от понедельника
		}
		startTime := time.Date(now.Year(), now.Month(), now.Day()-daysFromMonday, 0, 0, 0, 0, now.Location())
		endTime := startTime.Add(7 * 24 * time.Hour).Add(-time.Second)
		return startTime, endTime
	case ReportPeriodMonth:
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
