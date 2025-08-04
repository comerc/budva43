package transform

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zelenin/go-tdlib/client"

	"github.com/comerc/budva43/app/config"
	"github.com/comerc/budva43/app/domain"
	"github.com/comerc/budva43/app/engine_config"
	"github.com/comerc/budva43/app/log"
	"github.com/comerc/budva43/app/testing/spylog"
	"github.com/comerc/budva43/service/transform/mocks"
)

// data for service.transform - -101xx

func TestMain(m *testing.M) {
	initDestinations := func([]domain.ChatId) {}
	engine_config.Reload(initDestinations)
	os.Exit(m.Run())
}

func Test(t *testing.T) {
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
			name: "with_sources_and_markdown",
			formattedText: &client.FormattedText{
				Text:     "test message",
				Entities: []*client.TextEntity{},
			},
			withSources: true,
			src: &client.Message{
				ChatId:       -10121,
				Id:           123,
				MediaAlbumId: 0,
			},
			dstChatId:        -10123,
			expectedText:     "test message\n\n*Test Source*\n\n[🔗*Source Link*](https://t.me/test/123)",
			expectedEntities: []*client.TextEntity{},
			setup: func(t *testing.T) *Service {
				telegramRepo := mocks.NewTelegramRepo(t)

				// Mock для addSourceSign
				telegramRepo.EXPECT().ParseTextEntities(&client.ParseTextEntitiesRequest{
					Text: "*Test Source*",
					ParseMode: &client.TextParseModeMarkdown{
						Version: 2,
					},
				}).Return(&client.FormattedText{
					Text:     "*Test Source*",
					Entities: []*client.TextEntity{},
				}, nil)

				// Mock для addSourceLink
				telegramRepo.EXPECT().GetMessageLink(&client.GetMessageLinkRequest{
					ChatId:    -10121,
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

				return New(telegramRepo, nil, nil)
			},
		},
		{
			name: "with_sources_and_escaped_markdown",
			formattedText: &client.FormattedText{
				Text:     "test message",
				Entities: []*client.TextEntity{},
			},
			withSources: true,
			src: &client.Message{
				ChatId:       -10122,
				Id:           123,
				MediaAlbumId: 0,
			},
			dstChatId:        -10123,
			expectedText:     "test message\n\nTest Source\\_\\*\\{\\}\\[\\]\\(\\)\\#\\+\\-\\.\\!\\~\\`\\>\\=\\|\n\n[🔗Source Link\\_\\*\\{\\}\\[\\]\\(\\)\\#\\+\\-\\.\\!\\~\\`\\>\\=\\|](https://t.me/test/123)",
			expectedEntities: []*client.TextEntity{},
			setup: func(t *testing.T) *Service {
				telegramRepo := mocks.NewTelegramRepo(t)

				// Mock для addSourceSign
				telegramRepo.EXPECT().ParseTextEntities(&client.ParseTextEntitiesRequest{
					Text: "Test Source\\_\\*\\{\\}\\[\\]\\(\\)\\#\\+\\-\\.\\!\\~\\`\\>\\=\\|",
					ParseMode: &client.TextParseModeMarkdown{
						Version: 2,
					},
				}).Return(&client.FormattedText{
					Text:     "Test Source\\_\\*\\{\\}\\[\\]\\(\\)\\#\\+\\-\\.\\!\\~\\`\\>\\=\\|",
					Entities: nil,
				}, nil)

				// Mock для addSourceLink
				telegramRepo.EXPECT().GetMessageLink(&client.GetMessageLinkRequest{
					ChatId:    -10122,
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

				return New(telegramRepo, nil, nil)
			},
		},
		{
			name: "without_sources",
			formattedText: &client.FormattedText{
				Text:     "test message",
				Entities: []*client.TextEntity{},
			},
			withSources: false,
			src: &client.Message{
				ChatId:       -10121,
				Id:           123,
				MediaAlbumId: 0,
			},
			dstChatId:        0, // не используется
			expectedText:     "test message",
			expectedEntities: []*client.TextEntity{},
			setup: func(t *testing.T) *Service {
				return New(nil, nil, nil)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			transformService := test.setup(t)
			transformService.Transform(test.formattedText, test.withSources, test.src, test.dstChatId, 0, config.Engine)

			assert.Equal(t, test.expectedText, test.formattedText.Text)
			assert.Equal(t, test.expectedEntities, test.formattedText.Entities)
		})
	}
}

func Test_replaceMyselfLinks(t *testing.T) {
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
			name: "destination_not_found",
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
			srcChatId: -10100,
			dstChatId: -10199, // не существует в config.yml
			expectedEntities: []*client.TextEntity{
				{
					Type: &client.TextEntityTypeTextUrl{
						Url: "https://t.me/test/123",
					},
				},
			},
			setup: func(t *testing.T) *Service {
				return New(nil, nil, nil)
			},
			expectedError: log.NewError("destination not found"),
		},
		{
			name: "replace_myself_links_is_nil",
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
			srcChatId: -10100,
			dstChatId: -10113, // destination без replace-myself-links
			expectedEntities: []*client.TextEntity{
				{
					Type: &client.TextEntityTypeTextUrl{
						Url: "https://t.me/test/123",
					},
				},
			},
			setup: func(t *testing.T) *Service {
				return New(nil, nil, nil)
			},
			expectedError: log.NewError("replaceMyselfLinks is nil"),
		},
		{
			name: "replace_myself_links_is_empty-both_false",
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
			srcChatId: -10100,
			dstChatId: -10117, // destination с пустым replace-myself-links (для теста empty)
			expectedEntities: []*client.TextEntity{
				{
					Type: &client.TextEntityTypeTextUrl{
						Url: "https://t.me/test/123",
					},
				},
			},
			setup: func(t *testing.T) *Service {
				return New(nil, nil, nil)
			},
			expectedError: log.NewError("replaceMyselfLinks.Run is false"),
		},
		{
			name: "get_chat_error",
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
			srcChatId: -10100,
			dstChatId: -10114,
			expectedEntities: []*client.TextEntity{
				{
					Type: &client.TextEntityTypeTextUrl{
						Url: "https://t.me/test/123",
					},
				},
			},
			setup: func(t *testing.T) *Service {
				telegramRepo := mocks.NewTelegramRepo(t)
				telegramRepo.EXPECT().GetChat(&client.GetChatRequest{
					ChatId: int64(-10100),
				}).Return(nil, errors.New("get chat error"))
				return New(telegramRepo, nil, nil)
			},
		},
		{
			name: "no_text_url_entities",
			formattedText: &client.FormattedText{
				Text: "test",
				Entities: []*client.TextEntity{
					{
						Type: &client.TextEntityTypeBold{},
					},
				},
			},
			srcChatId: -10100,
			dstChatId: -10114,
			expectedEntities: []*client.TextEntity{
				{
					Type: &client.TextEntityTypeBold{},
				},
			},
			setup: func(t *testing.T) *Service {
				telegramRepo := mocks.NewTelegramRepo(t)
				telegramRepo.EXPECT().GetChat(&client.GetChatRequest{
					ChatId: int64(-10100),
				}).Return(&client.Chat{
					Type: &client.ChatTypeSupergroup{},
				}, nil)
				return New(telegramRepo, nil, nil)
			},
		},

		{
			name: "get_message_link_info_error",
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
			srcChatId: -10100,
			dstChatId: -10114,
			expectedEntities: []*client.TextEntity{
				{
					Type: &client.TextEntityTypeTextUrl{
						Url: "https://t.me/test/123",
					},
				},
			},
			setup: func(t *testing.T) *Service {
				telegramRepo := mocks.NewTelegramRepo(t)
				telegramRepo.EXPECT().GetChat(&client.GetChatRequest{
					ChatId: int64(-10100),
				}).Return(&client.Chat{
					Type: &client.ChatTypeSupergroup{},
				}, nil)
				telegramRepo.EXPECT().GetMessageLinkInfo(&client.GetMessageLinkInfoRequest{
					Url: "https://t.me/test/123",
				}).Return(nil, errors.New("message link info error"))
				return New(telegramRepo, nil, nil)
			},
		},
		{
			name: "get_message_link_info_returns_nil_message",
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
			srcChatId: -10100,
			dstChatId: -10114,
			expectedEntities: []*client.TextEntity{
				{
					Type: &client.TextEntityTypeTextUrl{
						Url: "https://t.me/test/123",
					},
				},
			},
			setup: func(t *testing.T) *Service {
				telegramRepo := mocks.NewTelegramRepo(t)
				telegramRepo.EXPECT().GetChat(&client.GetChatRequest{
					ChatId: int64(-10100),
				}).Return(&client.Chat{
					Type: &client.ChatTypeSupergroup{},
				}, nil)
				telegramRepo.EXPECT().GetMessageLinkInfo(&client.GetMessageLinkInfoRequest{
					Url: "https://t.me/test/123",
				}).Return(&client.MessageLinkInfo{
					Message: nil,
				}, nil)
				return New(telegramRepo, nil, nil)
			},
		},
		{
			name: "message_not_from_source_chat",
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
			srcChatId: -10100,
			dstChatId: -10114,
			expectedEntities: []*client.TextEntity{
				{
					Type: &client.TextEntityTypeTextUrl{
						Url: "https://t.me/test/123",
					},
				},
			},
			setup: func(t *testing.T) *Service {
				telegramRepo := mocks.NewTelegramRepo(t)
				telegramRepo.EXPECT().GetChat(&client.GetChatRequest{
					ChatId: int64(-10100),
				}).Return(&client.Chat{
					Type: &client.ChatTypeSupergroup{},
				}, nil)
				telegramRepo.EXPECT().GetMessageLinkInfo(&client.GetMessageLinkInfoRequest{
					Url: "https://t.me/test/123",
				}).Return(&client.MessageLinkInfo{
					Message: &client.Message{
						ChatId: -10118, // другой источник (не srcChatId -10100)
						Id:     123,
					},
				}, nil)
				return New(telegramRepo, nil, nil)
			},
		},
		{
			name: "no_copied_messages_found",
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
			srcChatId: -10100,
			dstChatId: -10116, // destination с run=true и deleteExternal=false
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
				telegramRepo.EXPECT().GetChat(&client.GetChatRequest{
					ChatId: int64(-10100),
				}).Return(&client.Chat{
					Type: &client.ChatTypeSupergroup{},
				}, nil)
				telegramRepo.EXPECT().GetMessageLinkInfo(&client.GetMessageLinkInfoRequest{
					Url: "https://t.me/test/123",
				}).Return(&client.MessageLinkInfo{
					Message: &client.Message{
						ChatId: -10100,
						Id:     123,
					},
				}, nil)
				storageService.EXPECT().GetCopiedMessageIds(int64(-10100), int64(123)).Return([]string{})
				return New(telegramRepo, storageService, nil)
			},
		},
		{
			name: "tmp_message_id_zero",
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
			srcChatId: -10100,
			dstChatId: -10114,
			expectedEntities: []*client.TextEntity{
				{
					Type: &client.TextEntityTypeStrikethrough{},
				},
			},
			setup: func(t *testing.T) *Service {
				telegramRepo := mocks.NewTelegramRepo(t)
				storageService := mocks.NewStorageService(t)
				telegramRepo.EXPECT().GetChat(&client.GetChatRequest{
					ChatId: int64(-10100),
				}).Return(&client.Chat{
					Type: &client.ChatTypeSupergroup{},
				}, nil)
				telegramRepo.EXPECT().GetMessageLinkInfo(&client.GetMessageLinkInfoRequest{
					Url: "https://t.me/test/123",
				}).Return(&client.MessageLinkInfo{
					Message: &client.Message{
						ChatId: -10100,
						Id:     123,
					},
				}, nil)
				storageService.EXPECT().GetCopiedMessageIds(int64(-10100), int64(123)).Return([]string{
					"rule1:-10119:789", // другое назначение (не dstChatId -10114)
				})
				return New(telegramRepo, storageService, nil)
			},
		},
		{
			name: "new_message_id_zero",
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
			srcChatId: -10100,
			dstChatId: -10114,
			expectedEntities: []*client.TextEntity{
				{
					Type: &client.TextEntityTypeStrikethrough{},
				},
			},
			setup: func(t *testing.T) *Service {
				telegramRepo := mocks.NewTelegramRepo(t)
				storageService := mocks.NewStorageService(t)
				telegramRepo.EXPECT().GetChat(&client.GetChatRequest{
					ChatId: int64(-10100),
				}).Return(&client.Chat{
					Type: &client.ChatTypeSupergroup{},
				}, nil)
				telegramRepo.EXPECT().GetMessageLinkInfo(&client.GetMessageLinkInfoRequest{
					Url: "https://t.me/test/123",
				}).Return(&client.MessageLinkInfo{
					Message: &client.Message{
						ChatId: -10100,
						Id:     123,
					},
				}, nil)
				storageService.EXPECT().GetCopiedMessageIds(int64(-10100), int64(123)).Return([]string{
					"rule1:-10114:789",
				})
				storageService.EXPECT().GetNewMessageId(int64(-10114), int64(789)).Return(int64(0))
				return New(telegramRepo, storageService, nil)
			},
		},
		{
			name: "get_message_link_error",
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
			srcChatId: -10100,
			dstChatId: -10114,
			expectedEntities: []*client.TextEntity{
				{
					Type: &client.TextEntityTypeStrikethrough{},
				},
			},
			setup: func(t *testing.T) *Service {
				telegramRepo := mocks.NewTelegramRepo(t)
				storageService := mocks.NewStorageService(t)
				telegramRepo.EXPECT().GetChat(&client.GetChatRequest{
					ChatId: int64(-10100),
				}).Return(&client.Chat{
					Type: &client.ChatTypeSupergroup{},
				}, nil)
				telegramRepo.EXPECT().GetMessageLinkInfo(&client.GetMessageLinkInfoRequest{
					Url: "https://t.me/test/123",
				}).Return(&client.MessageLinkInfo{
					Message: &client.Message{
						ChatId: -10100,
						Id:     123,
					},
				}, nil)
				storageService.EXPECT().GetCopiedMessageIds(int64(-10100), int64(123)).Return([]string{
					"rule1:-10114:789",
				})
				storageService.EXPECT().GetNewMessageId(int64(-10114), int64(789)).Return(int64(456))
				telegramRepo.EXPECT().GetMessageLink(&client.GetMessageLinkRequest{
					ChatId:    -10114,
					MessageId: 456,
				}).Return(nil, errors.New("message link error"))
				return New(telegramRepo, storageService, nil)
			},
		},
		{
			name: "successful_link_replacement",
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
			srcChatId: -10100,
			dstChatId: -10114,
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
				telegramRepo.EXPECT().GetChat(&client.GetChatRequest{
					ChatId: int64(-10100),
				}).Return(&client.Chat{
					Type: &client.ChatTypeSupergroup{},
				}, nil)
				telegramRepo.EXPECT().GetMessageLinkInfo(&client.GetMessageLinkInfoRequest{
					Url: "https://t.me/test/123",
				}).Return(&client.MessageLinkInfo{
					Message: &client.Message{
						ChatId: -10100,
						Id:     123,
					},
				}, nil)
				storageService.EXPECT().GetCopiedMessageIds(int64(-10100), int64(123)).Return([]string{
					"rule1:-10114:789",
				})
				storageService.EXPECT().GetNewMessageId(int64(-10114), int64(789)).Return(int64(456))
				telegramRepo.EXPECT().GetMessageLink(&client.GetMessageLinkRequest{
					ChatId:    -10114,
					MessageId: 456,
				}).Return(&client.MessageLink{
					Link:     "https://t.me/newchat/456",
					IsPublic: true,
				}, nil)
				return New(telegramRepo, storageService, nil)
			},
		},

		{
			name: "url_entity-successful_replacement",
			formattedText: &client.FormattedText{
				Text: "Check https://t.me/test/123 for details",
				Entities: []*client.TextEntity{
					{
						Offset: 6,
						Length: 21,
						Type:   &client.TextEntityTypeUrl{},
					},
				},
			},
			srcChatId: -10100,
			dstChatId: -10114,
			expectedEntities: []*client.TextEntity{
				{
					Offset: 6,
					Length: 24, // длина "https://t.me/newchat/456"
					Type: &client.TextEntityTypeTextUrl{
						Url: "https://t.me/newchat/456",
					},
				},
			},
			setup: func(t *testing.T) *Service {
				telegramRepo := mocks.NewTelegramRepo(t)
				storageService := mocks.NewStorageService(t)
				telegramRepo.EXPECT().GetChat(&client.GetChatRequest{
					ChatId: int64(-10100),
				}).Return(&client.Chat{
					Type: &client.ChatTypeSupergroup{},
				}, nil)
				telegramRepo.EXPECT().GetMessageLinkInfo(&client.GetMessageLinkInfoRequest{
					Url: "https://t.me/test/123",
				}).Return(&client.MessageLinkInfo{
					Message: &client.Message{
						ChatId: -10100,
						Id:     123,
					},
				}, nil)
				storageService.EXPECT().GetCopiedMessageIds(int64(-10100), int64(123)).Return([]string{
					"rule1:-10114:789",
				})
				storageService.EXPECT().GetNewMessageId(int64(-10114), int64(789)).Return(int64(456))
				telegramRepo.EXPECT().GetMessageLink(&client.GetMessageLinkRequest{
					ChatId:    -10114,
					MessageId: 456,
				}).Return(&client.MessageLink{
					Link:     "https://t.me/newchat/456",
					IsPublic: true,
				}, nil)
				return New(telegramRepo, storageService, nil)
			},
		},
		{
			name: "url_entity_delete_external",
			formattedText: &client.FormattedText{
				Text: "Check https://t.me/test/123 for details",
				Entities: []*client.TextEntity{
					{
						Offset: 6,
						Length: 21,
						Type:   &client.TextEntityTypeUrl{},
					},
				},
			},
			srcChatId: -10100,
			dstChatId: -10114, // destination с deleteExternal=true
			expectedEntities: []*client.TextEntity{
				{
					Offset: 6,
					Length: 12, // длина domain.DELETED_LINK
					Type:   &client.TextEntityTypeStrikethrough{},
				},
			},
			setup: func(t *testing.T) *Service {
				telegramRepo := mocks.NewTelegramRepo(t)
				storageService := mocks.NewStorageService(t)
				telegramRepo.EXPECT().GetChat(&client.GetChatRequest{
					ChatId: int64(-10100),
				}).Return(&client.Chat{
					Type: &client.ChatTypeSupergroup{},
				}, nil)
				telegramRepo.EXPECT().GetMessageLinkInfo(&client.GetMessageLinkInfoRequest{
					Url: "https://t.me/test/123",
				}).Return(&client.MessageLinkInfo{
					Message: &client.Message{
						ChatId: -10100,
						Id:     123,
					},
				}, nil)
				storageService.EXPECT().GetCopiedMessageIds(int64(-10100), int64(123)).Return([]string{})
				telegramRepo.EXPECT().ParseTextEntities(&client.ParseTextEntitiesRequest{
					Text: domain.DELETED_LINK,
					ParseMode: &client.TextParseModeMarkdown{
						Version: 2,
					},
				}).Return(&client.FormattedText{
					Text: domain.DELETED_LINK,
					Entities: []*client.TextEntity{
						{
							Offset: 0,
							Length: 12,
							Type:   &client.TextEntityTypeStrikethrough{},
						},
					},
				}, nil)
				return New(telegramRepo, storageService, nil)
			},
		},
		{
			name: "message_link_is_not_public",
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
			srcChatId: -10100,
			dstChatId: -10114,
			expectedEntities: []*client.TextEntity{
				{
					Type: &client.TextEntityTypeStrikethrough{},
				},
			},
			setup: func(t *testing.T) *Service {
				telegramRepo := mocks.NewTelegramRepo(t)
				storageService := mocks.NewStorageService(t)
				telegramRepo.EXPECT().GetChat(&client.GetChatRequest{
					ChatId: int64(-10100),
				}).Return(&client.Chat{
					Type: &client.ChatTypeSupergroup{},
				}, nil)
				telegramRepo.EXPECT().GetMessageLinkInfo(&client.GetMessageLinkInfoRequest{
					Url: "https://t.me/test/123",
				}).Return(&client.MessageLinkInfo{
					Message: &client.Message{
						ChatId: -10100,
						Id:     123,
					},
				}, nil)
				storageService.EXPECT().GetCopiedMessageIds(int64(-10100), int64(123)).Return([]string{
					"rule1:-10114:789",
				})
				storageService.EXPECT().GetNewMessageId(int64(-10114), int64(789)).Return(int64(456))
				telegramRepo.EXPECT().GetMessageLink(&client.GetMessageLinkRequest{
					ChatId:    -10114,
					MessageId: 456,
				}).Return(&client.MessageLink{
					Link:     "https://t.me/newchat/456",
					IsPublic: false, // НЕ публичная ссылка
				}, nil)
				return New(telegramRepo, storageService, nil)
			},
		},
		{
			name: "external_chat_link_not_our_chat",
			formattedText: &client.FormattedText{
				Text: "test",
				Entities: []*client.TextEntity{
					{
						Type: &client.TextEntityTypeTextUrl{
							Url: "https://t.me/external/123",
						},
					},
				},
			},
			srcChatId: -10100,
			dstChatId: -10114,
			expectedEntities: []*client.TextEntity{
				{
					Type: &client.TextEntityTypeTextUrl{
						Url: "https://t.me/external/123",
					},
				},
			},
			setup: func(t *testing.T) *Service {
				telegramRepo := mocks.NewTelegramRepo(t)
				telegramRepo.EXPECT().GetChat(&client.GetChatRequest{
					ChatId: int64(-10100),
				}).Return(&client.Chat{
					Type: &client.ChatTypeSupergroup{},
				}, nil)
				telegramRepo.EXPECT().GetMessageLinkInfo(&client.GetMessageLinkInfoRequest{
					Url: "https://t.me/external/123",
				}).Return(&client.MessageLinkInfo{
					Message: &client.Message{
						ChatId: -10200, // другой чат (не srcChatId -10100)
						Id:     123,
					},
				}, nil)
				return New(telegramRepo, nil, nil)
			},
		},
		{
			name: "message_not_found_by_link",
			formattedText: &client.FormattedText{
				Text: "test",
				Entities: []*client.TextEntity{
					{
						Type: &client.TextEntityTypeTextUrl{
							Url: "https://t.me/notfound/123",
						},
					},
				},
			},
			srcChatId: -10100,
			dstChatId: -10114,
			expectedEntities: []*client.TextEntity{
				{
					Type: &client.TextEntityTypeTextUrl{
						Url: "https://t.me/notfound/123",
					},
				},
			},
			setup: func(t *testing.T) *Service {
				telegramRepo := mocks.NewTelegramRepo(t)
				telegramRepo.EXPECT().GetChat(&client.GetChatRequest{
					ChatId: int64(-10100),
				}).Return(&client.Chat{
					Type: &client.ChatTypeSupergroup{},
				}, nil)
				telegramRepo.EXPECT().GetMessageLinkInfo(&client.GetMessageLinkInfoRequest{
					Url: "https://t.me/notfound/123",
				}).Return(&client.MessageLinkInfo{
					Message: nil, // сообщение не найдено
				}, nil)
				return New(telegramRepo, nil, nil)
			},
		},
		{
			name: "url_entity_message_link_is_not_public",
			formattedText: &client.FormattedText{
				Text: "Check https://t.me/test/123 for details",
				Entities: []*client.TextEntity{
					{
						Offset: 6,
						Length: 21,
						Type:   &client.TextEntityTypeUrl{},
					},
				},
			},
			srcChatId: -10100,
			dstChatId: -10114,
			expectedEntities: []*client.TextEntity{
				{
					Offset: 6,
					Length: 12, // длина domain.DELETED_LINK
					Type:   &client.TextEntityTypeStrikethrough{},
				},
			},
			setup: func(t *testing.T) *Service {
				telegramRepo := mocks.NewTelegramRepo(t)
				storageService := mocks.NewStorageService(t)
				telegramRepo.EXPECT().GetChat(&client.GetChatRequest{
					ChatId: int64(-10100),
				}).Return(&client.Chat{
					Type: &client.ChatTypeSupergroup{},
				}, nil)
				telegramRepo.EXPECT().GetMessageLinkInfo(&client.GetMessageLinkInfoRequest{
					Url: "https://t.me/test/123",
				}).Return(&client.MessageLinkInfo{
					Message: &client.Message{
						ChatId: -10100,
						Id:     123,
					},
				}, nil)
				storageService.EXPECT().GetCopiedMessageIds(int64(-10100), int64(123)).Return([]string{
					"rule1:-10114:789",
				})
				storageService.EXPECT().GetNewMessageId(int64(-10114), int64(789)).Return(int64(456))
				telegramRepo.EXPECT().GetMessageLink(&client.GetMessageLinkRequest{
					ChatId:    -10114,
					MessageId: 456,
				}).Return(&client.MessageLink{
					Link:     "https://t.me/newchat/456",
					IsPublic: false, // НЕ публичная ссылка
				}, nil)
				telegramRepo.EXPECT().ParseTextEntities(&client.ParseTextEntitiesRequest{
					Text: domain.DELETED_LINK,
					ParseMode: &client.TextParseModeMarkdown{
						Version: 2,
					},
				}).Return(&client.FormattedText{
					Text: domain.DELETED_LINK,
					Entities: []*client.TextEntity{
						{
							Offset: 0,
							Length: 12,
							Type:   &client.TextEntityTypeStrikethrough{},
						},
					},
				}, nil)
				return New(telegramRepo, storageService, nil)
			},
		},
		{
			name: "url_entity_external_chat_link",
			formattedText: &client.FormattedText{
				Text: "Check https://t.me/external/123 for details",
				Entities: []*client.TextEntity{
					{
						Offset: 6,
						Length: 25,
						Type:   &client.TextEntityTypeUrl{},
					},
				},
			},
			srcChatId: -10100,
			dstChatId: -10114,
			expectedEntities: []*client.TextEntity{
				{
					Offset: 6,
					Length: 25, // длина не изменилась
					Type:   &client.TextEntityTypeUrl{},
				},
			},
			setup: func(t *testing.T) *Service {
				telegramRepo := mocks.NewTelegramRepo(t)
				telegramRepo.EXPECT().GetChat(&client.GetChatRequest{
					ChatId: int64(-10100),
				}).Return(&client.Chat{
					Type: &client.ChatTypeSupergroup{},
				}, nil)
				telegramRepo.EXPECT().GetMessageLinkInfo(&client.GetMessageLinkInfoRequest{
					Url: "https://t.me/external/123",
				}).Return(&client.MessageLinkInfo{
					Message: &client.Message{
						ChatId: -10200, // другой чат (не srcChatId -10100)
						Id:     123,
					},
				}, nil)
				return New(telegramRepo, nil, nil)
			},
		},
		{
			name: "url_entity_message_not_found_by_link",
			formattedText: &client.FormattedText{
				Text: "Check https://t.me/notfound/123 for details",
				Entities: []*client.TextEntity{
					{
						Offset: 6,
						Length: 25,
						Type:   &client.TextEntityTypeUrl{},
					},
				},
			},
			srcChatId: -10100,
			dstChatId: -10114,
			expectedEntities: []*client.TextEntity{
				{
					Offset: 6,
					Length: 25, // длина не изменилась
					Type:   &client.TextEntityTypeUrl{},
				},
			},
			setup: func(t *testing.T) *Service {
				telegramRepo := mocks.NewTelegramRepo(t)
				telegramRepo.EXPECT().GetChat(&client.GetChatRequest{
					ChatId: int64(-10100),
				}).Return(&client.Chat{
					Type: &client.ChatTypeSupergroup{},
				}, nil)
				telegramRepo.EXPECT().GetMessageLinkInfo(&client.GetMessageLinkInfoRequest{
					Url: "https://t.me/notfound/123",
				}).Return(&client.MessageLinkInfo{
					Message: nil, // сообщение не найдено
				}, nil)
				return New(telegramRepo, nil, nil)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var transformService *Service
			spylogHandler := spylog.GetHandler(t.Name(), func() {
				transformService = test.setup(t)
			})

			transformService.replaceMyselfLinks(test.formattedText, test.srcChatId, test.dstChatId, config.Engine)

			if test.expectedError != nil {
				records := spylogHandler.GetRecords()
				require.Equal(t, 1, len(records))
				record := records[0]
				assert.Equal(t, test.expectedError.Error(), record.Message)
			}

			assert.Equal(t, test.expectedEntities, test.formattedText.Entities)
		})
	}
}

