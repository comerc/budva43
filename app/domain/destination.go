package domain

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
	// DeletedLinkText текст для замены внешней ссылки
	DeletedLinkText string
}

// DELETED_LINK константа для замены внешней ссылки
const DELETED_LINK = "DELETED_LINK"

// ReplaceFragment представляет настройки для замены фрагмента текста
type ReplaceFragment struct {
	// From исходный текст
	From string
	// To текст для замены
	To string
}
