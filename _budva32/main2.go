package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dgraph-io/badger"
)

// НЕТ: не перенесено, предлагаю - service/report/service.go (GenerateDailyReports)
func runReports() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for t := range ticker.C {
		utc := t.UTC()
		// w := utc.Weekday()
		// if w == 0 || w == 1 {
		// 	continue
		// }
		h := utc.Hour()
		m := utc.Minute()
		if h == 0 && m == 0 {
			// configData := getConfig()
			for _, toChatId := range configData.Reports.For {
				date := utc.Add(-1 * time.Minute).Format("2006-01-02")
				var viewed, forwarded int64
				{
					key := []byte(fmt.Sprintf("%s:%d:%s", viewedMessagesPrefix, toChatId, date))
					val := getForDB(key)
					if len(val) == 0 {
						viewed = 0
					} else {
						viewed = int64(bytesToUint64(val))
					}
				}
				{
					key := []byte(fmt.Sprintf("%s:%d:%s", forwardedMessagesPrefix, toChatId, date))
					val := getForDB(key)
					if len(val) == 0 {
						forwarded = 0
					} else {
						forwarded = int64(bytesToUint64(val))
					}
				}
				formattedText, err := tdlibClient.ParseTextEntities(&client.ParseTextEntitiesRequest{
					Text: fmt.Sprintf(configData.Reports.Template, forwarded, viewed),
					ParseMode: &client.TextParseModeMarkdown{
						Version: 2,
					},
				})
				if err != nil {
					log.Print("ParseTextEntities > ", err)
				} else {
					if _, err := tdlibClient.SendMessage(&client.SendMessageRequest{
						ChatId: toChatId,
						InputMessageContent: &client.InputMessageText{
							Text:                  formattedText,
							DisableWebPagePreview: true,
							ClearDraft:            true,
						},
						Options: &client.MessageSendOptions{
							DisableNotification: true,
						},
					}); err != nil {
						log.Print("SendMessage > ", err)
					}
				}
			}
		}
	}
}

// OK: перенесено - util/primitive.go (ConvertToInt)
func convertToInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		log.Print("convertToInt > ", err)
		return 0
	}
	return int(i)
}

// ****

// type FromChatMessageId string // copiedMessageIdsPrefix:srcChatId:srcMessageId
// type ToChatMessageId string // forwardKey:dstChatId:tmpMessageId // !! tmp

// var copiedMessageIds = make(map[FromChatMessageId][]ToChatMessageId)

const copiedMessageIdsPrefix = "copiedMsgIds"

// OK: перенесено - service/storage/service.go (DeleteCopiedMessageIds)
func deleteCopiedMessageIds(fromChatMessageId string) {
	key := []byte(fmt.Sprintf("%s:%s", copiedMessageIdsPrefix, fromChatMessageId))
	err := badgerDB.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})
	if err != nil {
		log.Print(err)
	}
	log.Printf("deleteCopiedMessageIds > fromChatMessageId: %s", fromChatMessageId)
}

// OK: перенесено - service/storage/service.go (SetCopiedMessageId)
func setCopiedMessageId(fromChatMessageId string, toChatMessageId string) {
	key := []byte(fmt.Sprintf("%s:%s", copiedMessageIdsPrefix, fromChatMessageId))
	var (
		err error
		val []byte
	)
	err = badgerDB.Update(func(txn *badger.Txn) error {
		var item *badger.Item
		item, err = txn.Get(key)
		if err != nil && err != badger.ErrKeyNotFound {
			return err
		}
		if err != badger.ErrKeyNotFound {
			val, err = item.ValueCopy(nil)
			if err != nil {
				return err
			}
		}
		result := []string{}
		s := fmt.Sprintf("%s", val)
		if s != "" {
			// workaround https://stackoverflow.com/questions/28330908/how-to-string-split-an-empty-string-in-go
			result = strings.Split(s, ",")
		}
		val = []byte(strings.Join(distinct(append(result, toChatMessageId)), ","))
		// val = []byte(strings.Join(toChatMessageIds, ","))
		return txn.Set(key, val)
	})
	if err != nil {
		log.Print("setCopiedMessageId > ", err)
	}
	log.Printf("setCopiedMessageId > fromChatMessageId: %s toChatMessageId: %s val: %s", fromChatMessageId, toChatMessageId, val)
}