func Test_replaceFragments(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		formattedText *client.FormattedText
		dstChatId     int64
		expectedText  string
		expectedError error
	}{
		{
			name: "destination_not_found",
			formattedText: &client.FormattedText{
				Text: "some text",
			},
			dstChatId:     -10199, // не существует в config.yml
			expectedText:  "some text",
			expectedError: log.NewError("destination not found"),
		},
		{
			name: "no_replace_fragments",
			formattedText: &client.FormattedText{
				Text: "some text",
			},
			dstChatId:    -10113,
			expectedText: "some text",
		},
		{
			name: "single_replacement",
			formattedText: &client.FormattedText{
				Text: "hello world",
			},
			dstChatId:    -10110,
			expectedText: "12345 67890",
		},
		{
			name: "case_insensitive_replacement",
			formattedText: &client.FormattedText{
				Text: "This is a test and TEST and Test",
			},
			dstChatId:    -10111,
			expectedText: "This is a Тест and Тест and Тест",
		},
		{
			name: "multiple_occurrences",
			formattedText: &client.FormattedText{
				Text: "old text with old values and old data",
			},
			dstChatId:    -10112,
			expectedText: "new text with new values and new data",
		},
		{
			name: "no_matches",
			formattedText: &client.FormattedText{
				Text: "some random text",
			},
			dstChatId:    -10110,
			expectedText: "some random text",
		},
		{
			name: "partial_word_match",
			formattedText: &client.FormattedText{
				Text: "hello? and world!",
			},
			dstChatId:    -10110,
			expectedText: "12345? and 67890!",
		},
		{
			name: "empty_text",
			formattedText: &client.FormattedText{
				Text: "",
			},
			dstChatId:    -10110,
			expectedText: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var transformService *Service
			spylogHandler := spylog.GetHandler(t.Name(), func() {
				transformService = New(nil, nil, nil)
			})

			transformService.replaceFragments(test.formattedText, test.dstChatId, config.Engine)

			if test.expectedError != nil {
				records := spylogHandler.GetRecords()
				require.Equal(t, len(records), 1)
				record := records[0]
				assert.Equal(t, test.expectedError.Error(), record.Message)
			}

			assert.Equal(t, test.expectedText, test.formattedText.Text)
		})
	}
}

