package log

type Options struct {
	ErrorSource *SourceOptions
}

type SourceOptions struct {
	Type         SourceType
	RelativePath bool
}

type SourceType = string

var TypeSourceNone SourceType = ""
var TypeSourceSimple SourceType = "simple"
var TypeSourceCallStack SourceType = "callstack"

var options *Options