// OK: перенесено - service/storage/service.go (GetCopiedMessageIds)
func getCopiedMessageIds(fromChatMessageId string) []string {
	key := []byte(fmt.Sprintf("%s:%s", copiedMessageIdsPrefix, fromChatMessageId))
	var (
		err error
		val []byte
	)
	err = badgerDB.View(func(txn *badger.Txn) error {
		var item *badger.Item
		item, err = txn.Get(key)
		if err != nil {
			return err
		}
		if err != badger.ErrKeyNotFound {
			val, err = item.ValueCopy(nil)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		log.Print("getCopiedMessageIds > ", err)
	}
	toChatMessageIds := []string{}
	s := fmt.Sprintf("%s", val)
	if s != "" {
		// workaround https://stackoverflow.com/questions/28330908/how-to-string-split-an-empty-string-in-go
		toChatMessageIds = strings.Split(s, ",")
	}
	log.Printf("getCopiedMessageIds > fromChatMessageId: %s toChatMessageIds: %v", fromChatMessageId, toChatMessageIds)
	return toChatMessageIds
}

// var newMessageIds = make(map[string]int64)

const newMessageIdPrefix = "newMsgId"

// OK: перенесено - service/storage/service.go (SetNewMessageId)
func setNewMessageId(chatId, tmpMessageId, newMessageId int64) {
	key := []byte(fmt.Sprintf("%s:%d:%d", newMessageIdPrefix, chatId, tmpMessageId))
	val := []byte(fmt.Sprintf("%d", newMessageId))
	err := badgerDB.Update(func(txn *badger.Txn) error {
		err := txn.Set(key, val)
		return err
	})
	if err != nil {
		log.Print("setNewMessageId > ", err)
	}
	log.Printf("setNewMessageId > key: %d:%d val: %d", chatId, tmpMessageId, newMessageId)
	// newMessageIds[ChatMessageId(fmt.Sprintf("%d:%d", chatId, tmpMessageId))] = newMessageId
}

// OK: перенесено - service/storage/service.go (GetNewMessageId)
func getNewMessageId(chatId, tmpMessageId int64) int64 {
	key := []byte(fmt.Sprintf("%s:%d:%d", newMessageIdPrefix, chatId, tmpMessageId))
	var (
		err error
		val []byte
	)
	err = badgerDB.View(func(txn *badger.Txn) error {
		var item *badger.Item
		item, err = txn.Get(key)
		if err != nil {
			return err
		}
		val, err = item.ValueCopy(nil)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		log.Print("getNewMessageId > ", err)
		return 0
	}
	newMessageId := int64(convertToInt(fmt.Sprintf("%s", val)))
	log.Printf("getNewMessageId > key: %d:%d val: %d", chatId, tmpMessageId, newMessageId)
	return newMessageId
	// return newMessageIds[ChatMessageId(fmt.Sprintf("%d:%d", chatId, tmpMessageId))]
}

// OK: перенесено - service/storage/service.go (DeleteNewMessageId)
func deleteNewMessageId(chatId, tmpMessageId int64) {
	key := []byte(fmt.Sprintf("%s:%d:%d", newMessageIdPrefix, chatId, tmpMessageId))
	err := badgerDB.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})
	if err != nil {
		log.Print(err)
	}
	log.Printf("deleteNewMessageId > key: %d:%d", chatId, tmpMessageId)
}

const tmpMessageIdPrefix = "tmpMsgId"

// ДА: перенесено - service/storage/service.go (SetTmpMessageId)
func setTmpMessageId(chatId, newMessageId, tmpMessageId int64) {
	key := []byte(fmt.Sprintf("%s:%d:%d", tmpMessageIdPrefix, chatId, newMessageId))
	val := []byte(fmt.Sprintf("%d", tmpMessageId))
	err := badgerDB.Update(func(txn *badger.Txn) error {
		err := txn.Set(key, val)
		return err
	})
	if err != nil {
		log.Print("setTmpMessageId > ", err)
	}
	log.Printf("setTmpMessageId > key: %d:%d val: %d", chatId, newMessageId, tmpMessageId)
}

// ДА: перенесено - service/storage/service.go (GetTmpMessageId)
func getTmpMessageId(chatId, newMessageId int64) int64 {
	key := []byte(fmt.Sprintf("%s:%d:%d", tmpMessageIdPrefix, chatId, newMessageId))
	var (
		err error
		val []byte
	)
	err = badgerDB.View(func(txn *badger.Txn) error {
		var item *badger.Item
		item, err = txn.Get(key)
		if err != nil {
			return err
		}
		val, err = item.ValueCopy(nil)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		log.Print("getTmpMessageId > ", err)
		return 0
	}
	tmpMessageId := int64(convertToInt(fmt.Sprintf("%s", val)))
	log.Printf("getTmpMessageId > key: %d:%d val: %d", chatId, newMessageId, tmpMessageId)
	return tmpMessageId
}

// ДА: перенесено - service/storage/service.go (DeleteTmpMessageId)
func deleteTmpMessageId(chatId, newMessageId int64) {
	key := []byte(fmt.Sprintf("%s:%d:%d", tmpMessageIdPrefix, chatId, newMessageId))
	err := badgerDB.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})
	if err != nil {
		log.Print(err)
	}
	log.Printf("deleteTmpMessageId > key: %d:%d", chatId, newMessageId)
}