func Test_addAutoAnswer(t *testing.T) {
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
			name:          "source_not_found",
			formattedText: &client.FormattedText{},
			src: &client.Message{
				ChatId: -10199, // не существует в config.yml
				Id:     123,
			},
			expectedText:     "",
			expectedEntities: nil,
			setup: func(t *testing.T, src *client.Message) *Service {
				return New(nil, nil, nil)
			},
			expectedError: log.NewError("source not found"),
		},
		{
			name:          "auto_answer_disabled",
			formattedText: &client.FormattedText{},
			src: &client.Message{
				ChatId: -10107,
				Id:     123,
			},
			expectedText:     "",
			expectedEntities: nil,
			setup: func(t *testing.T, src *client.Message) *Service {
				return New(nil, nil, nil)
			},
			expectedError: log.NewError("source.AutoAnswer is false"),
		},
		{
			name:          "reply_markup_data_is_empty",
			formattedText: &client.FormattedText{},
			src: &client.Message{
				ChatId: -10106,
				Id:     123,
			},
			expectedText:     "",
			expectedEntities: nil,
			setup: func(t *testing.T, src *client.Message) *Service {
				messageService := mocks.NewMessageService(t)
				messageService.EXPECT().GetReplyMarkupData(src).Return([]byte{})
				return New(nil, nil, messageService)
			},
			expectedError: log.NewError("replyMarkupData is empty"),
		},
		{
			name:          "get_callback_query_answer_error",
			formattedText: &client.FormattedText{},
			src: &client.Message{
				ChatId: -10106,
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
			name: "successful_auto_answer",
			formattedText: &client.FormattedText{
				Text:     "existing text",
				Entities: []*client.TextEntity{},
			},
			src: &client.Message{
				ChatId: -10106,
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
			name:          "empty_formatted_text",
			formattedText: &client.FormattedText{},
			src: &client.Message{
				ChatId: -10106,
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
			spylogHandler := spylog.GetHandler(t.Name(), func() {
				transformService = test.setup(t, test.src)
			})

			transformService.addAutoAnswer(test.formattedText, test.src, config.Engine)

			if test.expectedError != nil {
				records := spylogHandler.GetRecords()
				require.Equal(t, len(records), 1)
				record := records[0]
				assert.Equal(t, test.expectedError.Error(), record.Message)
			}

			assert.Equal(t, test.expectedText, test.formattedText.Text)
			assert.Equal(t, test.expectedEntities, test.formattedText.Entities)
		})
	}
}

