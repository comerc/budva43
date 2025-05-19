package message

import (
	"log/slog"

	"github.com/zelenin/go-tdlib/client"
)

// Service предоставляет методы для обработки и преобразования сообщений
type Service struct {
	log *slog.Logger
	//
}

// New создает новый экземпляр сервиса для работы с сообщениями
func New() *Service {
	return &Service{
		log: slog.With("module", "service.message"),
		//
	}
}

// GetFormattedText извлекает содержимое сообщения для поддерживаемых типов
func (s *Service) GetFormattedText(message *client.Message) *client.FormattedText {
	if message == nil || message.Content == nil {
		return nil
	}
	switch content := message.Content.(type) {
	case *client.MessageText:
		return content.Text
	case *client.MessagePhoto:
		return content.Caption
	case *client.MessageVideo:
		return content.Caption
	case *client.MessageDocument:
		return content.Caption
	case *client.MessageAudio:
		return content.Caption
	case *client.MessageAnimation:
		return content.Caption
	case *client.MessageVoiceNote:
		return content.Caption
	default:
		return nil
	}
}

// IsSystemMessage проверяет, является ли сообщение системным
func (s *Service) IsSystemMessage(message *client.Message) bool {
	switch message.Content.(type) {
	case
		*client.MessageChatChangeTitle,
		*client.MessageChatChangePhoto,
		*client.MessageChatDeletePhoto,
		*client.MessageChatAddMembers,
		*client.MessageChatDeleteMember,
		*client.MessageChatJoinByLink,
		*client.MessagePinMessage:
		return true
	default:
		return false
	}
}

// GetReplyMarkupData извлекает данные из replyMarkup
func (s *Service) GetReplyMarkupData(message *client.Message) ([]byte, bool) {
	if message.ReplyMarkup != nil {
		if a, ok := message.ReplyMarkup.(*client.ReplyMarkupInlineKeyboard); ok {
			row := a.Rows[0]
			btn := row[0]
			if callback, ok := btn.Type.(*client.InlineKeyboardButtonTypeCallback); ok {
				return callback.Data, true
			}
		}
	}
	return nil, false
}

// GetInputMessageContent преобразует содержимое сообщения во входной контент
func (s *Service) GetInputMessageContent(message *client.Message, formattedText *client.FormattedText) client.InputMessageContent {
	messageContent := message.Content
	switch message.Content.(type) {
	case *client.MessageText:
		messageText := messageContent.(*client.MessageText)
		return &client.InputMessageText{
			Text:               formattedText,
			LinkPreviewOptions: messageText.LinkPreviewOptions,
			ClearDraft:         true,
		}
	case *client.MessageAnimation:
		messageAnimation := messageContent.(*client.MessageAnimation)
		return &client.InputMessageAnimation{
			Animation: &client.InputFileRemote{
				Id: messageAnimation.Animation.Animation.Remote.Id,
			},
			// TODO: AddedStickerFileIds , // if applicable?
			Duration: messageAnimation.Animation.Duration,
			Width:    messageAnimation.Animation.Width,
			Height:   messageAnimation.Animation.Height,
			Caption:  formattedText,
		}
	case *client.MessageAudio:
		messageAudio := messageContent.(*client.MessageAudio)
		return &client.InputMessageAudio{
			Audio: &client.InputFileRemote{
				Id: messageAudio.Audio.Audio.Remote.Id,
			},
			AlbumCoverThumbnail: getInputThumbnail(messageAudio.Audio.AlbumCoverThumbnail),
			Title:               messageAudio.Audio.Title,
			Duration:            messageAudio.Audio.Duration,
			Performer:           messageAudio.Audio.Performer,
			Caption:             formattedText,
		}
	case *client.MessageDocument:
		messageDocument := messageContent.(*client.MessageDocument)
		return &client.InputMessageDocument{
			Document: &client.InputFileRemote{
				Id: messageDocument.Document.Document.Remote.Id,
			},
			Thumbnail: getInputThumbnail(messageDocument.Document.Thumbnail),
			Caption:   formattedText,
		}
	case *client.MessagePhoto:
		messagePhoto := messageContent.(*client.MessagePhoto)
		return &client.InputMessagePhoto{
			Photo: &client.InputFileRemote{
				Id: messagePhoto.Photo.Sizes[0].Photo.Remote.Id,
			},
			// Thumbnail: , // https://github.com/tdlib/td/issues/1505
			// A: if you use InputFileRemote, then there is no way to change the thumbnail, so there are no reasons to specify it.
			// TODO: AddedStickerFileIds: ,
			Width:   messagePhoto.Photo.Sizes[0].Width,
			Height:  messagePhoto.Photo.Sizes[0].Height,
			Caption: formattedText,
			// Ttl: ,
		}
	case *client.MessageVideo:
		messageVideo := messageContent.(*client.MessageVideo)
		// TODO: https://github.com/tdlib/td/issues/1504
		// var stickerSets *client.StickerSets
		// var AddedStickerFileIds []int32 // ????
		// if messageVideo.Video.HasStickers {
		// 	var err error
		// 	stickerSets, err = tdlibClient.GetAttachedStickerSets(&client.GetAttachedStickerSetsRequest{
		// 		FileId: messageVideo.Video.Video.Id,
		// 	})
		// 	if err != nil {
		// 		log.Print("GetAttachedStickerSets > ", err)
		// 	}
		// }
		return &client.InputMessageVideo{
			Video: &client.InputFileRemote{
				Id: messageVideo.Video.Video.Remote.Id,
			},
			Thumbnail: getInputThumbnail(messageVideo.Video.Thumbnail),
			// TODO: AddedStickerFileIds: ,
			Duration:          messageVideo.Video.Duration,
			Width:             messageVideo.Video.Width,
			Height:            messageVideo.Video.Height,
			SupportsStreaming: messageVideo.Video.SupportsStreaming,
			Caption:           formattedText,
			// Ttl: ,
		}
	case *client.MessageVoiceNote:
		return &client.InputMessageVoiceNote{
			// TODO: support ContentModeVoiceNote
			// VoiceNote: ,
			// Duration: ,
			// Waveform: ,
			Caption: formattedText,
		}
	}
	return nil
}

// getInputThumbnail преобразует thumbnail в входной контент
func getInputThumbnail(thumbnail *client.Thumbnail) *client.InputThumbnail {
	if thumbnail == nil || thumbnail.File == nil && thumbnail.File.Remote == nil {
		return nil
	}
	return &client.InputThumbnail{
		Thumbnail: &client.InputFileRemote{
			Id: thumbnail.File.Remote.Id,
		},
		Width:  thumbnail.Width,
		Height: thumbnail.Height,
	}
}
