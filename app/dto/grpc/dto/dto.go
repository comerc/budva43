package dto

type Chat struct {
	Id       int64
	Name     string
	Messages []*Message
}

type Message struct {
	Id       int64
	ChatId   int64
	Text     string
	Forward  bool
	FilePath string
}

type NewMessage struct {
	ChatId           int64
	Text             string
	ReplyToMessageId int64
	FilePath         string
}
