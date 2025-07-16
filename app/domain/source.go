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

// YETI_MESSAGE константа для игнорирования сообщений
const YETI_MESSAGE = "YETI_MESSAGE"