func Test_addSourceSign(t *testing.T) {
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
			name:          "source_not_found",
			formattedText: &client.FormattedText{},
			src: &client.Message{
				ChatId: -10199, // не существует в config.yml
				Id:     123,
			},
			dstChatId:        0, // не используется
			expectedText:     "",
			expectedEntities: nil,
			setup: func(t *testing.T, src *client.Message) *Service {
				return New(nil, nil, nil)
			},
			expectedError: log.NewError("source not found"),
		},
		{
			name:          "sign_not_for_this_chat",
			formattedText: &client.FormattedText{},
			src: &client.Message{
				ChatId: -10103, // sign.for = [-10108], а не -10109
				Id:     123,
			},
			dstChatId:        -10109,
			expectedText:     "",
			expectedEntities: nil,
			setup: func(t *testing.T, src *client.Message) *Service {
				return New(nil, nil, nil)
			},
			expectedError: log.NewError("source.Sign without dstChatId"),
		},
		{
			name:          "no_sign_configured",
			formattedText: &client.FormattedText{},
			src: &client.Message{
				ChatId: -10104, // empty source
				Id:     123,
			},
			dstChatId:        -10109,
			expectedText:     "",
			expectedEntities: nil,
			setup: func(t *testing.T, src *client.Message) *Service {
				return New(nil, nil, nil)
			},
			expectedError: log.NewError("source.Sign without dstChatId"),
		},
		{
			name:          "successful_sign_addition",
			formattedText: &client.FormattedText{},
			src: &client.Message{
				ChatId: -10100, // sign для dstChatId -10109
				Id:     123,
			},
			dstChatId:        -10109,
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
			name: "sign_addition_to_existing_text",
			formattedText: &client.FormattedText{
				Text:     "existing text",
				Entities: []*client.TextEntity{},
			},
			src: &client.Message{
				ChatId: -10100,
				Id:     123,
			},
			dstChatId:        -10109,
			expectedText:     "existing text\n\nTest Source",
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
				return New(telegramRepo, nil, nil)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var transformService *Service
			spylogHandler := spylog.GetHandler(t.Name(), func() {
				transformService = test.setup(t, test.src)
			})

			transformService.addSourceSign(test.formattedText, test.src, test.dstChatId, config.Engine)

			if test.expectedError != nil {
				records := spylogHandler.GetRecords()
				require.Equal(t, 1, len(records))
				record := records[0]
				assert.Equal(t, test.expectedError.Error(), record.Message)
			}

			assert.Equal(t, test.expectedText, test.formattedText.Text)
			assert.Equal(t, test.expectedEntities, test.formattedText.Entities)
		})
	}
}

