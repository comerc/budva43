package transform

import (
	"errors"
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zelenin/go-tdlib/client"

	"github.com/comerc/budva43/app/log"
	"github.com/comerc/budva43/app/spylog"
	"github.com/comerc/budva43/service/transform/mocks"
)

// data for service.transform - 101xx

func TestTransformService_Transform(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		formattedText    *client.FormattedText
		withSources      bool
		src              *client.Message
		dstChatId        int64
		expectedText     string
		expectedEntities []*client.TextEntity // TODO: пока заглушка, надо подставлять реальные значения
		setup            func(t *testing.T) *Service
	}{
		{
			name: "with sources and markdown",
			formattedText: &client.FormattedText{
				Text:     "test message",
				Entities: []*client.TextEntity{},
			},
			withSources: true,
			src: &client.Message{
				ChatId:       10001,
				Id:           123,
				MediaAlbumId: 0,
			},
			dstChatId:        10003,
			expectedText:     "test message\n\n*Test Source*\n\n[🔗*Source Link*](https://t.me/test/123)",
			expectedEntities: []*client.TextEntity{},
			setup: func(t *testing.T) *Service {
				telegramRepo := mocks.NewTelegramRepo(t)
				storageService := mocks.NewStorageService(t)
				messageService := mocks.NewMessageService(t)

				// Mock для addSources - sign
				telegramRepo.EXPECT().ParseTextEntities(&client.ParseTextEntitiesRequest{
					Text: "*Test Source*",
					ParseMode: &client.TextParseModeMarkdown{
						Version: 2,
					},
				}).Return(&client.FormattedText{
					Text:     "*Test Source*",
					Entities: []*client.TextEntity{},
				}, nil)

				// Mock для addSources - link
				telegramRepo.EXPECT().GetMessageLink(&client.GetMessageLinkRequest{
					ChatId:    10001,
					MessageId: 123,
					ForAlbum:  false,
				}).Return(&client.MessageLink{
					Link: "https://t.me/test/123",
				}, nil)

				telegramRepo.EXPECT().ParseTextEntities(&client.ParseTextEntitiesRequest{
					Text: "[🔗*Source Link*](https://t.me/test/123)",
					ParseMode: &client.TextParseModeMarkdown{
						Version: 2,
					},
				}).Return(&client.FormattedText{
					Text:     "[🔗*Source Link*](https://t.me/test/123)",
					Entities: []*client.TextEntity{},
				}, nil)

				return New(telegramRepo, storageService, messageService)
			},
		},
		{
			name: "with sources and escaped markdown",
			formattedText: &client.FormattedText{
				Text:     "test message",
				Entities: []*client.TextEntity{},
			},
			withSources: true,
			src: &client.Message{
				ChatId:       10002,
				Id:           123,
				MediaAlbumId: 0,
			},
			dstChatId:        10003,
			expectedText:     "test message\n\nTest Source\\_\\*\\{\\}\\[\\]\\(\\)\\#\\+\\-\\.\\!\\~\\`\\>\\=\\|\n\n[🔗Source Link\\_\\*\\{\\}\\[\\]\\(\\)\\#\\+\\-\\.\\!\\~\\`\\>\\=\\|](https://t.me/test/123)",
			expectedEntities: []*client.TextEntity{},
			setup: func(t *testing.T) *Service {
				telegramRepo := mocks.NewTelegramRepo(t)
				storageService := mocks.NewStorageService(t)
				messageService := mocks.NewMessageService(t)

				// Mock для addSources - sign
				telegramRepo.EXPECT().ParseTextEntities(&client.ParseTextEntitiesRequest{
					Text: "Test Source\\_\\*\\{\\}\\[\\]\\(\\)\\#\\+\\-\\.\\!\\~\\`\\>\\=\\|",
					ParseMode: &client.TextParseModeMarkdown{
						Version: 2,
					},
				}).Return(&client.FormattedText{
					Text:     "Test Source\\_\\*\\{\\}\\[\\]\\(\\)\\#\\+\\-\\.\\!\\~\\`\\>\\=\\|",
					Entities: nil,
				}, nil)

				// Mock для addSources - link
				telegramRepo.EXPECT().GetMessageLink(&client.GetMessageLinkRequest{
					ChatId:    10002,
					MessageId: 123,
					ForAlbum:  false,
				}).Return(&client.MessageLink{
					Link: "https://t.me/test/123",
				}, nil)

				telegramRepo.EXPECT().ParseTextEntities(&client.ParseTextEntitiesRequest{
					Text: "[🔗Source Link\\_\\*\\{\\}\\[\\]\\(\\)\\#\\+\\-\\.\\!\\~\\`\\>\\=\\|](https://t.me/test/123)",
					ParseMode: &client.TextParseModeMarkdown{
						Version: 2,
					},
				}).Return(&client.FormattedText{
					Text:     "[🔗Source Link\\_\\*\\{\\}\\[\\]\\(\\)\\#\\+\\-\\.\\!\\~\\`\\>\\=\\|](https://t.me/test/123)",
					Entities: []*client.TextEntity{},
				}, nil)

				return New(telegramRepo, storageService, messageService)
			},
		},
		{
			name: "without sources",
			formattedText: &client.FormattedText{
				Text:     "test message",
				Entities: []*client.TextEntity{},
			},
			withSources: false,
			src: &client.Message{
				ChatId:       10001,
				Id:           123,
				MediaAlbumId: 0,
			},
			dstChatId:        0, // не используется
			expectedText:     "test message",
			expectedEntities: []*client.TextEntity{},
			setup: func(t *testing.T) *Service {
				telegramRepo := mocks.NewTelegramRepo(t)
				storageService := mocks.NewStorageService(t)
				messageService := mocks.NewMessageService(t)
				return New(telegramRepo, storageService, messageService)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			transformService := test.setup(t)
			transformService.Transform(test.formattedText, test.withSources, test.src, test.dstChatId)

			assert.Equal(t, test.expectedText, test.formattedText.Text)
			assert.Equal(t, test.expectedEntities, test.formattedText.Entities)
		})
	}
}

