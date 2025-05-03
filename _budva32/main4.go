package main

import (
	"container/list"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"
	"unicode/utf16"

	"github.com/zelenin/go-tdlib/client"
)

// TODO: потокобезопасное взаимодействие с queue?

var queue = list.New()

// OK: перенесено - service/queue/service.go (runQueue)
func runQueue() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for t := range ticker.C {
		_ = t
		// log.Print(t.UTC().Second())
		front := queue.Front()
		if front != nil {
			fn := front.Value.(func())
			fn()
			// This will remove the allocated memory and avoid memory leaks
			queue.Remove(front)
		}
	}
}

// OK: перенесено - util/primitive.go (Distinct)
func distinct(a []string) []string {
	set := make(map[string]struct{})
	for _, val := range a {
		set[val] = struct{}{}
	}
	result := make([]string, 0, len(set))
	for key := range set {
		result = append(result, key)
	}
	return result
}

const waitForForward = 3 * time.Second // чтобы бот успел отреагировать на сообщение

// OK: перенесено - service/engine/service.go (getInputThumbnail)
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

const answerMessageIdPrefix = "answerMsgId"

// OK: перенесено - service/storage/service.go (SetAnswerMessageId)
func setAnswerMessageId(dstChatId, tmpMessageId int64, fromChatMessageId string) {
	key := []byte(fmt.Sprintf("%s:%d:%d", answerMessageIdPrefix, dstChatId, tmpMessageId))
	val := []byte(fromChatMessageId)
	setForDB(key, val)
}

// OK: перенесено - service/storage/service.go (GetAnswerMessageId)
func getAnswerMessageId(dstChatId, tmpMessageId int64) string {
	key := []byte(fmt.Sprintf("%s:%d:%d", answerMessageIdPrefix, dstChatId, tmpMessageId))
	val := getForDB(key)
	return string(val)
}

// НЕТ: перенесено частично - service/storage/service.go (DeleteAnswerMessageId)
func deleteAnswerMessageId(dstChatId, tmpMessageId int64) {
	key := []byte(fmt.Sprintf("%s:%d:%d", answerMessageIdPrefix, dstChatId, tmpMessageId))
	deleteForDB(key)
}

// НЕТ: перенесено частично - service/transform/service.go (AddAutoAnswer)
func addAutoAnswer(formattedText *client.FormattedText, src *client.Message) {
	if configAnswer, ok := configData.Answers[src.ChatId]; ok && configAnswer.Auto {
		if data, ok := getReplyMarkupData(src); ok {
			if answer, err := tdlibClient.GetCallbackQueryAnswer(
				&client.GetCallbackQueryAnswerRequest{
					ChatId:    src.ChatId,
					MessageId: src.Id,
					Payload:   &client.CallbackQueryPayloadData{Data: data},
				},
			); err != nil {
				log.Print(err)
			} else {
				sourceAnswer, err := tdlibClient.ParseTextEntities(&client.ParseTextEntitiesRequest{
					Text: escapeAll(answer.Text),
					ParseMode: &client.TextParseModeMarkdown{
						Version: 2,
					},
				})
				if err != nil {
					log.Print("ParseTextEntities > ", err)
				} else {
					offset := int32(strLenForUTF16(formattedText.Text))
					if offset > 0 {
						formattedText.Text += "\n\n"
						offset = offset + 2
					}
					for _, entity := range sourceAnswer.Entities {
						entity.Offset += offset
					}
					formattedText.Text += sourceAnswer.Text
					formattedText.Entities = append(formattedText.Entities, sourceAnswer.Entities...)
				}
				log.Printf("addAutoAnswer > %#v", formattedText)
			}
		}
	}
}

// НЕТ: перенесено частично - service/transform/service.go (addSourceSign)
func addSourceSign(formattedText *client.FormattedText, title string) {
	sourceSign, err := tdlibClient.ParseTextEntities(&client.ParseTextEntitiesRequest{
		Text: title,
		ParseMode: &client.TextParseModeMarkdown{
			Version: 2,
		},
	})
	if err != nil {
		log.Print("ParseTextEntities > ", err)
	} else {
		offset := int32(strLenForUTF16(formattedText.Text))
		if offset > 0 {
			formattedText.Text += "\n\n"
			offset = offset + 2
		}
		for _, entity := range sourceSign.Entities {
			entity.Offset += offset
		}
		formattedText.Text += sourceSign.Text
		formattedText.Entities = append(formattedText.Entities, sourceSign.Entities...)
	}
	log.Printf("addSourceSign > %#v", formattedText)
}