func Test_addSourceLink(t *testing.T) {
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
			name:          "source_not_found",
			formattedText: &client.FormattedText{},
			src: &client.Message{
				ChatId: -10199, // не существует в config.yml
				Id:     123,
			},
			dstChatId:        0, // не используется
			expectedText:     "",
			expectedEntities: nil,
			setup: func(t *testing.T, src *client.Message) *Service {
				return New(nil, nil, nil)
			},
			expectedError: log.NewError("source not found"),
		},
		{
			name:          "link_not_for_this_chat",
			formattedText: &client.FormattedText{},
			src: &client.Message{
				ChatId:       -10100, // у этого source нет link для dstChatId -10109
				Id:           123,
				MediaAlbumId: 0,
			},
			dstChatId:        -10109,
			expectedText:     "",
			expectedEntities: nil,
			setup: func(t *testing.T, src *client.Message) *Service {
				return New(nil, nil, nil)
			},
			expectedError: log.NewError("source.Link without dstChatId"),
		},
		{
			name:          "no_link_configured",
			formattedText: &client.FormattedText{},
			src: &client.Message{
				ChatId:       -10104, // empty source
				Id:           123,
				MediaAlbumId: 0,
			},
			dstChatId:        -10109,
			expectedText:     "",
			expectedEntities: nil,
			setup: func(t *testing.T, src *client.Message) *Service {
				return New(nil, nil, nil)
			},
			expectedError: log.NewError("source.Link without dstChatId"),
		},
		{
			name:          "successful_link_addition",
			formattedText: &client.FormattedText{},
			src: &client.Message{
				ChatId:       -10101, // link для dstChatId -10109
				Id:           123,
				MediaAlbumId: 0,
			},
			dstChatId:        -10109,
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
			name: "link_addition_to_existing_text",
			formattedText: &client.FormattedText{
				Text:     "existing text",
				Entities: []*client.TextEntity{},
			},
			src: &client.Message{
				ChatId:       -10101,
				Id:           123,
				MediaAlbumId: 0,
			},
			dstChatId:        -10109,
			expectedText:     "existing text\n\n[🔗Source Link](https://t.me/test/123)",
			expectedEntities: []*client.TextEntity{},
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
			name:          "get_message_link_error",
			formattedText: &client.FormattedText{},
			src: &client.Message{
				ChatId:       -10101,
				Id:           123,
				MediaAlbumId: 0,
			},
			dstChatId:        -10109,
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
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var transformService *Service
			spylogHandler := spylog.GetHandler(t.Name(), func() {
				transformService = test.setup(t, test.src)
			})

			transformService.addSourceLink(test.formattedText, test.src, test.dstChatId, config.Engine)

			if test.expectedError != nil {
				records := spylogHandler.GetRecords()
				require.Equal(t, 1, len(records))
				record := records[0]
				assert.Equal(t, test.expectedError.Error(), record.Message)
			}

			assert.Equal(t, test.expectedText, test.formattedText.Text)
			assert.Equal(t, test.expectedEntities, test.formattedText.Entities)
		})
	}
}