func TestTransformService_replaceMyselfLinks(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		formattedText    *client.FormattedText
		srcChatId        int64
		dstChatId        int64
		expectedEntities []*client.TextEntity
		setup            func(t *testing.T) *Service
		expectedError    error
	}{
		{
			name: "destination not found",
			formattedText: &client.FormattedText{
				Text: "test",
				Entities: []*client.TextEntity{
					{
						Type: &client.TextEntityTypeTextUrl{
							Url: "https://t.me/test/123",
						},
					},
				},
			},
			srcChatId: 10100,
			dstChatId: 99999, // не существует в config.yml
			expectedEntities: []*client.TextEntity{
				{
					Type: &client.TextEntityTypeTextUrl{
						Url: "https://t.me/test/123",
					},
				},
			},
			setup: func(t *testing.T) *Service {
				telegramRepo := mocks.NewTelegramRepo(t)
				storageService := mocks.NewStorageService(t)
				return New(telegramRepo, storageService, nil)
			},
			expectedError: log.NewError("dstChatId not found"),
		},
		{
			name: "replace myself links disabled",
			formattedText: &client.FormattedText{
				Text: "test",
				Entities: []*client.TextEntity{
					{
						Type: &client.TextEntityTypeTextUrl{
							Url: "https://t.me/test/123",
						},
					},
				},
			},
			srcChatId: 10100,
			dstChatId: 10115,
			expectedEntities: []*client.TextEntity{
				{
					Type: &client.TextEntityTypeTextUrl{
						Url: "https://t.me/test/123",
					},
				},
			},
			setup: func(t *testing.T) *Service {
				telegramRepo := mocks.NewTelegramRepo(t)
				storageService := mocks.NewStorageService(t)
				return New(telegramRepo, storageService, nil)
			},
			expectedError: log.NewError("Run is disabled"),
		},
		{
			name: "no text url entities",
			formattedText: &client.FormattedText{
				Text: "test",
				Entities: []*client.TextEntity{
					{
						Type: &client.TextEntityTypeBold{},
					},
				},
			},
			srcChatId: 10100,
			dstChatId: 10114,
			expectedEntities: []*client.TextEntity{
				{
					Type: &client.TextEntityTypeBold{},
				},
			},
			setup: func(t *testing.T) *Service {
				telegramRepo := mocks.NewTelegramRepo(t)
				storageService := mocks.NewStorageService(t)
				return New(telegramRepo, storageService, nil)
			},
		},
		{
			name: "get message link info error",
			formattedText: &client.FormattedText{
				Text: "test",
				Entities: []*client.TextEntity{
					{
						Type: &client.TextEntityTypeTextUrl{
							Url: "https://t.me/test/123",
						},
					},
				},
			},
			srcChatId: 10100,
			dstChatId: 10114,
			expectedEntities: []*client.TextEntity{
				{
					Type: &client.TextEntityTypeTextUrl{
						Url: "https://t.me/test/123",
					},
				},
			},
			setup: func(t *testing.T) *Service {
				telegramRepo := mocks.NewTelegramRepo(t)
				storageService := mocks.NewStorageService(t)
				telegramRepo.EXPECT().GetMessageLinkInfo(&client.GetMessageLinkInfoRequest{
					Url: "https://t.me/test/123",
				}).Return(nil, errors.New("link info error"))
				return New(telegramRepo, storageService, nil)
			},
			expectedError: log.NewError("link info error"),
		},
		{
			name: "message not from source chat",
			formattedText: &client.FormattedText{
				Text: "test",
				Entities: []*client.TextEntity{
					{
						Type: &client.TextEntityTypeTextUrl{
							Url: "https://t.me/test/123",
						},
					},
				},
			},
			srcChatId: 10100,
			dstChatId: 10114,
			expectedEntities: []*client.TextEntity{
				{
					Type: &client.TextEntityTypeTextUrl{
						Url: "https://t.me/test/123",
					},
				},
			},
			setup: func(t *testing.T) *Service {
				telegramRepo := mocks.NewTelegramRepo(t)
				storageService := mocks.NewStorageService(t)
				telegramRepo.EXPECT().GetMessageLinkInfo(&client.GetMessageLinkInfoRequest{
					Url: "https://t.me/test/123",
				}).Return(&client.MessageLinkInfo{
					Message: &client.Message{
						ChatId: 99999, // другой чат
						Id:     123,
					},
				}, nil)
				return New(telegramRepo, storageService, nil)
			},
		},
		{
			name: "successful link replacement",
			formattedText: &client.FormattedText{
				Text: "test",
				Entities: []*client.TextEntity{
					{
						Type: &client.TextEntityTypeTextUrl{
							Url: "https://t.me/test/123",
						},
					},
				},
			},
			srcChatId: 10100,
			dstChatId: 10114,
			expectedEntities: []*client.TextEntity{
				{
					Type: &client.TextEntityTypeTextUrl{
						Url: "https://t.me/newchat/456",
					},
				},
			},
			setup: func(t *testing.T) *Service {
				telegramRepo := mocks.NewTelegramRepo(t)
				storageService := mocks.NewStorageService(t)
				telegramRepo.EXPECT().GetMessageLinkInfo(&client.GetMessageLinkInfoRequest{
					Url: "https://t.me/test/123",
				}).Return(&client.MessageLinkInfo{
					Message: &client.Message{
						ChatId: 10100,
						Id:     123,
					},
				}, nil)

				storageService.EXPECT().GetCopiedMessageIds("10100:123").Return([]string{
					"rule1:10114:789",
				})

				storageService.EXPECT().GetNewMessageId(int64(10114), int64(789)).Return(int64(456))

				telegramRepo.EXPECT().GetMessageLink(&client.GetMessageLinkRequest{
					ChatId:    10114,
					MessageId: 456,
				}).Return(&client.MessageLink{
					Link: "https://t.me/newchat/456",
				}, nil)
				return New(telegramRepo, storageService, nil)
			},
		},
		{
			name: "no copied messages found - delete external",
			formattedText: &client.FormattedText{
				Text: "test",
				Entities: []*client.TextEntity{
					{
						Type: &client.TextEntityTypeTextUrl{
							Url: "https://t.me/test/123",
						},
					},
				},
			},
			srcChatId: 10100,
			dstChatId: 10114,
			expectedEntities: []*client.TextEntity{
				{
					Type: &client.TextEntityTypeStrikethrough{},
				},
			},
			setup: func(t *testing.T) *Service {
				telegramRepo := mocks.NewTelegramRepo(t)
				storageService := mocks.NewStorageService(t)
				telegramRepo.EXPECT().GetMessageLinkInfo(&client.GetMessageLinkInfoRequest{
					Url: "https://t.me/test/123",
				}).Return(&client.MessageLinkInfo{
					Message: &client.Message{
						ChatId: 10100,
						Id:     123,
					},
				}, nil)

				storageService.EXPECT().GetCopiedMessageIds("10100:123").Return([]string{})
				return New(telegramRepo, storageService, nil)
			},
		},
		{
			name: "no copied messages found - no delete external",
			formattedText: &client.FormattedText{
				Text: "test",
				Entities: []*client.TextEntity{
					{
						Type: &client.TextEntityTypeTextUrl{
							Url: "https://t.me/test/123",
						},
					},
				},
			},
			srcChatId: 10100,
			dstChatId: 10116,
			expectedEntities: []*client.TextEntity{
				{
					Type: &client.TextEntityTypeTextUrl{
						Url: "https://t.me/test/123",
					},
				},
			},
			setup: func(t *testing.T) *Service {
				telegramRepo := mocks.NewTelegramRepo(t)
				storageService := mocks.NewStorageService(t)
				telegramRepo.EXPECT().GetMessageLinkInfo(&client.GetMessageLinkInfoRequest{
					Url: "https://t.me/test/123",
				}).Return(&client.MessageLinkInfo{
					Message: &client.Message{
						ChatId: 10100,
						Id:     123,
					},
				}, nil)

				storageService.EXPECT().GetCopiedMessageIds("10100:123").Return([]string{})
				return New(telegramRepo, storageService, nil)
			},
		},
		{
			name: "get new message id returns zero",
			formattedText: &client.FormattedText{
				Text: "test",
				Entities: []*client.TextEntity{
					{
						Type: &client.TextEntityTypeTextUrl{
							Url: "https://t.me/test/123",
						},
					},
				},
			},
			srcChatId: 10100,
			dstChatId: 10114,
			expectedEntities: []*client.TextEntity{
				{
					Type: &client.TextEntityTypeTextUrl{
						Url: "https://t.me/test/123",
					},
				},
			},
			setup: func(t *testing.T) *Service {
				telegramRepo := mocks.NewTelegramRepo(t)
				storageService := mocks.NewStorageService(t)
				telegramRepo.EXPECT().GetMessageLinkInfo(&client.GetMessageLinkInfoRequest{
					Url: "https://t.me/test/123",
				}).Return(&client.MessageLinkInfo{
					Message: &client.Message{
						ChatId: 10100,
						Id:     123,
					},
				}, nil)

				storageService.EXPECT().GetCopiedMessageIds("10100:123").Return([]string{
					"rule1:10114:789",
				})

				storageService.EXPECT().GetNewMessageId(int64(10114), int64(789)).Return(int64(0))
				return New(telegramRepo, storageService, nil)
			},
			expectedError: log.NewError("GetNewMessageId return 0"),
		},
		{
			name: "get message link error",
			formattedText: &client.FormattedText{
				Text: "test",
				Entities: []*client.TextEntity{
					{
						Type: &client.TextEntityTypeTextUrl{
							Url: "https://t.me/test/123",
						},
					},
				},
			},
			srcChatId: 10100,
			dstChatId: 10114,
			expectedEntities: []*client.TextEntity{
				{
					Type: &client.TextEntityTypeTextUrl{
						Url: "https://t.me/test/123",
					},
				},
			},
			setup: func(t *testing.T) *Service {
				telegramRepo := mocks.NewTelegramRepo(t)
				storageService := mocks.NewStorageService(t)
				telegramRepo.EXPECT().GetMessageLinkInfo(&client.GetMessageLinkInfoRequest{
					Url: "https://t.me/test/123",
				}).Return(&client.MessageLinkInfo{
					Message: &client.Message{
						ChatId: 10100,
						Id:     123,
					},
				}, nil)

				storageService.EXPECT().GetCopiedMessageIds("10100:123").Return([]string{
					"rule1:10114:789",
				})

				storageService.EXPECT().GetNewMessageId(int64(10114), int64(789)).Return(int64(456))

				telegramRepo.EXPECT().GetMessageLink(&client.GetMessageLinkRequest{
					ChatId:    10114,
					MessageId: 456,
				}).Return(nil, errors.New("message link error"))
				return New(telegramRepo, storageService, nil)
			},
			expectedError: log.NewError("message link error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var transformService *Service
			spylogHandler := spylog.GetModuleLogHandler("service.transform", t.Name(), func() {
				transformService = test.setup(t)
			})

			transformService.replaceMyselfLinks(test.formattedText, test.srcChatId, test.dstChatId)

			if test.expectedError != nil {
				records := spylogHandler.GetRecords()
				require.Equal(t, len(records), 1)
				assert.Equal(t, slog.LevelError, records[0].Level)
				assert.Equal(t, test.expectedError.Error(), records[0].Message)
			}

			assert.Equal(t, test.expectedEntities, test.formattedText.Entities)
		})
	}
}

