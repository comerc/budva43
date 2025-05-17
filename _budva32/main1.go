package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/zelenin/go-tdlib/client"
)

// OK: перенесено - service/engine/service.go (handleUpdates)
func handleUpdate(update *client.Update) {
	if update.GetClass() == client.ClassUpdate {
		switch updateType := update.(type) {
		case *client.UpdateNewMessage:
			// НЕТ: перенесено частично - service/engine/service.go (handleUpdateNewMessage)
			updateNewMessage := updateType
			src := updateNewMessage.Message
			// OK: перенесено - service/engine/service.go (deleteSystemMessage)
			go func() {
				if _, ok := configData.DeleteSystemMessages[src.ChatId]; ok {
					needDelete := false
					switch src.Content.(type) {
					case *client.MessageChatChangeTitle:
						needDelete = true
					case *client.MessageChatChangePhoto:
						needDelete = true
					case *client.MessageChatDeletePhoto:
						needDelete = true
					case *client.MessageChatAddMembers:
						needDelete = true
					case *client.MessageChatDeleteMember:
						needDelete = true
					case *client.MessageChatJoinByLink:
						needDelete = true
					case *client.MessagePinMessage:
						needDelete = true
					}
					if needDelete {
						_, err := tdlibClient.DeleteMessages(&client.DeleteMessagesRequest{
							ChatId:     src.ChatId,
							MessageIds: []int64{src.Id},
							Revoke:     true,
						})
						if err != nil {
							log.Print(err)
						}
					}
				}
			}()
			if _, ok := uniqueFrom[src.ChatId]; !ok {
				continue
			}
			// TODO: так нельзя отключать, а почему?
			// if src.IsOutgoing {
			// 	log.Print("src.IsOutgoing > ", src.ChatId)
			// 	continue // !!
			// }
			if _, contentMode := getFormattedText(src.Content); contentMode == "" {
				continue
			}
			isExist := false
			checkFns := make(map[int64]func())
			otherFns := make(map[int64]func())
			forwardedTo := make(map[int64]bool)
			// var wg sync.WaitGroup
			// configData := getConfig()
			for forwardKey, forward := range configData.Forwards {
				// !!!! copy for go routine
				var (
					forwardKey = forwardKey
					forward    = forward
				)
				if src.ChatId == forward.From && (forward.SendCopy || src.CanBeForwarded) {
					isExist = true
					for _, dstChatId := range forward.To {
						_, isPresent := forwardedTo[dstChatId]
						if !isPresent {
							forwardedTo[dstChatId] = false
						}
					}
					if src.MediaAlbumId == 0 {
						// wg.Add(1)
						// log.Print("wg.Add > src.Id: ", src.Id)
						fn := func() {
							// defer func() {
							// 	wg.Done()
							// 	log.Print("wg.Done > src.Id: ", src.Id)
							// }()
							doUpdateNewMessage([]*client.Message{src}, forwardKey, forward, forwardedTo, checkFns, otherFns)
						}
						queue.PushBack(fn)
					} else {
						isFirstMessage := addMessageToMediaAlbum(forwardKey, src)
						if isFirstMessage {
							// wg.Add(1)
							// log.Print("wg.Add > src.Id: ", src.Id)
							fn := func() {
								handleMediaAlbum(forwardKey, src.MediaAlbumId,
									func(messages []*client.Message) {
										// defer func() {
										// 	wg.Done()
										// 	log.Print("wg.Done > src.Id: ", src.Id)
										// }()
										doUpdateNewMessage(messages, forwardKey, forward, forwardedTo, checkFns, otherFns)
									})
							}
							queue.PushBack(fn)
						}
					}
				}
			}
			if isExist {
				fn := func() {
					// wg.Wait()
					// log.Print("wg.Wait > src.Id: ", src.Id)
					for dstChatId, isForwarded := range forwardedTo {
						if isForwarded {
							incrementForwardedMessages(dstChatId)
						}
						incrementViewedMessages(dstChatId)
					}
					for check, fn := range checkFns {
						if fn == nil {
							log.Printf("check: %d is nil", check)
							continue
						}
						log.Printf("check: %d is fn()", check)
						fn()
					}
					for other, fn := range otherFns {
						if fn == nil {
							log.Printf("other: %d is nil", other)
							continue
						}
						log.Printf("other: %d is fn()", other)
						fn()
					}
				}
				queue.PushBack(fn)
			}
		case *client.UpdateMessageEdited:
			// НЕТ: перенесено частично - service/engine/service.go (handleUpdateMessageEdited)
			updateMessageEdited := updateType
			chatId := updateMessageEdited.ChatId
			if _, ok := uniqueFrom[chatId]; !ok {
				continue
			}
			messageId := updateMessageEdited.MessageId
			repeat := 0
			var fn func()
			fn = func() {
				var result []string
				fromChatMessageId := fmt.Sprintf("%d:%d", chatId, messageId)
				toChatMessageIds := getCopiedMessageIds(fromChatMessageId)
				log.Printf("UpdateMessageEdited > do > fromChatMessageId: %s toChatMessageIds: %v", fromChatMessageId, toChatMessageIds)
				defer func() {
					log.Printf("UpdateMessageEdited > ok > result: %v", result)
				}()
				if len(toChatMessageIds) == 0 {
					return
				}
				var newMessageIds = make(map[string]int64)
				isUpdateMessageSendSucceeded := true
				for _, toChatMessageId := range toChatMessageIds {
					a := strings.Split(toChatMessageId, ":")
					// forwardKey := a[0]
					dstChatId := int64(convertToInt(a[1]))
					tmpMessageId := int64(convertToInt(a[2]))
					newMessageId := getNewMessageId(dstChatId, tmpMessageId)
					if newMessageId == 0 {
						isUpdateMessageSendSucceeded = false
						break
					}
					newMessageIds[fmt.Sprintf("%d:%d", dstChatId, tmpMessageId)] = newMessageId
				}
				if !isUpdateMessageSendSucceeded {
					repeat++
					if repeat < 3 {
						log.Print("isUpdateMessageSendSucceeded > repeat: ", repeat)
						queue.PushBack(fn)
					} else {
						log.Print("isUpdateMessageSendSucceeded > repeat limit !!!")
					}
					return
				}
				src, err := tdlibClient.GetMessage(&client.GetMessageRequest{
					ChatId:    chatId,
					MessageId: messageId,
				})
				if err != nil {
					log.Print("GetMessage > ", err)
					return
				}
				// TODO: isAnswer
				_, hasReplyMarkupData := getReplyMarkupData(src)
				srcFormattedText, contentMode := getFormattedText(src.Content)
				log.Printf("srcChatId: %d srcId: %d hasText: %t MediaAlbumId: %d", src.ChatId, src.Id, srcFormattedText.Text != "", src.MediaAlbumId)
				checkFns := make(map[int64]func())
				for _, toChatMessageId := range toChatMessageIds {
					a := strings.Split(toChatMessageId, ":")
					forwardKey := a[0]
					dstChatId := int64(convertToInt(a[1]))
					tmpMessageId := int64(convertToInt(a[2]))
					formattedText := copyFormattedText(srcFormattedText)
					if forward, ok := configData.Forwards[forwardKey]; ok {
						if forward.CopyOnce {
							continue
						}
						if (forward.SendCopy || src.CanBeForwarded) && checkFilters(formattedText, forward) == FiltersCheck {
							_, ok := checkFns[forward.Check]
							if !ok {
								checkFns[forward.Check] = func() {
									const isSendCopy = false // обязательно надо форвардить, иначе невидно текущего сообщения
									forwardNewMessages(tdlibClient, []*client.Message{src}, src.ChatId, forward.Check, isSendCopy, forwardKey)
								}
							}
							continue
						}
					} else {
						continue
					}
					// hasFiltersCheck := false
					// testChatId := dstChatId
					// for _, forward := range configData.Forwards {
					// 	forward := forward // !!!! copy for go routine
					// 	if src.ChatId == forward.From && (forward.SendCopy || src.CanBeForwarded) {
					// 		for _, dstChatId := range forward.To {
					// 			if testChatId == dstChatId {
					// 				if checkFilters(formattedText, forward) == FiltersCheck {
					// 					hasFiltersCheck = true
					// 					_, ok := checkFns[forward.Check]
					// 					if !ok {
					// 						checkFns[forward.Check] = func() {
					// 							const isSendCopy = false // обязательно надо форвардить, иначе невидно текущего сообщения
					// 							forwardNewMessages(tdlibClient, []*client.Message{src}, src.ChatId, forward.Check, isSendCopy)
					// 						}
					// 					}
					// 				}
					// 			}
					// 		}
					// 	}
					// }
					// if hasFiltersCheck {
					// 	continue
					// }
					addAutoAnswer(formattedText, src)
					replaceMyselfLinks(formattedText, src.ChatId, dstChatId)
					replaceFragments(formattedText, dstChatId)
					// resetEntities(formattedText, dstChatId)
					addSources(formattedText, src, dstChatId)
					newMessageId := newMessageIds[fmt.Sprintf("%d:%d", dstChatId, tmpMessageId)]
					result = append(result, fmt.Sprintf("toChatMessageId: %s, newMessageId: %d", toChatMessageId, newMessageId))
					log.Print("contentMode: ", contentMode)
					switch contentMode {
					case ContentModeText:
						content := getInputMessageContent(src.Content, formattedText, contentMode)
						dst, err := tdlibClient.EditMessageText(&client.EditMessageTextRequest{
							ChatId:              dstChatId,
							MessageId:           newMessageId,
							InputMessageContent: content,
							// ReplyMarkup: src.ReplyMarkup, // это не надо, юзер-бот игнорит изменение
						})
						if err != nil {
							log.Print("EditMessageText > ", err)
						}
						log.Printf("EditMessageText > dst: %#v", dst)
					case ContentModeAnimation:
						fallthrough
					case ContentModeDocument:
						fallthrough
					case ContentModeAudio:
						fallthrough
					case ContentModeVideo:
						fallthrough
					case ContentModePhoto:
						content := getInputMessageContent(src.Content, formattedText, contentMode)
						dst, err := tdlibClient.EditMessageMedia(&client.EditMessageMediaRequest{
							ChatId:              dstChatId,
							MessageId:           newMessageId,
							InputMessageContent: content,
						})
						if err != nil {
							log.Print("EditMessageMedia > ", err)
						}
						log.Printf("EditMessageMedia > dst: %#v", dst)
					case ContentModeVoiceNote:
						dst, err := tdlibClient.EditMessageCaption(&client.EditMessageCaptionRequest{
							ChatId:    dstChatId,
							MessageId: newMessageId,
							Caption:   formattedText,
						})
						if err != nil {
							log.Print("EditMessageCaption > ", err)
						}
						log.Printf("EditMessageCaption > dst: %#v", dst)
					default:
						continue
					}
					// TODO: isAnswer
					if hasReplyMarkupData {
						setAnswerMessageId(dstChatId, tmpMessageId, fromChatMessageId)
					} else {
						deleteAnswerMessageId(dstChatId, tmpMessageId)
					}
				}
				for check, fn := range checkFns {
					if fn == nil {
						log.Printf("check: %d is nil", check)
						continue
					}
					log.Printf("check: %d is fn()", check)
					fn()
				}
			}
			queue.PushBack(fn)
		case *client.UpdateMessageSendSucceeded:
			// ДА: перенесено - service/engine/service.go (handleUpdateMessageSendSucceeded)
			updateMessageSendSucceeded := updateType
			message := updateMessageSendSucceeded.Message
			tmpMessageId := updateMessageSendSucceeded.OldMessageId
			fn := func() {
				log.Print("UpdateMessageSendSucceeded > go")
				setNewMessageId(message.ChatId, tmpMessageId, message.Id)
				setTmpMessageId(message.ChatId, message.Id, tmpMessageId)
				log.Print("UpdateMessageSendSucceeded > ok")
			}
			queue.PushBack(fn)
		case *client.UpdateDeleteMessages:
			// НЕТ: перенесено частично - service/engine/service.go (handleUpdateDeleteMessages)
			updateDeleteMessages := updateType
			if !updateDeleteMessages.IsPermanent {
				continue
			}
			chatId := updateDeleteMessages.ChatId
			if _, ok := uniqueFrom[chatId]; !ok {
				continue
			}
			// TODO: а если удаление произошло в Forward.To - тоже надо чистить БД
			messageIds := updateDeleteMessages.MessageIds
			repeat := 0
			var fn func()
			fn = func() {
				var result []string
				log.Printf("UpdateDeleteMessages > do > chatId: %d messageIds: %v", chatId, messageIds)
				defer func() {
					log.Printf("UpdateDeleteMessages > ok > result: %v", result)
				}()
				var copiedMessageIds = make(map[string][]string)
				var newMessageIds = make(map[string]int64)
				isUpdateMessageSendSucceeded := true
				for _, messageId := range messageIds {
					fromChatMessageId := fmt.Sprintf("%d:%d", chatId, messageId)
					toChatMessageIds := getCopiedMessageIds(fromChatMessageId)
					copiedMessageIds[fromChatMessageId] = toChatMessageIds
					for _, toChatMessageId := range toChatMessageIds {
						a := strings.Split(toChatMessageId, ":")
						// forwardKey := a[0]
						dstChatId := int64(convertToInt(a[1]))
						tmpMessageId := int64(convertToInt(a[2]))
						newMessageId := getNewMessageId(dstChatId, tmpMessageId)
						if newMessageId == 0 {
							isUpdateMessageSendSucceeded = false
							break
						}
						newMessageIds[fmt.Sprintf("%d:%d", dstChatId, tmpMessageId)] = newMessageId
					}
				}
				if !isUpdateMessageSendSucceeded {
					repeat++
					if repeat < 3 {
						log.Print("isUpdateMessageSendSucceeded > repeat: ", repeat)
						queue.PushBack(fn)
					} else {
						log.Print("isUpdateMessageSendSucceeded > repeat limit !!!")
					}
					return
				}
				for _, messageId := range messageIds {
					fromChatMessageId := fmt.Sprintf("%d:%d", chatId, messageId)
					toChatMessageIds := copiedMessageIds[fromChatMessageId]
					for _, toChatMessageId := range toChatMessageIds {
						a := strings.Split(toChatMessageId, ":")
						forwardKey := a[0]
						dstChatId := int64(convertToInt(a[1]))
						tmpMessageId := int64(convertToInt(a[2]))
						if forward, ok := configData.Forwards[forwardKey]; ok {
							if forward.Indelible {
								continue
							}
						} else {
							continue
						}
						deleteAnswerMessageId(dstChatId, tmpMessageId)
						newMessageId := newMessageIds[fmt.Sprintf("%d:%d", dstChatId, tmpMessageId)]
						result = append(result, fmt.Sprintf("%d:%d:%d", dstChatId, tmpMessageId, newMessageId))
						deleteTmpMessageId(dstChatId, newMessageId)
						deleteNewMessageId(dstChatId, tmpMessageId)
						_, err := tdlibClient.DeleteMessages(&client.DeleteMessagesRequest{
							ChatId:     dstChatId,
							MessageIds: []int64{newMessageId},
							Revoke:     true,
						})
						if err != nil {
							log.Print("DeleteMessages > ", err)
							continue
						}
					}
					if len(toChatMessageIds) > 0 {
						deleteCopiedMessageIds(fromChatMessageId)
					}
				}
			}
			queue.PushBack(fn)
		}
	}
}