func Test_addText(t *testing.T) {
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
			expectedError:    assert.AnError,
			expectedText:     "",
			expectedEntities: nil,
		},
		{
			name: "with_existing_text",
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
			name: "add_text_with_entities",
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
					Offset: 11,
					Length: 4,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			telegramRepo := mocks.NewTelegramRepo(t)
			var returnEntities []*client.TextEntity
			if test.name == "add_text_with_entities" {
				returnEntities = []*client.TextEntity{
					{
						Type:   &client.TextEntityTypeBold{},
						Offset: 1,
						Length: 4,
					},
				}
			}

			telegramRepo.EXPECT().ParseTextEntities(&client.ParseTextEntitiesRequest{
				Text: test.text,
				ParseMode: &client.TextParseModeMarkdown{
					Version: 2,
				},
			}).Return(&client.FormattedText{
				Text:     test.text,
				Entities: returnEntities,
			}, test.expectedError)

			var transformService *Service
			spylogHandler := spylog.GetHandler(t.Name(), func() {
				transformService = New(telegramRepo, nil, nil)
			})

			transformService.addText(test.formattedText, test.text)

			if test.expectedError != nil {
				records := spylogHandler.GetRecords()
				require.Equal(t, len(records), 1)
				assert.Equal(t, test.expectedError.Error(), records[0].Message)
			}

			assert.Equal(t, test.expectedText, test.formattedText.Text)
			assert.Equal(t, test.expectedEntities, test.formattedText.Entities)
		})
	}
}

func Test_applyReplacements(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		setup            func() (*client.FormattedText, []*replacement)
		expectedText     string
		expectedEntities []*client.TextEntity
	}{
		{
			name: "only_change_entity_type_no_text_change",
			setup: func() (*client.FormattedText, []*replacement) {
				entity := &client.TextEntity{
					Offset: 5,
					Length: 7,
					Type:   &client.TextEntityTypeTextUrl{Url: "https://example.com"},
				}
				formattedText := &client.FormattedText{
					Text:     "test message",
					Entities: []*client.TextEntity{entity},
				}
				replacements := []*replacement{
					{
						OldEntity:     entity, // используем тот же объект
						NewText:       "",     // пустой - только меняем тип
						NewEntityType: &client.TextEntityTypeStrikethrough{},
					},
				}
				return formattedText, replacements
			},
			expectedText: "test message",
			expectedEntities: []*client.TextEntity{
				{
					Offset: 5,
					Length: 7,
					Type:   &client.TextEntityTypeStrikethrough{},
				},
			},
		},
		{
			name: "entity_fully_inside_replaced_text_gets_deleted",
			setup: func() (*client.FormattedText, []*replacement) {
				urlEntity := &client.TextEntity{
					Offset: 6,
					Length: 21, // URL entity
					Type:   &client.TextEntityTypeUrl{},
				}
				boldEntity := &client.TextEntity{
					Offset: 12, // Bold entity внутри URL (t.me/test/123)
					Length: 9,
					Type:   &client.TextEntityTypeBold{},
				}
				formattedText := &client.FormattedText{
					Text:     "Check https://t.me/test/123 here",
					Entities: []*client.TextEntity{urlEntity, boldEntity},
				}
				replacements := []*replacement{
					{
						OldEntity:     urlEntity,
						NewText:       domain.DELETED_LINK,
						NewEntityType: &client.TextEntityTypeStrikethrough{},
					},
				}
				return formattedText, replacements
			},
			expectedText: "Check DELETED_LINK here",
			expectedEntities: []*client.TextEntity{
				{
					Offset: 6,
					Length: 12, // длина domain.DELETED_LINK
					Type:   &client.TextEntityTypeStrikethrough{},
				},
				// Bold entity должен быть удален, так как был внутри заменяемого текста
			},
		},
		{
			name: "entity_after_replacement_gets_shifted",
			setup: func() (*client.FormattedText, []*replacement) {
				urlEntity := &client.TextEntity{
					Offset: 6,
					Length: 21, // URL entity
					Type:   &client.TextEntityTypeUrl{},
				}
				boldEntity := &client.TextEntity{
					Offset: 32, // Bold entity после URL
					Length: 9,
					Type:   &client.TextEntityTypeBold{},
				}
				formattedText := &client.FormattedText{
					Text:     "Start https://t.me/test/123 and bold text",
					Entities: []*client.TextEntity{urlEntity, boldEntity},
				}
				replacements := []*replacement{
					{
						OldEntity:     urlEntity,
						NewText:       domain.DELETED_LINK, // 12 символов вместо 21
						NewEntityType: &client.TextEntityTypeStrikethrough{},
					},
				}
				return formattedText, replacements
			},
			expectedText: "Start DELETED_LINK and bold text",
			expectedEntities: []*client.TextEntity{
				{
					Offset: 6,
					Length: 12,
					Type:   &client.TextEntityTypeStrikethrough{},
				},
				{
					Offset: 23, // сдвинулся на -9 (21-12)
					Length: 9,
					Type:   &client.TextEntityTypeBold{},
				},
			},
		},
		{
			name: "entity_before_replacement_stays_unchanged",
			setup: func() (*client.FormattedText, []*replacement) {
				boldEntity := &client.TextEntity{
					Offset: 0, // Bold entity до URL
					Length: 4,
					Type:   &client.TextEntityTypeBold{},
				}
				urlEntity := &client.TextEntity{
					Offset: 10,
					Length: 21, // URL entity
					Type:   &client.TextEntityTypeUrl{},
				}
				formattedText := &client.FormattedText{
					Text:     "Bold text https://t.me/test/123 end",
					Entities: []*client.TextEntity{boldEntity, urlEntity},
				}
				replacements := []*replacement{
					{
						OldEntity:     urlEntity,
						NewText:       domain.DELETED_LINK,
						NewEntityType: &client.TextEntityTypeStrikethrough{},
					},
				}
				return formattedText, replacements
			},
			expectedText: "Bold text DELETED_LINK end",
			expectedEntities: []*client.TextEntity{
				{
					Offset: 0, // остается без изменений
					Length: 4,
					Type:   &client.TextEntityTypeBold{},
				},
				{
					Offset: 10,
					Length: 12,
					Type:   &client.TextEntityTypeStrikethrough{},
				},
			},
		},
		{
			name: "multiple_replacements_processed_in_reverse_order",
			setup: func() (*client.FormattedText, []*replacement) {
				url1Entity := &client.TextEntity{
					Offset: 6,
					Length: 21, // первый URL
					Type:   &client.TextEntityTypeUrl{},
				}
				url2Entity := &client.TextEntity{
					Offset: 39,
					Length: 21, // второй URL
					Type:   &client.TextEntityTypeUrl{},
				}
				formattedText := &client.FormattedText{
					Text:     "First https://t.me/test/123 and second https://t.me/test/456 end",
					Entities: []*client.TextEntity{url1Entity, url2Entity},
				}
				replacements := []*replacement{
					{
						OldEntity:     url1Entity,
						NewText:       "LINK1",
						NewEntityType: &client.TextEntityTypeStrikethrough{},
					},
					{
						OldEntity:     url2Entity,
						NewText:       "LINK2",
						NewEntityType: &client.TextEntityTypeStrikethrough{},
					},
				}
				return formattedText, replacements
			},
			expectedText: "First LINK1 and second LINK2 end",
			expectedEntities: []*client.TextEntity{
				{
					Offset: 6,
					Length: 5, // LINK1
					Type:   &client.TextEntityTypeStrikethrough{},
				},
				{
					Offset: 23, // сдвинулся после замены первого URL
					Length: 5,  // LINK2
					Type:   &client.TextEntityTypeStrikethrough{},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			formattedText, replacements := test.setup()
			service := New(nil, nil, nil)
			service.applyReplacements(formattedText, replacements)

			assert.Equal(t, test.expectedText, formattedText.Text)
			assert.Equal(t, test.expectedEntities, formattedText.Entities)
		})
	}
}

