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
	// Prev текст для ссылки на предыдущую версию сообщения
	Prev string
}

// Sign представляет настройки подписи для сообщений
type Sign struct {
	// Title текст подписи (с поддержкой разметки)
	Title string
	// For список идентификаторов чатов, для которых применяется подпись
	For []ChatId // TODO: убрать
}

// Link представляет настройки ссылки на источник сообщений
type Link struct {
	// Title текст ссылки (с поддержкой разметки)
	Title string
	// For список идентификаторов чатов, для которых применяется ссылка
	For []ChatId // TODO: убрать
}