// НЕТ: перенесено частично - service/transform/service.go (addSourceLink)
func addSourceLink(message *client.Message, formattedText *client.FormattedText, title string) {
	messageLink, err := tdlibClient.GetMessageLink(&client.GetMessageLinkRequest{
		ChatId:     message.ChatId,
		MessageId:  message.Id,
		ForAlbum:   message.MediaAlbumId != 0,
		ForComment: false,
	})
	if err != nil {
		log.Printf("GetMessageLink > ChatId: %d MessageId: %d %s", message.ChatId, message.Id, err)
	} else {
		sourceLink, err := tdlibClient.ParseTextEntities(&client.ParseTextEntitiesRequest{
			Text: fmt.Sprintf("[%s%s](%s)", "\U0001f517", title, messageLink.Link),
			ParseMode: &client.TextParseModeMarkdown{
				Version: 2,
			},
		})
		if err != nil {
			log.Print("ParseTextEntities > ", err)
		} else {
			// TODO: тут упало на опросе https://t.me/Full_Time_Trading/40922
			offset := int32(strLenForUTF16(formattedText.Text))
			if offset > 0 {
				formattedText.Text += "\n\n"
				offset = offset + 2
			}
			for _, entity := range sourceLink.Entities {
				entity.Offset += offset
			}
			formattedText.Text += sourceLink.Text
			formattedText.Entities = append(formattedText.Entities, sourceLink.Entities...)
		}
	}
	log.Printf("addSourceLink > %#v", formattedText)
}

