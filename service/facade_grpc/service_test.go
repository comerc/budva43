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
	s := New(tg, ms)

	chatId := int64(1)
	msgIds := []int64{10, 20}
	msg1 := &client.Message{Id: 10}
	msg2 := &client.Message{Id: 20}
	msgs := &client.Messages{Messages: []*client.Message{msg1, msg2}}
	tg.EXPECT().GetMessages(&client.GetMessagesRequest{ChatId: chatId, MessageIds: msgIds}).Return(msgs, nil)
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
	s := New(tg, ms)

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
	s := New(tg, ms)

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
	s := New(tg, ms)

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
	s := New(tg, ms)

	chatId := int64(1)
	msgId := int64(2)
	msg := &client.Message{Id: msgId}
	tg.EXPECT().GetMessage(&client.GetMessageRequest{ChatId: chatId, MessageId: msgId}).Return(msg, nil)
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
	s := New(tg, ms)

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
	s := New(tg, ms)

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
	s := New(tg, ms)

	tg.EXPECT().GetMessages(&client.GetMessagesRequest{ChatId: 1, MessageIds: []int64{1}}).Return(nil, errors.New("fail"))
	msgs, err := s.GetMessages(1, []int64{1})
	assert.Error(t, err)
	assert.Nil(t, msgs)
}
