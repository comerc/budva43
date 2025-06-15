package transform

import (
	"errors"
	"strings"
	"testing"

	"github.com/comerc/budva43/service/transform/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/zelenin/go-tdlib/client"
)

// data for service.transform - 101xx

func TestTransformService_addSources(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		formattedText    *client.FormattedText
		src              *client.Message
		dstChatId        int64
		expectedText     string
		expectedEntities []*client.TextEntity
		setupMocks       func(telegramRepo *mocks.TelegramRepo, src *client.Message)
	}{
		{
			name:          "source not found",
			formattedText: &client.FormattedText{},
			src: &client.Message{
				ChatId: 99999, // не существует в config.yml
				Id:     123,
			},
			// dstChatId:        10109,
			expectedText:     "",
			expectedEntities: nil,
		},
		{
			name:          "sign only",
			formattedText: &client.FormattedText{},
			src: &client.Message{
				ChatId: 10100,
				Id:     123,
			},
			dstChatId:        10109,
			expectedText:     "Test Source",
			expectedEntities: nil,
			setupMocks: func(telegramRepo *mocks.TelegramRepo, src *client.Message) {
				telegramRepo.EXPECT().ParseTextEntities(&client.ParseTextEntitiesRequest{
					Text: "Test Source",
					ParseMode: &client.TextParseModeMarkdown{
						Version: 2,
					},
				}).Return(&client.FormattedText{
					Text:     "Test Source",
					Entities: nil,
				}, nil)
			},
		},
		{
			name:          "link only",
			formattedText: &client.FormattedText{},
			src: &client.Message{
				ChatId:       10101,
				Id:           123,
				MediaAlbumId: 0,
			},
			dstChatId:        10109,
			expectedText:     "[🔗Source Link](https://t.me/test/123)",
			expectedEntities: nil,
			setupMocks: func(telegramRepo *mocks.TelegramRepo, src *client.Message) {
				telegramRepo.EXPECT().GetMessageLink(&client.GetMessageLinkRequest{
					ChatId:    src.ChatId,
					MessageId: src.Id,
					ForAlbum:  src.MediaAlbumId != 0,
				}).Return(&client.MessageLink{
					Link: "https://t.me/test/123",
				}, nil)

				telegramRepo.EXPECT().ParseTextEntities(&client.ParseTextEntitiesRequest{
					Text: "[🔗Source Link](https://t.me/test/123)",
					ParseMode: &client.TextParseModeMarkdown{
						Version: 2,
					},
				}).Return(&client.FormattedText{
					Text:     "[🔗Source Link](https://t.me/test/123)",
					Entities: nil,
				}, nil)
			},
		},
		{
			name: "sign and link",
			formattedText: &client.FormattedText{
				Text:     "existing",
				Entities: []*client.TextEntity{},
			},
			src: &client.Message{
				ChatId:       10102,
				Id:           123,
				MediaAlbumId: 0,
			},
			dstChatId:        10109,
			expectedText:     "existing\n\nTest Source\n\n[🔗Source Link](https://t.me/test/123)",
			expectedEntities: []*client.TextEntity{},
			setupMocks: func(telegramRepo *mocks.TelegramRepo, src *client.Message) {
				telegramRepo.EXPECT().ParseTextEntities(&client.ParseTextEntitiesRequest{
					Text: "Test Source",
					ParseMode: &client.TextParseModeMarkdown{
						Version: 2,
					},
				}).Return(&client.FormattedText{
					Text:     "Test Source",
					Entities: nil,
				}, nil)

				telegramRepo.EXPECT().GetMessageLink(&client.GetMessageLinkRequest{
					ChatId:    src.ChatId,
					MessageId: src.Id,
					ForAlbum:  src.MediaAlbumId != 0,
				}).Return(&client.MessageLink{
					Link: "https://t.me/test/123",
				}, nil)

				telegramRepo.EXPECT().ParseTextEntities(&client.ParseTextEntitiesRequest{
					Text: "[🔗Source Link](https://t.me/test/123)",
					ParseMode: &client.TextParseModeMarkdown{
						Version: 2,
					},
				}).Return(&client.FormattedText{
					Text:     "[🔗Source Link](https://t.me/test/123)",
					Entities: nil,
				}, nil)
			},
		},
		{
			name:          "sign not for this chat",
			formattedText: &client.FormattedText{},
			src: &client.Message{
				ChatId: 10103,
				Id:     123,
			},
			dstChatId:        10109,
			expectedText:     "",
			expectedEntities: nil,
		},
		{
			name:          "empty source",
			formattedText: &client.FormattedText{},
			src: &client.Message{
				ChatId: 10104,
				Id:     123,
			},
			dstChatId:        10109,
			expectedText:     "",
			expectedEntities: nil,
		},
		{
			name:          "get message link error",
			formattedText: &client.FormattedText{},
			src: &client.Message{
				ChatId:       10105,
				Id:           123,
				MediaAlbumId: 0,
			},
			dstChatId:        10109,
			expectedText:     "",
			expectedEntities: nil,
			setupMocks: func(telegramRepo *mocks.TelegramRepo, src *client.Message) {
				telegramRepo.EXPECT().GetMessageLink(&client.GetMessageLinkRequest{
					ChatId:    src.ChatId,
					MessageId: src.Id,
					ForAlbum:  src.MediaAlbumId != 0,
				}).Return(nil, errors.New("get message link error"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			telegramRepo := mocks.NewTelegramRepo(t)

			if test.setupMocks != nil {
				test.setupMocks(telegramRepo, test.src)
			}

			service := New(telegramRepo, nil, nil)
			service.addSources(test.formattedText, test.src, test.dstChatId)

			assert.Equal(t, test.expectedText, test.formattedText.Text)
			assert.Equal(t, test.expectedEntities, test.formattedText.Entities)
		})
	}
}

func TestTransformService_addText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		formattedText    *client.FormattedText
		text             string
		expectedError    error
		expectedText     string
		expectedEntities []*client.TextEntity
	}{
		{
			name:             "empty",
			formattedText:    &client.FormattedText{},
			text:             "test",
			expectedError:    nil,
			expectedText:     "test",
			expectedEntities: nil,
		},
		{
			name:             "with error",
			formattedText:    &client.FormattedText{},
			text:             "new text",
			expectedError:    errors.New("error"),
			expectedText:     "",
			expectedEntities: nil,
		},
		{
			name: "with existing text",
			formattedText: &client.FormattedText{
				Text:     "existing",
				Entities: []*client.TextEntity{},
			},
			text:             "new text",
			expectedError:    nil,
			expectedText:     "existing\n\nnew text",
			expectedEntities: []*client.TextEntity{},
		},
		{
			name: "add text with entities",
			formattedText: &client.FormattedText{
				Text:     "existing",
				Entities: []*client.TextEntity{},
			},
			text:          "*bold*",
			expectedError: nil,
			expectedText:  "existing\n\n*bold*",
			expectedEntities: []*client.TextEntity{
				{
					Type:   &client.TextEntityTypeBold{},
					Offset: 10,
					Length: 6,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			telegramRepo := mocks.NewTelegramRepo(t)
			telegramRepo.EXPECT().ParseTextEntities(&client.ParseTextEntitiesRequest{
				Text: test.text,
				ParseMode: &client.TextParseModeMarkdown{
					Version: 2,
				},
			}).Return(&client.FormattedText{
				Text:     test.text,
				Entities: test.expectedEntities,
			}, test.expectedError)

			service := New(telegramRepo, nil, nil)
			service.addText(test.formattedText, test.text)

			assert.Equal(t, test.expectedText, test.formattedText.Text)
			assert.Equal(t, test.expectedEntities, test.formattedText.Entities)
		})
	}
}

func TestEscapeMarkdown(t *testing.T) {
	t.Parallel()

	s1 := "_ * ( ) ~ ` > # + = | { } . !"
	s2 := `\[ \] \-`
	a := strings.Split(s1+" "+s2, " ")
	for _, v := range a {
		assert.Equal(t, `\`+v, escapeMarkdown(v))
	}
}