var (
	lastForwarded   = make(map[int64]time.Time)
	lastForwardedMu sync.Mutex
)

// НЕТ: не перенесено, предлагаю - service/rate_limiter/service.go (GetLastForwardedDiff)
func getLastForwardedDiff(chatId int64) time.Duration {
	lastForwardedMu.Lock()
	defer lastForwardedMu.Unlock()
	return time.Since(lastForwarded[chatId])
}

// НЕТ: не перенесено, предлагаю - service/rate_limiter/service.go (SetLastForwarded)
func setLastForwarded(chatId int64) {
	lastForwardedMu.Lock()
	defer lastForwardedMu.Unlock()
	lastForwarded[chatId] = time.Now()
}

// НЕТ: перенесено частично - service/engine/service.go (forwardMessages)
func forwardNewMessages(tdlibClient *client.Client, messages []*client.Message, srcChatId, dstChatId int64, isSendCopy bool, forwardKey string) {
	log.Printf("forwardNewMessages > srcChatId: %d dstChatId: %d", srcChatId, dstChatId)
	diff := getLastForwardedDiff(dstChatId)
	if diff < waitForForward {
		time.Sleep(waitForForward - diff)
	}
	setLastForwarded(dstChatId)
	var (
		result *client.Messages
		err    error
	)
	if isSendCopy {
		contents := make([]client.InputMessageContent, 0)
		for i, message := range messages {
			if message.ForwardInfo != nil {
				if origin, ok := message.ForwardInfo.Origin.(*client.MessageForwardOriginChannel); ok {
					if originMessage, err := tdlibClient.GetMessage(&client.GetMessageRequest{
						ChatId:    origin.ChatId,
						MessageId: origin.MessageId,
					}); err != nil {
						log.Print("originMessage ", err)
					} else {
						targetMessage := message
						targetFormattedText, _ := getFormattedText(targetMessage.Content)
						originFormattedText, _ := getFormattedText(originMessage.Content)
						// workaround for https://github.com/tdlib/td/issues/1572
						if targetFormattedText.Text == originFormattedText.Text {
							messages[i] = originMessage
						} else {
							log.Print("targetMessage != originMessage")
						}
					}
				}
			}
			src := messages[i] // !!!! for origin message
			srcFormattedText, contentMode := getFormattedText(src.Content)
			formattedText := copyFormattedText(srcFormattedText)
			addAutoAnswer(formattedText, src)
			replaceMyselfLinks(formattedText, src.ChatId, dstChatId)
			replaceFragments(formattedText, dstChatId)
			// resetEntities(formattedText, dstChatId)
			if i == 0 {
				addSources(formattedText, src, dstChatId)
			}
			content := getInputMessageContent(src.Content, formattedText, contentMode)
			if content != nil {
				contents = append(contents, content)
			}
		}
		var replyToMessageId int64 = 0
		src := messages[0]
		if src.ReplyToMessageId > 0 && src.ReplyInChatId == src.ChatId {
			fromChatMessageId := fmt.Sprintf("%d:%d", src.ReplyInChatId, src.ReplyToMessageId)
			toChatMessageIds := getCopiedMessageIds(fromChatMessageId)
			var tmpMessageId int64 = 0
			for _, toChatMessageId := range toChatMessageIds {
				a := strings.Split(toChatMessageId, ":")
				if int64(convertToInt(a[1])) == dstChatId {
					tmpMessageId = int64(convertToInt(a[2]))
					break
				}
			}
			if tmpMessageId != 0 {
				replyToMessageId = getNewMessageId(dstChatId, tmpMessageId)
			}
		}
		if len(contents) == 1 {
			var message *client.Message
			message, err = tdlibClient.SendMessage(&client.SendMessageRequest{
				ChatId:              dstChatId,
				InputMessageContent: contents[0],
				ReplyToMessageId:    replyToMessageId,
			})
			if err != nil {
				// nothing
			} else {
				result = &client.Messages{
					TotalCount: 1,
					Messages:   []*client.Message{message},
				}
			}
		} else {
			result, err = tdlibClient.SendMessageAlbum(&client.SendMessageAlbumRequest{
				ChatId:               dstChatId,
				InputMessageContents: contents,
				ReplyToMessageId:     replyToMessageId,
			})
		}
	} else {
		result, err = tdlibClient.ForwardMessages(&client.ForwardMessagesRequest{
			ChatId:     dstChatId,
			FromChatId: srcChatId,
			MessageIds: func() []int64 {
				var messageIds []int64
				for _, message := range messages {
					messageIds = append(messageIds, message.Id)
				}
				return messageIds
			}(),
			Options: &client.MessageSendOptions{
				DisableNotification: false,
				FromBackground:      false,
				SchedulingState: &client.MessageSchedulingStateSendAtDate{
					SendDate: int32(time.Now().Unix()),
				},
			},
			SendCopy:      false,
			RemoveCaption: false,
		})
	}
	if err != nil {
		log.Print("forwardNewMessages > ", err)
	} else if len(result.Messages) != int(result.TotalCount) || result.TotalCount == 0 {
		log.Print("forwardNewMessages > invalid TotalCount")
	} else if len(result.Messages) != len(messages) {
		log.Print("forwardNewMessages > invalid len(messages)")
	} else if isSendCopy {
		for i, dst := range result.Messages {
			if dst == nil {
				log.Printf("!!!! dst == nil !!!! result: %#v messages: %#v", result, messages)
				continue
			}
			tmpMessageId := dst.Id
			src := messages[i] // !!!! for origin message
			toChatMessageId := fmt.Sprintf("%s:%d:%d", forwardKey, dstChatId, tmpMessageId)
			fromChatMessageId := fmt.Sprintf("%d:%d", src.ChatId, src.Id)
			setCopiedMessageId(fromChatMessageId, toChatMessageId)
			// TODO: isAnswer
			if _, ok := getReplyMarkupData(src); ok {
				setAnswerMessageId(dstChatId, tmpMessageId, fromChatMessageId)
			}
		}
	}
}

