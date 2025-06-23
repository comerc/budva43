package entity

// Настройки источника ошибок
type ErrorSource struct {
	Type         ErrorSourceType
	RelativePath bool
}

type ErrorSourceType = string

var TypeErrorSourceNone ErrorSourceType = ""
var TypeErrorSourceOne ErrorSourceType = "one"
var TypeErrorSourceMore ErrorSourceType = "more"
