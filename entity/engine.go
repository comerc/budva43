package entity

import (
	"regexp"
)

// ChatId идентификатор чата
type ChatId = int64

// ReplaceMyselfLink представляет настройки для замены ссылок на текущего бота
type ReplaceMyselfLink struct {
	// Id идентификатор чата-источника - обогощаем при загрузке
	ChatId ChatId
	// DeleteExternal если true, то внешние ссылки удаляются
	DeleteExternal bool
}

// Replacements карта замен (ключ - исходный текст, значение - текст для замены)
type Replacements = map[string]string

// ReplaceFragment представляет настройки для замены фрагментов текста
type ReplaceFragment struct {
	// Id идентификатор чата-источника - обогощаем при загрузке
	ChatId ChatId
	// Replacements карта замен (ключ - исходный текст, значение - текст для замены)
	Replacements
}

// Source представляет настройки источника сообщений
type Source struct {
	// Id идентификатор чата-источника - обогощаем при загрузке
	ChatId ChatId
	// Sign настройки подписи для сообщений из этого источника
	Sign *Sign
	// Link настройки ссылки на источник
	Link *Link
}

// Sign представляет настройки подписи для сообщений
type Sign struct {
	// Title текст подписи (с поддержкой разметки)
	Title string
	// For список идентификаторов чатов, для которых применяется подпись
	For []ChatId
}

// Link представляет настройки ссылки на источник сообщений
type Link struct {
	// Title текст ссылки (с поддержкой разметки)
	Title string
	// For список идентификаторов чатов, для которых применяется ссылка
	For []ChatId
}

type ForwardRuleId string

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
	// ExcludeRegexp скомпилированное регулярное выражение для исключения
	ExcludeRegexp *regexp.Regexp
	// Include регулярное выражение для включения сообщений
	Include string
	// IncludeRegexp скомпилированное регулярное выражение для включения
	IncludeRegexp *regexp.Regexp
	// IncludeSubmatch правила для подстрок в сообщениях
	IncludeSubmatch []SubmatchRule
	// Other идентификатор чата для отправки сообщений, которые прошли включающий фильтр
	Other ChatId
	// Check идентификатор чата для отправки сообщений, которые прошли исключающий фильтр
	Check ChatId
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

// type Answer struct {
// 	ChatId ChatId
// 	Auto   bool
// }
