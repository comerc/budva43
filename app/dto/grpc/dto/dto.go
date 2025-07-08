package dto

type Chat struct {
	Id       int64
	Name     string
	Messages []*Message
}

type Message struct {
	Id      int64
	Content string
	ChatId  int64
}

type NewMessage struct {
	Content string
	ChatId  int64
}