func TestTransformService_replaceFragments(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		formattedText *client.FormattedText
		dstChatId     int64
		expectedText  string
		expectedError error
	}{
		{
			name: "destination not found",
			formattedText: &client.FormattedText{
				Text: "some text",
			},
			dstChatId:     99999, // не существует в config.yml
			expectedText:  "some text",
			expectedError: log.NewError("destination not found"),
		},
		{
			name: "no replace fragments",
			formattedText: &client.FormattedText{
				Text: "some text",
			},
			dstChatId:    10113,
			expectedText: "some text",
		},
		{
			name: "single replacement",
			formattedText: &client.FormattedText{
				Text: "hello world",
			},
			dstChatId:    10110,
			expectedText: "12345 67890",
		},
		{
			name: "case insensitive replacement",
			formattedText: &client.FormattedText{
				Text: "This is a test and TEST and Test",
			},
			dstChatId:    10111,
			expectedText: "This is a Тест and Тест and Тест",
		},
		{
			name: "multiple occurrences",
			formattedText: &client.FormattedText{
				Text: "old text with old values and old data",
			},
			dstChatId:    10112,
			expectedText: "new text with new values and new data",
		},
		{
			name: "no matches",
			formattedText: &client.FormattedText{
				Text: "some random text",
			},
			dstChatId:    10110,
			expectedText: "some random text",
		},
		{
			name: "partial word match",
			formattedText: &client.FormattedText{
				Text: "hello? and world!",
			},
			dstChatId:    10110,
			expectedText: "12345? and 67890!",
		},
		{
			name: "empty text",
			formattedText: &client.FormattedText{
				Text: "",
			},
			dstChatId:    10110,
			expectedText: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var transformService *Service
			spylogHandler := spylog.GetModuleLogHandler("service.transform", t.Name(), func() {
				transformService = New(nil, nil, nil)
			})

			transformService.replaceFragments(test.formattedText, test.dstChatId)

			if test.expectedError != nil {
				records := spylogHandler.GetRecords()
				require.Equal(t, len(records), 1)
				assert.Equal(t, slog.LevelError, records[0].Level)
				assert.Equal(t, test.expectedError.Error(), records[0].Message)
			}

			assert.Equal(t, test.expectedText, test.formattedText.Text)
		})
	}
}

