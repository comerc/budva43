package entity

import (
	"regexp"
)

// ChatId идентификатор чата
type ChatId = int64

type Destination struct {
	// Id идентификатор чата-получателя - обогощаем при загрузке
	ChatId ChatId
	// ReplaceMyselfLinks настройки для замены ссылок на текущего бота
	ReplaceMyselfLinks *ReplaceMyselfLinks
	// ReplaceFragments настройки для замены фрагментов текста
	ReplaceFragments []*ReplaceFragment
}

// ReplaceMyselfLinks настройки для замены ссылок на текущего бота
type ReplaceMyselfLinks struct {
	// Run если true, то замена ссылок включена
	Run bool
	// DeleteExternal если true, то внешние ссылки удаляются
	DeleteExternal bool
}

// ReplaceFragment представляет настройки для замены фрагмента текста
type ReplaceFragment struct {
	// From исходный текст
	From string
	// To текст для замены
	To string
}

// Source представляет настройки источника сообщений
type Source struct {
	// Id идентификатор чата-источника - обогощаем при загрузке
	ChatId ChatId
	// Sign настройки подписи для сообщений из этого источника
	Sign *Sign
	// Link настройки ссылки на источник
	Link *Link
	// AutoAnswer настройки автоматического ответа
	AutoAnswer bool
	// DeleteSystemMessages настройки удаления системных сообщений
	DeleteSystemMessages bool
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

type ForwardRuleId = string

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

// MediaAlbumKey ключ для пересылаемого медиа-альбома
type MediaAlbumKey = string
