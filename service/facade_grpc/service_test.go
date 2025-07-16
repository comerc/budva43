package facade_grpc

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zelenin/go-tdlib/client"

	"github.com/comerc/budva43/app/dto/grpc/dto"
	"github.com/comerc/budva43/service/facade_grpc/mocks"
)

func TestGetMessages(t *testing.T) {
	t.Parallel()

	tg := mocks.NewTelegramRepo(t)
	ms := mocks.NewMessageService(t)
	s := New(tg, ms, nil)

	chatId := int64(1)
	msgIds := []int64{10, 20}
	msg1 := &client.Message{Id: 10}
	msg2 := &client.Message{Id: 20}
	msgs := &client.Messages{Messages: []*client.Message{msg1, msg2}}
	tg.EXPECT().GetMessages(&client.GetMessagesRequest{ChatId: chatId, MessageIds: msgIds}).Return(msgs, nil)
	tg.EXPECT().GetMarkdownText(&client.GetMarkdownTextRequest{
		Text: &client.FormattedText{Text: "foo"},
	}).Return(&client.FormattedText{Text: "foo"}, nil)
	tg.EXPECT().GetMarkdownText(&client.GetMarkdownTextRequest{
		Text: &client.FormattedText{Text: "bar"},
	}).Return(&client.FormattedText{Text: "bar"}, nil)
	ms.EXPECT().GetFormattedText(msg1).Return(&client.FormattedText{Text: "foo"})
	ms.EXPECT().GetFormattedText(msg2).Return(&client.FormattedText{Text: "bar"})

	result, err := s.GetMessages(chatId, msgIds)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, int64(10), result[0].Id)
	assert.Equal(t, "foo", result[0].Text)
	assert.Equal(t, int64(20), result[1].Id)
	assert.Equal(t, "bar", result[1].Text)
}

func TestGetLastMessage(t *testing.T) {
	t.Parallel()

	tg := mocks.NewTelegramRepo(t)
	ms := mocks.NewMessageService(t)
	s := New(tg, ms, nil)

	chatId := int64(1)
	msg := &client.Message{Id: 42}
	tg.EXPECT().GetChatHistory(&client.GetChatHistoryRequest{
		ChatId:    chatId,
		Limit:     1,
		OnlyLocal: true,
	}).Return(&client.Messages{
		TotalCount: 1,
		Messages:   []*client.Message{msg},
	}, nil)
	tg.EXPECT().GetMarkdownText(&client.GetMarkdownTextRequest{
		Text: &client.FormattedText{Text: "last"},
	}).Return(&client.FormattedText{Text: "last"}, nil)
	ms.EXPECT().GetFormattedText(msg).Return(&client.FormattedText{Text: "last"})

	result, err := s.GetLastMessage(chatId)
	assert.NoError(t, err)
	assert.Equal(t, int64(42), result.Id)
	assert.Equal(t, "last", result.Text)
}

func TestSendMessage(t *testing.T) {
	t.Parallel()

	tg := mocks.NewTelegramRepo(t)
	ms := mocks.NewMessageService(t)
	s := New(tg, ms, nil)

	in := &dto.NewMessage{ChatId: 1, Text: "hi", ReplyToMessageId: 2}
	msg := &client.Message{Id: 100}
	tg.EXPECT().SendMessage(&client.SendMessageRequest{
		ChatId: 1,
		InputMessageContent: &client.InputMessageText{
			Text:               &client.FormattedText{Text: "hi"},
			LinkPreviewOptions: &client.LinkPreviewOptions{IsDisabled: true},
			ClearDraft:         true,
		},
		ReplyTo: &client.InputMessageReplyToMessage{MessageId: 2},
	}).Return(msg, nil)
	tg.EXPECT().ParseTextEntities(&client.ParseTextEntitiesRequest{
		Text: "hi",
		ParseMode: &client.TextParseModeMarkdown{
			Version: 2,
		},
	}).Return(&client.FormattedText{Text: "hi"}, nil)
	tg.EXPECT().GetMarkdownText(&client.GetMarkdownTextRequest{
		Text: &client.FormattedText{Text: "hi"},
	}).Return(&client.FormattedText{Text: "hi"}, nil)
	ms.EXPECT().GetFormattedText(msg).Return(&client.FormattedText{Text: "hi"})

	result, err := s.SendMessage(in)
	assert.NoError(t, err)
	assert.Equal(t, int64(100), result.Id)
	assert.Equal(t, "hi", result.Text)
}

func TestForwardMessage(t *testing.T) {
	t.Parallel()

	tg := mocks.NewTelegramRepo(t)
	ms := mocks.NewMessageService(t)
	s := New(tg, ms, nil)

	chatId := int64(1)
	msgId := int64(2)
	msg := &client.Message{Id: msgId}
	tg.EXPECT().ForwardMessages(&client.ForwardMessagesRequest{
		ChatId:     chatId,
		MessageIds: []int64{msgId},
	}).Return(&client.Messages{
		TotalCount: 1,
		Messages:   []*client.Message{msg},
	}, nil)
	tg.EXPECT().GetMarkdownText(&client.GetMarkdownTextRequest{
		Text: &client.FormattedText{Text: "msg"},
	}).Return(&client.FormattedText{Text: "msg"}, nil)
	ms.EXPECT().GetFormattedText(msg).Return(&client.FormattedText{Text: "msg"})

	result, err := s.ForwardMessage(chatId, msgId)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), result.Id)
	assert.Equal(t, "msg", result.Text)
}

