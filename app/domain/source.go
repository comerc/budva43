package domain

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
	// Prev настройки ссылки на предыдущую версию сообщения
	Prev *Prev
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

// Prev представляет настройки ссылки на предыдущую версию сообщения
type Prev struct {
	// Title текст ссылки (с поддержкой разметки)
	Title string
	// For список идентификаторов чатов, для которых применяется ссылка
	For []ChatId
}

// SIGN_TITLE название подписи
const SIGN_TITLE = "Sign"

// LINK_TITLE название ссылки на источник сообщений
const LINK_TITLE = "Link"

// PREV_TITLE название ссылки на предыдущее сообщение
const PREV_TITLE = "Prev"
