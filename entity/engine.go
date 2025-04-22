package entity

import (
	"regexp"
)

// ReplaceMyselfLink представляет настройки для замены ссылок на текущего бота
type ReplaceMyselfLink struct {
	// ChatID идентификатор чата, для которого применяется замена
	ChatID int64
	// DeleteExternal если true, то внешние ссылки удаляются
	DeleteExternal bool
}

// ReplaceFragment представляет настройки для замены фрагментов текста
type ReplaceFragment struct {
	// ChatID идентификатор чата, для которого применяется замена
	ChatID int64
	// Replacements карта замен (ключ - исходный текст, значение - текст для замены)
	Replacements map[string]string
}

// Source представляет настройки источника сообщений
type Source struct {
	// ID идентификатор чата-источника
	ID int64
	// Sign настройки подписи для сообщений из этого источника
	Sign *SignSettings
	// Link настройки ссылки на источник
	Link *LinkSettings
}

// SignSettings представляет настройки подписи для сообщений
type SignSettings struct {
	// Title текст подписи (с поддержкой разметки)
	Title string
	// For список идентификаторов чатов, для которых применяется подпись
	For []int64
}

// LinkSettings представляет настройки ссылки на источник сообщений
type LinkSettings struct {
	// Title текст ссылки (с поддержкой разметки)
	Title string
	// For список идентификаторов чатов, для которых применяется ссылка
	For []int64
}

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
	Regexp string `json:"regexp"`
	// CompiledRegexp скомпилированное регулярное выражение
	CompiledRegexp *regexp.Regexp `json:"-"`
	// Group номер группы в регулярном выражении для сравнения
	Group int `json:"group"`
	// Match список строк для сравнения с подстрокой
	Match []string `json:"match"`
}

type Answer struct {
	ChatID int64
	Auto   bool
}
