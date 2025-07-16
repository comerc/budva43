package domain

import "regexp"

type ForwardRuleId = string
type FiltersMode = string

const (
	FiltersOK    FiltersMode = "ok"
	FiltersCheck FiltersMode = "check"
	FiltersOther FiltersMode = "other"
)

// ForwardRule представляет правило пересылки сообщений
type ForwardRule struct {
	// Id уникальный идентификатор правила - обогощаем при загрузке
	Id ForwardRuleId
	// From идентификатор чата-источника
	From ChatId
	// To список идентификаторов чатов-получателей
	To []ChatId
	// SendCopy если true, то отправляет копию сообщения вместо пересылки
	SendCopy bool
	// CopyOnce если true, то сообщение копируется однократно без синхронизации при редактировании
	CopyOnce bool
	// Indelible если true, то сообщение не удаляется при удалении оригинала
	Indelible bool
	// Exclude регулярное выражение для исключения сообщений
	Exclude string
	// Include регулярное выражение для включения сообщений
	Include string
	// IncludeSubmatch правила для подстрок в сообщениях
	IncludeSubmatch []*SubmatchRule
	// Other идентификатор чата для отправки сообщений, которые прошли включающий фильтр
	Other ChatId
	// Check идентификатор чата для отправки сообщений, которые прошли исключающий фильтр
	Check ChatId
}

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