func Test_collectMarkdownReplacements(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		inputText string
		parsed    *client.FormattedText
		expects   []*replacement
		err       error
	}{
		{
			name:      "plain_text_no_markdown",
			inputText: "hello world",
			parsed: &client.FormattedText{
				Text:     "hello world",
				Entities: nil,
			},
			expects: nil, // нет markdown, нет replacements
		},
		{
			name:      "markdown_bold",
			inputText: "*bold*",
			parsed: &client.FormattedText{
				Text:     "bold",
				Entities: []*client.TextEntity{{Offset: 0, Length: 4, Type: &client.TextEntityTypeBold{}}},
			},
			expects: []*replacement{
				{
					OldEntity:     &client.TextEntity{Offset: 0, Length: 6, Type: nil},
					NewText:       "bold",
					NewEntityType: nil,
				},
				{
					OldEntity:     &client.TextEntity{Offset: 0, Length: 4, Type: &client.TextEntityTypeBold{}},
					NewText:       "",
					NewEntityType: &client.TextEntityTypeBold{},
				},
			},
		},
		{
			name:      "parse_error",
			inputText: "*fail*",
			parsed:    nil,
			expects:   nil,
			err:       assert.AnError,
		},
		{
			name:      "entities_only_with_no_text_replacement",
			inputText: "bold",
			parsed: &client.FormattedText{
				Text:     "bold",
				Entities: []*client.TextEntity{{Offset: 0, Length: 4, Type: &client.TextEntityTypeBold{}}},
			},
			expects: []*replacement{
				{
					OldEntity:     &client.TextEntity{Offset: 0, Length: 4, Type: &client.TextEntityTypeBold{}},
					NewText:       "",
					NewEntityType: &client.TextEntityTypeBold{},
				},
			},
		},
		{
			name:      "multiple_entities",
			inputText: "*a*_b_",
			parsed: &client.FormattedText{
				Text: "ab",
				Entities: []*client.TextEntity{
					{Offset: 0, Length: 1, Type: &client.TextEntityTypeBold{}},
					{Offset: 1, Length: 1, Type: &client.TextEntityTypeItalic{}},
				},
			},
			expects: []*replacement{
				{
					OldEntity:     &client.TextEntity{Offset: 0, Length: 6, Type: nil},
					NewText:       "ab",
					NewEntityType: nil,
				},
				{
					OldEntity:     &client.TextEntity{Offset: 0, Length: 1, Type: &client.TextEntityTypeBold{}},
					NewText:       "",
					NewEntityType: &client.TextEntityTypeBold{},
				},
				{
					OldEntity:     &client.TextEntity{Offset: 1, Length: 1, Type: &client.TextEntityTypeItalic{}},
					NewText:       "",
					NewEntityType: &client.TextEntityTypeItalic{},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			telegramRepo := mocks.NewTelegramRepo(t)
			if test.err != nil {
				telegramRepo.EXPECT().ParseTextEntities(&client.ParseTextEntitiesRequest{
					Text:      test.inputText,
					ParseMode: &client.TextParseModeMarkdown{Version: 2},
				}).Return(nil, test.err)
			} else {
				telegramRepo.EXPECT().ParseTextEntities(&client.ParseTextEntitiesRequest{
					Text:      test.inputText,
					ParseMode: &client.TextParseModeMarkdown{Version: 2},
				}).Return(test.parsed, nil)
			}
			service := New(telegramRepo, nil, nil)
			result := service.collectMarkdownReplacements(0, test.inputText)
			if test.expects == nil {
				assert.Nil(t, result)
			} else {
				assert.Equal(t, len(test.expects), len(result))
				for i := range test.expects {
					assert.Equal(t, test.expects[i].OldEntity.Offset, result[i].OldEntity.Offset)
					assert.Equal(t, test.expects[i].OldEntity.Length, result[i].OldEntity.Length)
					assert.Equal(t, test.expects[i].NewText, result[i].NewText)
					assert.Equal(t, test.expects[i].NewEntityType, result[i].NewEntityType)
				}
			}
		})
	}
}

func Test_applyMarkdownReplacements(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		initial  *client.FormattedText
		repls    []*replacement
		expected *client.FormattedText
	}{
		{
			name: "replace_text_and_add_bold_entity",
			initial: &client.FormattedText{
				Text:     "*bold*",
				Entities: []*client.TextEntity{},
			},
			repls: []*replacement{
				{
					OldEntity:     &client.TextEntity{Offset: 0, Length: 6, Type: nil},
					NewText:       "bold",
					NewEntityType: nil,
				},
				{
					OldEntity:     &client.TextEntity{Offset: 0, Length: 4, Type: &client.TextEntityTypeBold{}},
					NewText:       "",
					NewEntityType: &client.TextEntityTypeBold{},
				},
			},
			expected: &client.FormattedText{
				Text:     "bold",
				Entities: []*client.TextEntity{{Offset: 0, Length: 4, Type: &client.TextEntityTypeBold{}}},
			},
		},
		{
			name: "only_entity_replacement",
			initial: &client.FormattedText{
				Text:     "bold",
				Entities: []*client.TextEntity{},
			},
			repls: []*replacement{
				{
					OldEntity:     &client.TextEntity{Offset: 0, Length: 4, Type: &client.TextEntityTypeBold{}},
					NewText:       "",
					NewEntityType: &client.TextEntityTypeBold{},
				},
			},
			expected: &client.FormattedText{
				Text:     "bold",
				Entities: []*client.TextEntity{{Offset: 0, Length: 4, Type: &client.TextEntityTypeBold{}}},
			},
		},
		{
			name: "offset_shift_for_entities_after_text_replacement",
			initial: &client.FormattedText{
				Text:     "abcde",
				Entities: []*client.TextEntity{{Offset: 3, Length: 2, Type: &client.TextEntityTypeItalic{}}},
			},
			repls: []*replacement{
				{
					OldEntity:     &client.TextEntity{Offset: 0, Length: 2, Type: nil},
					NewText:       "xyzuvw",
					NewEntityType: nil,
				},
			},
			expected: &client.FormattedText{
				Text:     "xyzuvwcde",
				Entities: []*client.TextEntity{{Offset: 7, Length: 2, Type: &client.TextEntityTypeItalic{}}},
			},
		},
		{
			name: "offset_shift_for_entity_replacements_after_text_replacement",
			initial: &client.FormattedText{
				Text:     "abcde",
				Entities: []*client.TextEntity{},
			},
			repls: []*replacement{
				{
					OldEntity:     &client.TextEntity{Offset: 0, Length: 2, Type: nil},
					NewText:       "xyzuvw",
					NewEntityType: nil,
				},
				{
					OldEntity:     &client.TextEntity{Offset: 3, Length: 2, Type: &client.TextEntityTypeItalic{}},
					NewText:       "",
					NewEntityType: &client.TextEntityTypeItalic{},
				},
			},
			expected: &client.FormattedText{
				Text:     "xyzuvwcde",
				Entities: []*client.TextEntity{{Offset: 7, Length: 2, Type: &client.TextEntityTypeItalic{}}},
			},
		},
		{
			name: "empty_replacements",
			initial: &client.FormattedText{
				Text:     "abc",
				Entities: []*client.TextEntity{},
			},
			repls: nil,
			expected: &client.FormattedText{
				Text:     "abc",
				Entities: []*client.TextEntity{},
			},
		},
	}

	service := New(nil, nil, nil)
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			ft := &client.FormattedText{
				Text:     test.initial.Text,
				Entities: append([]*client.TextEntity{}, test.initial.Entities...),
			}
			service.applyMarkdownReplacements(ft, test.repls)
			assert.Equal(t, test.expected.Text, ft.Text)
			assert.Equal(t, test.expected.Entities, ft.Entities)
		})
	}
}