func TestGetMessage(t *testing.T) {
	t.Parallel()

	tg := mocks.NewTelegramRepo(t)
	ms := mocks.NewMessageService(t)
	s := New(tg, ms, nil)

	chatId := int64(1)
	msgId := int64(2)
	msg := &client.Message{Id: msgId}
	tg.EXPECT().GetMessage(&client.GetMessageRequest{ChatId: chatId, MessageId: msgId}).Return(msg, nil)
	tg.EXPECT().GetMarkdownText(&client.GetMarkdownTextRequest{
		Text: &client.FormattedText{Text: "msg"},
	}).Return(&client.FormattedText{Text: "msg"}, nil)
	ms.EXPECT().GetFormattedText(msg).Return(&client.FormattedText{Text: "msg"})

	result, err := s.GetMessage(chatId, msgId)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), result.Id)
	assert.Equal(t, "msg", result.Text)
}

func TestUpdateMessage(t *testing.T) {
	t.Parallel()

	tg := mocks.NewTelegramRepo(t)
	ms := mocks.NewMessageService(t)
	s := New(tg, ms, nil)

	upd := &dto.Message{Id: 2, ChatId: 1, Text: "upd"}
	orig := &client.Message{Id: 2, ReplyMarkup: &client.ReplyMarkupInlineKeyboard{}} // пример
	newMsg := &client.Message{Id: 2}
	ft := &client.FormattedText{Text: "upd"}
	imc := &client.InputMessageText{Text: ft}
	tg.EXPECT().GetMessage(&client.GetMessageRequest{ChatId: 1, MessageId: 2}).Return(orig, nil)
	ms.EXPECT().GetInputMessageContent(orig, ft).Return(imc)
	tg.EXPECT().EditMessageText(&client.EditMessageTextRequest{
		ChatId:              1,
		MessageId:           2,
		ReplyMarkup:         orig.ReplyMarkup,
		InputMessageContent: imc,
	}).Return(newMsg, nil)
	tg.EXPECT().ParseTextEntities(&client.ParseTextEntitiesRequest{
		Text: "upd",
		ParseMode: &client.TextParseModeMarkdown{
			Version: 2,
		},
	}).Return(&client.FormattedText{Text: "upd"}, nil)
	tg.EXPECT().GetMarkdownText(&client.GetMarkdownTextRequest{
		Text: &client.FormattedText{Text: "upd"},
	}).Return(&client.FormattedText{Text: "upd"}, nil)
	ms.EXPECT().GetFormattedText(newMsg).Return(ft)

	result, err := s.UpdateMessage(upd)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), result.Id)
	assert.Equal(t, "upd", result.Text)
}

func TestDeleteMessages(t *testing.T) {
	t.Parallel()

	tg := mocks.NewTelegramRepo(t)
	ms := mocks.NewMessageService(t)
	s := New(tg, ms, nil)

	chatId := int64(1)
	msgIds := []int64{2, 3}
	tg.EXPECT().DeleteMessages(&client.DeleteMessagesRequest{ChatId: chatId, MessageIds: msgIds}).Return(&client.Ok{}, nil)

	ok, err := s.DeleteMessages(chatId, msgIds)
	assert.NoError(t, err)
	assert.True(t, ok)
}

func TestErrorFromRepo(t *testing.T) {
	t.Parallel()

	tg := mocks.NewTelegramRepo(t)
	ms := mocks.NewMessageService(t)
	s := New(tg, ms, nil)

	tg.EXPECT().GetMessages(&client.GetMessagesRequest{ChatId: 1, MessageIds: []int64{1}}).Return(nil, errors.New("fail"))
	msgs, err := s.GetMessages(1, []int64{1})
	assert.Error(t, err)
	assert.Nil(t, msgs)
}

func TestGetMessageLink(t *testing.T) {
	t.Parallel()

	tg := mocks.NewTelegramRepo(t)
	ms := mocks.NewMessageService(t)
	s := New(tg, ms, nil)

	chatId := int64(1)
	msgId := int64(2)
	link := "https://t.me/c/1/2"
	tg.EXPECT().GetMessageLink(&client.GetMessageLinkRequest{
		ChatId:    chatId,
		MessageId: msgId,
	}).Return(&client.MessageLink{Link: link}, nil)

	result, err := s.GetMessageLink(chatId, msgId)
	assert.NoError(t, err)
	assert.Equal(t, link, result)
}

func TestGetMessageLinkInfo(t *testing.T) {
	t.Parallel()

	tg := mocks.NewTelegramRepo(t)
	ms := mocks.NewMessageService(t)
	s := New(tg, ms, nil)

	link := "https://t.me/c/1/2"
	msg := &client.Message{Id: 2, ChatId: 1, ForwardInfo: &client.MessageForwardInfo{}}
	tg.EXPECT().GetMessageLinkInfo(&client.GetMessageLinkInfoRequest{Url: link}).Return(&client.MessageLinkInfo{Message: msg}, nil)

	result, err := s.GetMessageLinkInfo(link)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), result.Id)
	assert.Equal(t, int64(1), result.ChatId)
	assert.True(t, result.Forward)
}