// НЕТ: не перенесено, предлагаю - service/message/service.go (GetReplyMarkupData)
func getReplyMarkupData(message *client.Message) ([]byte, bool) {
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

// OK: не перенесено - используется client.TypeMessage*
type ContentMode string

const (
	ContentModeText      = "text"
	ContentModeAnimation = "animation"
	ContentModeAudio     = "audio"
	ContentModeDocument  = "document"
	ContentModePhoto     = "photo"
	ContentModeVideo     = "video"
	ContentModeVoiceNote = "voiceNote"
)

// OK: перенесено - service/message/service.go (GetContent)
func getFormattedText(messageContent client.MessageContent) (*client.FormattedText, ContentMode) {
	var (
		formattedText *client.FormattedText
		contentMode   ContentMode
	)
	if content, ok := messageContent.(*client.MessageText); ok {
		formattedText = content.Text
		contentMode = ContentModeText
	} else if content, ok := messageContent.(*client.MessagePhoto); ok {
		formattedText = content.Caption
		contentMode = ContentModePhoto
	} else if content, ok := messageContent.(*client.MessageAnimation); ok {
		formattedText = content.Caption
		contentMode = ContentModeAnimation
	} else if content, ok := messageContent.(*client.MessageAudio); ok {
		formattedText = content.Caption
		contentMode = ContentModeAudio
	} else if content, ok := messageContent.(*client.MessageDocument); ok {
		formattedText = content.Caption
		contentMode = ContentModeDocument
	} else if content, ok := messageContent.(*client.MessageVideo); ok {
		formattedText = content.Caption
		contentMode = ContentModeVideo
	} else if content, ok := messageContent.(*client.MessageVoiceNote); ok {
		formattedText = content.Caption
		contentMode = ContentModeVoiceNote
	} else {
		// TODO: надо поддерживать больше типов?
		// client.MessageExpiredPhoto
		// client.MessageSticker
		// client.MessageExpiredVideo
		// client.MessageVideoNote
		// client.MessageLocation
		// client.MessageVenue
		// client.MessageContact
		// client.MessageDice
		// client.MessageGame
		// client.MessagePoll
		// client.MessageInvoice
		formattedText = &client.FormattedText{Entities: make([]*client.TextEntity, 0)}
		contentMode = ""
	}
	return formattedText, contentMode
}

// OK: перенесено - slices.Contains()
func contains(a []string, s string) bool {
	for _, t := range a {
		if t == s {
			return true
		}
	}
	return false
}

// OK: перенесено - slices.Contains()
func containsInt64(a []int64, e int64) bool {
	for _, t := range a {
		if t == e {
			return true
		}
	}
	return false
}
