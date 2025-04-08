package entity

import (
	"encoding/json"
	"regexp"
)

// ForwardRule представляет правило пересылки сообщений
type ForwardRule struct {
	// ID уникальный идентификатор правила
	ID string `json:"id"`
	// From идентификатор чата-источника
	From int64 `json:"from"`
	// To список идентификаторов чатов-получателей
	To []int64 `json:"to"`
	// SendCopy если true, то отправляет копию сообщения вместо пересылки
	SendCopy bool `json:"send_copy"`
	// CopyOnce если true, то сообщение копируется однократно без синхронизации при редактировании
	CopyOnce bool `json:"copy_once"`
	// Indelible если true, то сообщение не удаляется при удалении оригинала
	Indelible bool `json:"indelible"`
	// Exclude регулярное выражение для исключения сообщений
	Exclude string `json:"exclude,omitempty"`
	// ExcludeRegexp скомпилированное регулярное выражение для исключения
	ExcludeRegexp *regexp.Regexp `json:"-"`
	// Include регулярное выражение для включения сообщений
	Include string `json:"include,omitempty"`
	// IncludeRegexp скомпилированное регулярное выражение для включения
	IncludeRegexp *regexp.Regexp `json:"-"`
	// IncludeSubmatch правила для подстрок в сообщениях
	IncludeSubmatch []SubmatchRule `json:"include_submatch,omitempty"`
	// Other идентификатор чата для отправки сообщений, которые прошли включающий фильтр
	Other int64 `json:"other,omitempty"`
	// Check идентификатор чата для отправки сообщений, которые прошли исключающий фильтр
	Check int64 `json:"check,omitempty"`
	// Status статус активности правила
	Status RuleStatus `json:"status"`
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

// MarshalJSON реализует интерфейс json.Marshaler для SubmatchRule
func (s SubmatchRule) MarshalJSON() ([]byte, error) {
	// Создаем копию структуры без полей, которые не должны быть сериализованы
	type SubmatchRuleAlias SubmatchRule
	return json.Marshal((*SubmatchRuleAlias)(&s))
}

// UnmarshalJSON реализует интерфейс json.Unmarshaler для SubmatchRule
func (s *SubmatchRule) UnmarshalJSON(data []byte) error {
	// Создаем копию структуры без полей, которые не должны быть десериализованы
	type SubmatchRuleAlias SubmatchRule
	alias := (*SubmatchRuleAlias)(s)

	if err := json.Unmarshal(data, alias); err != nil {
		return err
	}

	// Компилируем регулярное выражение если оно указано
	if s.Regexp != "" {
		var err error
		s.CompiledRegexp, err = regexp.Compile(s.Regexp)
		if err != nil {
			return err
		}
	}

	return nil
}

// MarshalJSON реализует интерфейс json.Marshaler для ForwardRule
func (r ForwardRule) MarshalJSON() ([]byte, error) {
	// Создаем копию структуры без полей, которые не должны быть сериализованы
	type ForwardRuleAlias ForwardRule
	return json.Marshal((*ForwardRuleAlias)(&r))
}

// UnmarshalJSON реализует интерфейс json.Unmarshaler для ForwardRule
func (r *ForwardRule) UnmarshalJSON(data []byte) error {
	// Создаем копию структуры без полей, которые не должны быть десериализованы
	type ForwardRuleAlias ForwardRule
	alias := (*ForwardRuleAlias)(r)

	if err := json.Unmarshal(data, alias); err != nil {
		return err
	}

	// Компилируем регулярные выражения если они указаны
	if r.Exclude != "" {
		var err error
		r.ExcludeRegexp, err = regexp.Compile(r.Exclude)
		if err != nil {
			return err
		}
	}

	if r.Include != "" {
		var err error
		r.IncludeRegexp, err = regexp.Compile(r.Include)
		if err != nil {
			return err
		}
	}

	return nil
}
