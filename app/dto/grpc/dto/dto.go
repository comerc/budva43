package dto

type Chat struct {
	Id       int64
	Name     string
	Messages []*Message
}

type Message struct {
	Id     int64
	Text   string
	ChatId int64
}

type NewMessage struct {
	Text             string
	ChatId           int64
	ReplyToMessageId int64
}
