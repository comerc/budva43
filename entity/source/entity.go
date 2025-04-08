package source

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

// NewSource создает новый экземпляр источника сообщений
func NewSource(id int64) *Source {
	return &Source{
		ID: id,
	}
}

// WithSign добавляет настройки подписи к источнику
func (s *Source) WithSign(title string, for_ []int64) *Source {
	s.Sign = &SignSettings{
		Title: title,
		For:   for_,
	}
	return s
}

// WithLink добавляет настройки ссылки к источнику
func (s *Source) WithLink(title string, for_ []int64) *Source {
	s.Link = &LinkSettings{
		Title: title,
		For:   for_,
	}
	return s
}
