package entity

// Настройки источника ошибок
type ErrorSource struct {
	Type         ErrorSourceType
	RelativePath bool
}

// TODO: убрать, т.к. противоречит соглашению по ошибкам - интересует только "more"-режим

type ErrorSourceType = string

var TypeErrorSourceNone ErrorSourceType = ""
var TypeErrorSourceOne ErrorSourceType = "one"
var TypeErrorSourceMore ErrorSourceType = "more"
