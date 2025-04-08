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

// AddReplacement добавляет новую замену фрагмента текста
func (r *ReplaceFragmentSettings) AddReplacement(from, to string) {
	if r.Replacements == nil {
		r.Replacements = make(map[string]string)
	}
	r.Replacements[from] = to
}