func TestTransformService_addAutoAnswer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		formattedText    *client.FormattedText
		src              *client.Message
		expectedText     string
		expectedEntities []*client.TextEntity
		setup            func(t *testing.T, src *client.Message) *Service
		expectedError    error
	}{
		{
			name:          "source not found",
			formattedText: &client.FormattedText{},
			src: &client.Message{
				ChatId: 99999, // не существует в config.yml
				Id:     123,
			},
			expectedText:     "",
			expectedEntities: nil,
			setup: func(t *testing.T, src *client.Message) *Service {
				telegramRepo := mocks.NewTelegramRepo(t)
				messageService := mocks.NewMessageService(t)
				return New(telegramRepo, nil, messageService)
			},
			expectedError: log.NewError("source not found"),
		},
		{
			name:          "auto answer disabled",
			formattedText: &client.FormattedText{},
			src: &client.Message{
				ChatId: 10107,
				Id:     123,
			},
			expectedText:     "",
			expectedEntities: nil,
			setup: func(t *testing.T, src *client.Message) *Service {
				telegramRepo := mocks.NewTelegramRepo(t)
				messageService := mocks.NewMessageService(t)
				return New(telegramRepo, nil, messageService)
			},
			expectedError: log.NewError("source.AutoAnswer is false"),
		},
		{
			name:          "reply markup data is empty",
			formattedText: &client.FormattedText{},
			src: &client.Message{
				ChatId: 10106,
				Id:     123,
			},
			expectedText:     "",
			expectedEntities: nil,
			setup: func(t *testing.T, src *client.Message) *Service {
				telegramRepo := mocks.NewTelegramRepo(t)
				messageService := mocks.NewMessageService(t)
				messageService.EXPECT().GetReplyMarkupData(src).Return([]byte{})
				return New(telegramRepo, nil, messageService)
			},
			expectedError: log.NewError("reply markup data is empty"),
		},
		{
			name:          "get callback query answer error",
			formattedText: &client.FormattedText{},
			src: &client.Message{
				ChatId: 10106,
				Id:     123,
			},
			expectedText:     "",
			expectedEntities: nil,
			setup: func(t *testing.T, src *client.Message) *Service {
				telegramRepo := mocks.NewTelegramRepo(t)
				messageService := mocks.NewMessageService(t)
				replyMarkupData := []byte("test_data")
				messageService.EXPECT().GetReplyMarkupData(src).Return(replyMarkupData)
				telegramRepo.EXPECT().GetCallbackQueryAnswer(&client.GetCallbackQueryAnswerRequest{
					ChatId:    src.ChatId,
					MessageId: src.Id,
					Payload:   &client.CallbackQueryPayloadData{Data: replyMarkupData},
				}).Return(nil, errors.New("callback query error"))
				return New(telegramRepo, nil, messageService)
			},
			expectedError: log.NewError("callback query error"),
		},
		{
			name: "successful auto answer",
			formattedText: &client.FormattedText{
				Text:     "existing text",
				Entities: []*client.TextEntity{},
			},
			src: &client.Message{
				ChatId: 10106,
				Id:     123,
			},
			expectedText:     "existing text\n\n\\*Auto Answer\\*",
			expectedEntities: []*client.TextEntity{},
			setup: func(t *testing.T, src *client.Message) *Service {
				telegramRepo := mocks.NewTelegramRepo(t)
				messageService := mocks.NewMessageService(t)
				replyMarkupData := []byte("test_data")
				messageService.EXPECT().GetReplyMarkupData(src).Return(replyMarkupData)
				telegramRepo.EXPECT().GetCallbackQueryAnswer(&client.GetCallbackQueryAnswerRequest{
					ChatId:    src.ChatId,
					MessageId: src.Id,
					Payload:   &client.CallbackQueryPayloadData{Data: replyMarkupData},
				}).Return(&client.CallbackQueryAnswer{
					Text: "*Auto Answer*",
				}, nil)
				telegramRepo.EXPECT().ParseTextEntities(&client.ParseTextEntitiesRequest{
					Text: "\\*Auto Answer\\*",
					ParseMode: &client.TextParseModeMarkdown{
						Version: 2,
					},
				}).Return(&client.FormattedText{
					Text:     "\\*Auto Answer\\*",
					Entities: nil,
				}, nil)
				return New(telegramRepo, nil, messageService)
			},
		},
		{
			name:          "empty formatted text",
			formattedText: &client.FormattedText{},
			src: &client.Message{
				ChatId: 10106,
				Id:     123,
			},
			expectedText:     "\\*Auto Answer\\*",
			expectedEntities: nil,
			setup: func(t *testing.T, src *client.Message) *Service {
				telegramRepo := mocks.NewTelegramRepo(t)
				messageService := mocks.NewMessageService(t)
				replyMarkupData := []byte("test_data")
				messageService.EXPECT().GetReplyMarkupData(src).Return(replyMarkupData)
				telegramRepo.EXPECT().GetCallbackQueryAnswer(&client.GetCallbackQueryAnswerRequest{
					ChatId:    src.ChatId,
					MessageId: src.Id,
					Payload:   &client.CallbackQueryPayloadData{Data: replyMarkupData},
				}).Return(&client.CallbackQueryAnswer{
					Text: "*Auto Answer*",
				}, nil)
				telegramRepo.EXPECT().ParseTextEntities(&client.ParseTextEntitiesRequest{
					Text: "\\*Auto Answer\\*",
					ParseMode: &client.TextParseModeMarkdown{
						Version: 2,
					},
				}).Return(&client.FormattedText{
					Text:     "\\*Auto Answer\\*",
					Entities: nil,
				}, nil)
				return New(telegramRepo, nil, messageService)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var transformService *Service
			spylogHandler := spylog.GetModuleLogHandler("service.transform", t.Name(), func() {
				transformService = test.setup(t, test.src)
			})

			transformService.addAutoAnswer(test.formattedText, test.src)

			if test.expectedError != nil {
				records := spylogHandler.GetRecords()
				require.Equal(t, len(records), 1)
				assert.Equal(t, slog.LevelError, records[0].Level)
				assert.Equal(t, test.expectedError.Error(), records[0].Message)
			}

			assert.Equal(t, test.expectedText, test.formattedText.Text)
			assert.Equal(t, test.expectedEntities, test.formattedText.Entities)
		})
	}
}