// OK: перенесено - service/engine/service.go (getInputMessageContent)
func getInputMessageContent(messageContent client.MessageContent, formattedText *client.FormattedText, contentMode ContentMode) client.InputMessageContent {
	switch contentMode {
	case ContentModeText:
		messageText := messageContent.(*client.MessageText)
		return &client.InputMessageText{
			Text:                  formattedText,
			DisableWebPagePreview: messageText.WebPage == nil || messageText.WebPage.Url == "",
			ClearDraft:            true,
		}
	case ContentModeAnimation:
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
	case ContentModeAudio:
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
	case ContentModeDocument:
		messageDocument := messageContent.(*client.MessageDocument)
		return &client.InputMessageDocument{
			Document: &client.InputFileRemote{
				Id: messageDocument.Document.Document.Remote.Id,
			},
			Thumbnail: getInputThumbnail(messageDocument.Document.Thumbnail),
			Caption:   formattedText,
		}
	case ContentModePhoto:
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
	case ContentModeVideo:
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
	case ContentModeVoiceNote:
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

// НЕТ: перенесено частично - service/transform/service.go (ReplaceMyselfLinks)
func replaceMyselfLinks(formattedText *client.FormattedText, srcChatId, dstChatId int64) {
	if data, ok := configData.ReplaceMyselfLinks[dstChatId]; ok {
		log.Printf("replaceMyselfLinks > srcChatId: %d dstChatId: %d", srcChatId, dstChatId)
		for _, entity := range formattedText.Entities {
			if textUrl, ok := entity.Type.(*client.TextEntityTypeTextUrl); ok {
				if messageLinkInfo, err := tdlibClient.GetMessageLinkInfo(&client.GetMessageLinkInfoRequest{
					Url: textUrl.Url,
				}); err != nil {
					log.Print("GetMessageLinkInfo > ", err)
				} else {
					src := messageLinkInfo.Message
					if src != nil && srcChatId == src.ChatId {
						isReplaced := false
						fromChatMessageId := fmt.Sprintf("%d:%d", src.ChatId, src.Id)
						toChatMessageIds := getCopiedMessageIds(fromChatMessageId)
						log.Printf("fromChatMessageId: %s toChatMessageIds: %v", fromChatMessageId, toChatMessageIds)
						var tmpMessageId int64 = 0
						for _, toChatMessageId := range toChatMessageIds {
							a := strings.Split(toChatMessageId, ":")
							if int64(convertToInt(a[1])) == dstChatId {
								tmpMessageId = int64(convertToInt(a[2]))
								break
							}
						}
						if tmpMessageId != 0 {
							if messageLink, err := tdlibClient.GetMessageLink(&client.GetMessageLinkRequest{
								ChatId:    dstChatId,
								MessageId: getNewMessageId(dstChatId, tmpMessageId),
							}); err != nil {
								log.Print("GetMessageLink > ", err)
							} else {
								entity.Type = &client.TextEntityTypeTextUrl{
									Url: messageLink.Link,
								}
								isReplaced = true
							}
						}
						if !isReplaced && data.DeleteExternal {
							entity.Type = &client.TextEntityTypeStrikethrough{}
						}
					}
				}
			}
		}
	}
}

// OK: перенесено - util/primitive.go (Copy)
func copyFormattedText(formattedText *client.FormattedText) *client.FormattedText {
	result := *formattedText
	return &result
}

// OK: перенесено - service/transform/service.go (escapeMarkdown)
func escapeAll(s string) string {
	// эскейпит все символы: которые нужны для markdown-разметки
	a := []string{
		"_",
		"*",
		`\[`,
		`\]`,
		"(",
		")",
		"~",
		"`",
		">",
		"#",
		"+",
		`\-`,
		"=",
		"|",
		"{",
		"}",
		".",
		"!",
	}
	re := regexp.MustCompile("[" + strings.Join(a, "|") + "]")
	return re.ReplaceAllString(s, `\$0`)
}

// НЕТ: перенесено частично - service/transform/service.go (AddSources)
func addSources(formattedText *client.FormattedText, src *client.Message, dstChatId int64) {
	if source, ok := configData.Sources[src.ChatId]; ok {
		if containsInt64(source.Sign.For, dstChatId) {
			addSourceSign(formattedText, source.Sign.Title)
		} else if containsInt64(source.Link.For, dstChatId) {
			addSourceLink(src, formattedText, source.Link.Title)
		}
	}
}

// func resetEntities(formattedText *client.FormattedText, dstChatId int64) {
//	// withResetEntities := containsInt64(configData.ResetEntities, dstChatId)
// 	if result, err := tdlibClient.ParseTextEntities(&client.ParseTextEntitiesRequest{
// 		Text: escapeAll(formattedText.Text),
// 		ParseMode: &client.TextParseModeMarkdown{
// 			Version: 2,
// 		},
// 	}); err != nil {
// 		log.Print(err)
// 	} else {
// 		*formattedText = *result
// 	}
// }

// НЕТ: перенесено частично - service/transform/service.go (ReplaceFragments)
func replaceFragments(formattedText *client.FormattedText, dstChatId int64) {
	if data, ok := configData.ReplaceFragments[dstChatId]; ok {
		isReplaced := false
		for from, to := range data {
			re := regexp.MustCompile("(?i)" + from)
			if re.FindString(formattedText.Text) != "" {
				isReplaced = true
				// вынес в конфиг
				// if strLenForUTF16(from) != strLenForUTF16(to) {
				// 	log.Print("error: strLenForUTF16(from) != strLenForUTF16(to)")
				// 	to = strings.Repeat(".", strLenForUTF16(from))
				// }
				formattedText.Text = re.ReplaceAllString(formattedText.Text, to)
			}
		}
		if isReplaced {
			log.Print("isReplaced")
		}
	}
}

// func replaceFragments2(formattedText *client.FormattedText, dstChatId int64) {
// 	if replaceFragments, ok := configData.ReplaceFragments[dstChatId]; ok {
// 		// TODO: нужно реализовать свою версию GetMarkdownText,
// 		// которая будет обрабатывать вложенные markdown-entities и экранировать markdown-элементы
// 		// https://github.com/tdlib/td/issues/1564
// 		log.Print(formattedText.Text)
// 		if markdownText, err := tdlibClient.GetMarkdownText(&client.GetMarkdownTextRequest{Text: formattedText}); err != nil {
// 			log.Print(err)
// 		} else {
// 			log.Print(markdownText.Text)
// 			isReplaced := false
// 			for from, to := range replaceFragments {
// 				re := regexp.MustCompile("(?i)" + from)
// 				if re.FindString(markdownText.Text) != "" {
// 					isReplaced = true
// 					markdownText.Text = re.ReplaceAllString(markdownText.Text, to)
// 				}
// 			}
// 			if isReplaced {
// 				var err error
// 				result, err := tdlibClient.ParseMarkdown(
// 					&client.ParseMarkdownRequest{
// 						Text: markdownText,
// 					},
// 				)
// 				if err != nil {
// 					log.Print(err)
// 				}
// 				*formattedText = *result
// 			}
// 		}
// 	}
// }

// OK: перенесено - util/primitive.go (RuneCountForUTF16)
func strLenForUTF16(s string) int {
	return len(utf16.Encode([]rune(s)))
}
