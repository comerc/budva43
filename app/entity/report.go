package entity

import "time"

// Report интерфейс для всех типов отчетов
type Report interface {
	GetId() string
	GetStartTime() time.Time
	GetEndTime() time.Time
	GetStatus() ReportStatus
}

// BaseReport базовая структура для всех типов отчетов
type BaseReport struct {
	// Id уникальный идентификатор отчета
	Id string
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
	// Status статус отчета
	Status ReportStatus
}

// GetId возвращает идентификатор отчета
func (r *BaseReport) GetId() string {
	return r.Id
}

// GetStartTime возвращает время начала периода отчета
func (r *BaseReport) GetStartTime() time.Time {
	return r.StartTime
}

// GetEndTime возвращает время окончания периода отчета
func (r *BaseReport) GetEndTime() time.Time {
	return r.EndTime
}

// GetStatus возвращает статус отчета
func (r *BaseReport) GetStatus() ReportStatus {
	return r.Status
}

// ActivityReport отчет об активности системы
type ActivityReport struct {
	BaseReport
	// Statistics статистика активности
	Statistics ActivityStatistics
}

// ActivityStatistics статистика активности системы
type ActivityStatistics struct {
	// TotalMessages общее количество сообщений, обработанных системой
	TotalMessages int
	// ActiveUsers количество активных пользователей
	ActiveUsers int
	// ActiveChats количество активных чатов
	ActiveChats int
	// UserActivity статистика активности пользователей
	UserActivity map[int64]int // пользователь -> количество сообщений
	// ChatActivity статистика активности чатов
	ChatActivity map[int64]int // чат -> количество сообщений
}

// ForwardingReport отчет о пересылке сообщений
type ForwardingReport struct {
	BaseReport
	// Statistics статистика пересылки
	Statistics ForwardingStatistics
}

// ForwardingStatistics статистика пересылки сообщений
type ForwardingStatistics struct {
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

// ErrorReport отчет об ошибках системы
type ErrorReport struct {
	BaseReport
	// Errors список ошибок
	Errors []SystemError
}

// SystemError представляет ошибку в системе
type SystemError struct {
	// Timestamp время возникновения ошибки
	Timestamp time.Time
	// Code код ошибки
	Code string
	// Message сообщение об ошибке
	Message string
	// Component компонент, в котором произошла ошибка
	Component string
	// Severity серьезность ошибки
	Severity ErrorSeverity
}

// ErrorSeverity уровень серьезности ошибки
type ErrorSeverity string

const (
	// ErrorSeverityInfo информационное сообщение
	ErrorSeverityInfo ErrorSeverity = "info"
	// ErrorSeverityWarning предупреждение
	ErrorSeverityWarning ErrorSeverity = "warning"
	// ErrorSeverityError ошибка
	ErrorSeverityError ErrorSeverity = "error"
	// ErrorSeverityCritical критическая ошибка
	ErrorSeverityCritical ErrorSeverity = "critical"
)

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

// SetCustomTimeRange устанавливает произвольный временной диапазон для отчета
func (r *BaseReport) SetCustomTimeRange(startTime, endTime time.Time) {
	r.Period = ReportPeriodCustom
	r.StartTime = startTime
	r.EndTime = endTime
}

// MarkGenerated помечает отчет как сгенерированный
func (r *BaseReport) MarkGenerated() {
	r.Status = ReportStatusGenerated
}

// MarkSent помечает отчет как отправленный
func (r *BaseReport) MarkSent() {
	r.Status = ReportStatusSent
}

// MarkFailed помечает отчет как неудачный
func (r *BaseReport) MarkFailed() {
	r.Status = ReportStatusFailed
}