func TestTransformService_addSources(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		formattedText    *client.FormattedText
		src              *client.Message
		dstChatId        int64
		expectedText     string
		expectedEntities []*client.TextEntity
		setup            func(t *testing.T, src *client.Message) *Service
		expectedError    error
	}{
		{
			name:          "source not found",
			formattedText: &client.FormattedText{},
			src: &client.Message{
				ChatId: 99999, // не существует в config.yml
				Id:     123,
			},
			dstChatId:        0, // не используется
			expectedText:     "",
			expectedEntities: nil,
			setup: func(t *testing.T, src *client.Message) *Service {
				telegramRepo := mocks.NewTelegramRepo(t)
				return New(telegramRepo, nil, nil)
			},
			expectedError: log.NewError("source not found"),
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
			setup: func(t *testing.T, src *client.Message) *Service {
				telegramRepo := mocks.NewTelegramRepo(t)
				telegramRepo.EXPECT().ParseTextEntities(&client.ParseTextEntitiesRequest{
					Text: "Test Source",
					ParseMode: &client.TextParseModeMarkdown{
						Version: 2,
					},
				}).Return(&client.FormattedText{
					Text:     "Test Source",
					Entities: nil,
				}, nil)
				return New(telegramRepo, nil, nil)
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
			setup: func(t *testing.T, src *client.Message) *Service {
				telegramRepo := mocks.NewTelegramRepo(t)
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
				return New(telegramRepo, nil, nil)
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
			setup: func(t *testing.T, src *client.Message) *Service {
				telegramRepo := mocks.NewTelegramRepo(t)
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
				return New(telegramRepo, nil, nil)
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
			setup: func(t *testing.T, src *client.Message) *Service {
				telegramRepo := mocks.NewTelegramRepo(t)
				return New(telegramRepo, nil, nil)
			},
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
			setup: func(t *testing.T, src *client.Message) *Service {
				telegramRepo := mocks.NewTelegramRepo(t)
				return New(telegramRepo, nil, nil)
			},
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
			setup: func(t *testing.T, src *client.Message) *Service {
				telegramRepo := mocks.NewTelegramRepo(t)
				telegramRepo.EXPECT().GetMessageLink(&client.GetMessageLinkRequest{
					ChatId:    src.ChatId,
					MessageId: src.Id,
					ForAlbum:  src.MediaAlbumId != 0,
				}).Return(nil, errors.New("get message link error"))
				return New(telegramRepo, nil, nil)
			},
			expectedError: log.NewError("get message link error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var transformService *Service
			spylogHandler := spylog.GetModuleLogHandler("service.transform", t.Name(), func() {
				transformService = test.setup(t, test.src)
			})

			transformService.addSources(test.formattedText, test.src, test.dstChatId)

			if test.expectedError != nil {
				records := spylogHandler.GetRecords()
				require.Equal(t, len(records), 1)
				assert.Equal(t, slog.LevelError, records[0].Level)
				assert.Equal(t, test.expectedError.Error(), records[0].Message)
			}

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

			var transformService *Service
			spylogHandler := spylog.GetModuleLogHandler("service.transform", t.Name(), func() {
				transformService = New(telegramRepo, nil, nil)
			})

			transformService.addText(test.formattedText, test.text)

			if test.expectedError != nil {
				records := spylogHandler.GetRecords()
				require.Equal(t, len(records), 1)
				assert.Equal(t, slog.LevelError, records[0].Level)
				assert.Equal(t, test.expectedError.Error(), records[0].Message)
			}

			assert.Equal(t, test.expectedText, test.formattedText.Text)
			assert.Equal(t, test.expectedEntities, test.formattedText.Entities)
		})
	}
}

func TestEscapeMarkdown(t *testing.T) {
	t.Parallel()

	s := "_ * ( ) ~ ` > # + = | { } . ! \\[ \\] \\-"
	a := strings.Split(s, " ")
	for _, v := range a {
		assert.Equal(t, "\\"+v, escapeMarkdown(v))
	}
}
