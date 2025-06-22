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
var TypeSourceOne SourceType = "one"
var TypeSourceMore SourceType = "more"

var options *Options
