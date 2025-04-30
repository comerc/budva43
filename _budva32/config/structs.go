package config

const filename = ".config.yml"

type Config struct {
	Answers              map[int64]Answer
	ReplaceMyselfLinks   map[int64]ReplaceMyselfLink
	ReplaceFragments     map[int64]map[string]string
	Sources              map[int64]Source
	Reports              Reports
	Forwards             map[string]Forward
	DeleteSystemMessages map[int64]struct{}
}

type Answer struct {
	Auto bool
}

type ReplaceMyselfLink struct {
	DeleteExternal bool
}

type Source struct {
	Sign Sign
	Link Link
}

type Sign struct {
	Title string
	For   []int64 // TODO: map[int64]struct{}
}

type Link struct {
	Title string
	For   []int64 // TODO: map[int64]struct{}
}

type Reports struct {
	Template string
	For      []int64 // TODO: map[int64]struct{}
}

type Forward struct {
	From            int64
	To              []int64 // TODO: map[int64]struct{}
	Exclude         string
	Include         string
	IncludeSubmatch []IncludeSubmatch
	SendCopy        bool
	CopyOnce        bool
	Indelible       bool
	Check           int64 // то, что нашёл Exclude
	Other           int64 // то, что отсек Exclude
}

type IncludeSubmatch struct {
	Regexp string
	Group  int64
	Match  []string
}