func Test_addPrevMessageId(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		formattedText    *client.FormattedText
		src              *client.Message
		dstChatId        int64
		prevMessageId    int64
		engineConfig     *domain.EngineConfig
		expectedText     string
		expectedEntities []*client.TextEntity
		setup            func(t *testing.T) *Service
	}{
		{
			name: "success_with_default_prev",
			formattedText: &client.FormattedText{
				Text:     "test message",
				Entities: []*client.TextEntity{},
			},
			src: &client.Message{
				ChatId:       -10121,
				Id:           123,
				MediaAlbumId: 0,
			},
			dstChatId:     -10123,
			prevMessageId: 456,
			engineConfig: &domain.EngineConfig{
				Sources: map[int64]*domain.Source{
					-10121: {
						Prev: "",
					},
				},
			},
			expectedText:     "test message\n\n[Prev](https://t.me/test/456)",
			expectedEntities: []*client.TextEntity{},
			setup: func(t *testing.T) *Service {
				telegramRepo := mocks.NewTelegramRepo(t)

				telegramRepo.EXPECT().GetMessageLink(&client.GetMessageLinkRequest{
					ChatId:    -10123,
					MessageId: 456,
					ForAlbum:  false,
				}).Return(&client.MessageLink{
					Link: "https://t.me/test/456",
				}, nil)

				telegramRepo.EXPECT().ParseTextEntities(&client.ParseTextEntitiesRequest{
					Text: "[Prev](https://t.me/test/456)",
					ParseMode: &client.TextParseModeMarkdown{
						Version: 2,
					},
				}).Return(&client.FormattedText{
					Text:     "[Prev](https://t.me/test/456)",
					Entities: []*client.TextEntity{},
				}, nil)

				return New(telegramRepo, nil, nil)
			},
		},
		{
			name: "success_with_custom_prev",
			formattedText: &client.FormattedText{
				Text:     "test message",
				Entities: []*client.TextEntity{},
			},
			src: &client.Message{
				ChatId:       -10121,
				Id:           123,
				MediaAlbumId: 0,
			},
			dstChatId:     -10123,
			prevMessageId: 456,
			engineConfig: &domain.EngineConfig{
				Sources: map[int64]*domain.Source{
					-10121: {
						Prev: "Предыдущее",
					},
				},
			},
			expectedText:     "test message\n\n[Предыдущее](https://t.me/test/456)",
			expectedEntities: []*client.TextEntity{},
			setup: func(t *testing.T) *Service {
				telegramRepo := mocks.NewTelegramRepo(t)

				telegramRepo.EXPECT().GetMessageLink(&client.GetMessageLinkRequest{
					ChatId:    -10123,
					MessageId: 456,
					ForAlbum:  false,
				}).Return(&client.MessageLink{
					Link: "https://t.me/test/456",
				}, nil)

				telegramRepo.EXPECT().ParseTextEntities(&client.ParseTextEntitiesRequest{
					Text: "[Предыдущее](https://t.me/test/456)",
					ParseMode: &client.TextParseModeMarkdown{
						Version: 2,
					},
				}).Return(&client.FormattedText{
					Text:     "[Предыдущее](https://t.me/test/456)",
					Entities: []*client.TextEntity{},
				}, nil)

				return New(telegramRepo, nil, nil)
			},
		},
		{
			name: "success_with_media_album",
			formattedText: &client.FormattedText{
				Text:     "test message",
				Entities: []*client.TextEntity{},
			},
			src: &client.Message{
				ChatId:       -10121,
				Id:           123,
				MediaAlbumId: 789,
			},
			dstChatId:     -10123,
			prevMessageId: 456,
			engineConfig: &domain.EngineConfig{
				Sources: map[int64]*domain.Source{
					-10121: {
						Prev: domain.PREV_LINK,
					},
				},
			},
			expectedText:     "test message\n\n[Prev](https://t.me/test/456)",
			expectedEntities: []*client.TextEntity{},
			setup: func(t *testing.T) *Service {
				telegramRepo := mocks.NewTelegramRepo(t)

				telegramRepo.EXPECT().GetMessageLink(&client.GetMessageLinkRequest{
					ChatId:    -10123,
					MessageId: 456,
					ForAlbum:  true,
				}).Return(&client.MessageLink{
					Link: "https://t.me/test/456",
				}, nil)

				telegramRepo.EXPECT().ParseTextEntities(&client.ParseTextEntitiesRequest{
					Text: "[Prev](https://t.me/test/456)",
					ParseMode: &client.TextParseModeMarkdown{
						Version: 2,
					},
				}).Return(&client.FormattedText{
					Text:     "[Prev](https://t.me/test/456)",
					Entities: []*client.TextEntity{},
				}, nil)

				return New(telegramRepo, nil, nil)
			},
		},
		{
			name: "source_not_found",
			formattedText: &client.FormattedText{
				Text:     "test message",
				Entities: []*client.TextEntity{},
			},
			src: &client.Message{
				ChatId:       -10121,
				Id:           123,
				MediaAlbumId: 0,
			},
			dstChatId:     -10123,
			prevMessageId: 456,
			engineConfig: &domain.EngineConfig{
				Sources: map[int64]*domain.Source{},
			},
			expectedText:     "test message",
			expectedEntities: []*client.TextEntity{},
			setup: func(t *testing.T) *Service {
				return New(nil, nil, nil)
			},
		},
		{
			name: "get_message_link_error",
			formattedText: &client.FormattedText{
				Text:     "test message",
				Entities: []*client.TextEntity{},
			},
			src: &client.Message{
				ChatId:       -10121,
				Id:           123,
				MediaAlbumId: 0,
			},
			dstChatId:     -10123,
			prevMessageId: 456,
			engineConfig: &domain.EngineConfig{
				Sources: map[int64]*domain.Source{
					-10121: {
						Prev: domain.PREV_LINK,
					},
				},
			},
			expectedText:     "test message",
			expectedEntities: []*client.TextEntity{},
			setup: func(t *testing.T) *Service {
				telegramRepo := mocks.NewTelegramRepo(t)

				telegramRepo.EXPECT().GetMessageLink(&client.GetMessageLinkRequest{
					ChatId:    -10123,
					MessageId: 456,
					ForAlbum:  false,
				}).Return(nil, assert.AnError)

				return New(telegramRepo, nil, nil)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			transformService := test.setup(t)
			transformService.addPrevMessageId(test.formattedText, test.src, test.dstChatId, test.prevMessageId, test.engineConfig)

			assert.Equal(t, test.expectedText, test.formattedText.Text)
			assert.Equal(t, test.expectedEntities, test.formattedText.Entities)
		})
	}
}
