package forward_rule

import "regexp"

// ForwardRule представляет правило пересылки сообщений
type ForwardRule struct {
	// ID уникальный идентификатор правила
	ID string
	// From идентификатор чата-источника
	From int64
	// To список идентификаторов чатов-получателей
	To []int64
	// SendCopy если true, то отправляет копию сообщения вместо пересылки
	SendCopy bool
	// CopyOnce если true, то сообщение копируется однократно без синхронизации при редактировании
	CopyOnce bool
	// Indelible если true, то сообщение не удаляется при удалении оригинала
	Indelible bool
	// Exclude регулярное выражение для исключения сообщений
	Exclude string
	// ExcludeRegexp скомпилированное регулярное выражение для исключения
	ExcludeRegexp *regexp.Regexp
	// Include регулярное выражение для включения сообщений
	Include string
	// IncludeRegexp скомпилированное регулярное выражение для включения
	IncludeRegexp *regexp.Regexp
	// IncludeSubmatch правила для подстрок в сообщениях
	IncludeSubmatch []SubmatchRule
	// Other идентификатор чата для отправки сообщений, которые прошли включающий фильтр
	Other int64
	// Check идентификатор чата для отправки сообщений, которые прошли исключающий фильтр
	Check int64
	// Status статус активности правила
	Status RuleStatus
}

// RuleStatus представляет статус правила
type RuleStatus string

const (
	// RuleStatusActive правило активно
	RuleStatusActive RuleStatus = "active"
	// RuleStatusInactive правило неактивно
	RuleStatusInactive RuleStatus = "inactive"
	// RuleStatusPaused правило временно приостановлено
	RuleStatusPaused RuleStatus = "paused"
)

// SubmatchRule представляет правило для работы с подстроками в сообщениях
type SubmatchRule struct {
	// Regexp регулярное выражение для поиска подстрок
	Regexp string
	// CompiledRegexp скомпилированное регулярное выражение
	CompiledRegexp *regexp.Regexp
	// Group номер группы в регулярном выражении для сравнения
	Group int
	// Match список строк для сравнения с подстрокой
	Match []string
}

// NewForwardRule создает новый экземпляр правила пересылки
func NewForwardRule(id string, from int64, to []int64) *ForwardRule {
	return &ForwardRule{
		ID:     id,
		From:   from,
		To:     to,
		Status: RuleStatusActive,
	}
}
