package entity

// ReplaceMyselfLinkSettings представляет настройки для замены ссылок на текущего бота
type ReplaceMyselfLinkSettings struct {
	// ChatID идентификатор чата, для которого применяется замена
	ChatID int64
	// DeleteExternal если true, то внешние ссылки удаляются
	DeleteExternal bool
}

// ReplaceFragmentSettings представляет настройки для замены фрагментов текста
type ReplaceFragmentSettings struct {
	// ChatID идентификатор чата, для которого применяется замена
	ChatID int64
	// Replacements карта замен (ключ - исходный текст, значение - текст для замены)
	Replacements map[string]string
}

// NewReplaceMyselfLinkSettings создает новые настройки замены ссылок
func NewReplaceMyselfLinkSettings(chatID int64, deleteExternal bool) *ReplaceMyselfLinkSettings {
	return &ReplaceMyselfLinkSettings{
		ChatID:         chatID,
		DeleteExternal: deleteExternal,
	}
}

// NewReplaceFragmentSettings создает новые настройки замены фрагментов текста
func NewReplaceFragmentSettings(chatID int64) *ReplaceFragmentSettings {
	return &ReplaceFragmentSettings{
		ChatID:       chatID,
		Replacements: make(map[string]string),
	}
}

// AddReplacement добавляет новую замену фрагмента текста
func (r *ReplaceFragmentSettings) AddReplacement(from, to string) {
	if r.Replacements == nil {
		r.Replacements = make(map[string]string)
	}
	r.Replacements[from] = to
}
