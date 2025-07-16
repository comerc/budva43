package domain

// Настройки движка форвардинга
type EngineConfig struct {
	// Настройки источников
	Sources map[ChatId]*Source
	// Настройки получателей
	Destinations map[ChatId]*Destination
	// Правила форвардинга
	ForwardRules map[ForwardRuleId]*ForwardRule
	// Уникальные источники
	UniqueSources map[ChatId]struct{} `mapstructure:"-"`
	// Уникальные получатели
	UniqueDestinations map[ChatId]struct{} `mapstructure:"-"`
	// Порядок форвардинга
	OrderedForwardRules []ForwardRuleId `mapstructure:"-"`
}
